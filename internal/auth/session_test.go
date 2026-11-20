package auth

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/gaborgoczan/nacho/internal/db"
)

// testStore connects to TEST_DATABASE_URL, clears the identity tables, and
// creates one waiter. Tests skip when the variable is unset.
func testStore(t *testing.T) (*SessionStore, *sql.DB, int64) {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })

	if _, err := sqlDB.Exec(`truncate sessions, users restart identity cascade`); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	userID, err := db.New(sqlDB).CreateUser(context.Background(), db.CreateUserParams{
		Username:     "tester",
		DisplayName:  "Test User",
		Role:         "waiter",
		PasswordHash: "x",
	})
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	return NewSessionStore(db.New(sqlDB), 30*time.Minute, 12*time.Hour), sqlDB, userID
}

func TestCreateLookupRevoke(t *testing.T) {
	store, sqlDB, userID := testStore(t)
	ctx := context.Background()

	token, expiresAt, err := store.Create(ctx, userID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if token == "" || !expiresAt.After(time.Now().Add(29*time.Minute)) {
		t.Fatalf("unexpected token/expiry: %q %v", token, expiresAt)
	}

	sess, err := store.Lookup(ctx, token)
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if sess.UserID != userID || sess.Role != "waiter" || sess.DisplayName != "Test User" {
		t.Fatalf("unexpected session: %+v", sess)
	}

	var (
		hashLen int
		hashHex string
	)
	err = sqlDB.QueryRow(
		`select length(token_hash), encode(token_hash, 'hex') from sessions`,
	).Scan(&hashLen, &hashHex)
	if err != nil {
		t.Fatalf("query stored hash: %v", err)
	}
	if hashLen != 32 {
		t.Fatalf("stored hash should be 32 bytes, got %d", hashLen)
	}
	if hashHex == token {
		t.Fatal("token must not be stored in plaintext")
	}

	if err := store.Revoke(ctx, token); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if _, err := store.Lookup(ctx, token); !errors.Is(err, ErrNoSession) {
		t.Fatalf("want ErrNoSession after revoke, got %v", err)
	}
}

func TestLookupRejectsExpired(t *testing.T) {
	store, sqlDB, userID := testStore(t)
	ctx := context.Background()

	token, _, err := store.Create(ctx, userID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := sqlDB.Exec(
		`update sessions set expires_at = now() - interval '1 second'`); err != nil {
		t.Fatalf("expire: %v", err)
	}

	if _, err := store.Lookup(ctx, token); !errors.Is(err, ErrNoSession) {
		t.Fatalf("want ErrNoSession, got %v", err)
	}
}

func TestLookupRejectsPastAbsoluteCap(t *testing.T) {
	store, sqlDB, userID := testStore(t)
	ctx := context.Background()

	token, _, err := store.Create(ctx, userID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	// Simulate a row that escaped cleanup: created 13h ago, expiry still future.
	if _, err := sqlDB.Exec(
		`update sessions set created_at = now() - interval '13 hours', expires_at = now() + interval '1 hour'`); err != nil {
		t.Fatalf("backdate: %v", err)
	}

	if _, err := store.Lookup(ctx, token); !errors.Is(err, ErrNoSession) {
		t.Fatalf("want ErrNoSession, got %v", err)
	}
}

func TestLookupRejectsDisabledUser(t *testing.T) {
	store, sqlDB, userID := testStore(t)
	ctx := context.Background()

	token, _, err := store.Create(ctx, userID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := sqlDB.Exec(
		`update users set disabled_at = now() where id = $1`, userID); err != nil {
		t.Fatalf("disable: %v", err)
	}

	if _, err := store.Lookup(ctx, token); !errors.Is(err, ErrNoSession) {
		t.Fatalf("want ErrNoSession, got %v", err)
	}
}

func TestLookupRefreshesIdleWindow(t *testing.T) {
	store, sqlDB, userID := testStore(t)
	ctx := context.Background()

	token, _, err := store.Create(ctx, userID)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := sqlDB.Exec(
		`update sessions set expires_at = now() + interval '5 minutes', last_seen_at = now() - interval '10 minutes'`); err != nil {
		t.Fatalf("age session: %v", err)
	}

	sess, err := store.Lookup(ctx, token)
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if time.Until(sess.ExpiresAt) < 20*time.Minute {
		t.Fatalf("expiry was not refreshed: %v", sess.ExpiresAt)
	}
}

func TestRevokeAllForUser(t *testing.T) {
	store, sqlDB, userID := testStore(t)
	ctx := context.Background()

	if _, _, err := store.Create(ctx, userID); err != nil {
		t.Fatalf("create first: %v", err)
	}
	if _, _, err := store.Create(ctx, userID); err != nil {
		t.Fatalf("create second: %v", err)
	}

	if err := store.RevokeAllForUser(ctx, userID); err != nil {
		t.Fatalf("revoke all: %v", err)
	}

	var remaining int
	if err := sqlDB.QueryRow(`select count(*) from sessions`).Scan(&remaining); err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("expected no sessions, found %d", remaining)
	}
}
