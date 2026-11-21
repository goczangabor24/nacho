package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/gaborgoczan/nacho/internal/db"
)

var ErrNoSession = errors.New("auth: no valid session")

type Session struct {
	UserID      int64
	DisplayName string
	Role        string
	ExpiresAt   time.Time
}

type SessionStore struct {
	q       *db.Queries
	idleTTL time.Duration
	maxTTL  time.Duration
}

func NewSessionStore(q *db.Queries, idleTTL, maxTTL time.Duration) *SessionStore {
	return &SessionStore{q: q, idleTTL: idleTTL, maxTTL: maxTTL}
}

func (s *SessionStore) Create(ctx context.Context, userID int64) (string, time.Time, error) {
	token, err := newToken()
	if err != nil {
		return "", time.Time{}, err
	}

	now := time.Now().UTC()
	expiresAt := s.expiryFrom(now, now)

	err = s.q.CreateSession(ctx, db.CreateSessionParams{
		TokenHash:  hashToken(token),
		UserID:     userID,
		ExpiresAt:  expiresAt,
		LastSeenAt: now,
	})
	if err != nil {
		return "", time.Time{}, fmt.Errorf("auth: create session: %w", err)
	}

	return token, expiresAt, nil
}

func (s *SessionStore) Lookup(ctx context.Context, token string) (Session, error) {
	now := time.Now().UTC()
	hashed := hashToken(token)

	row, err := s.q.GetSessionByToken(ctx, hashed)
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, ErrNoSession
	}
	if err != nil {
		return Session{}, fmt.Errorf("auth: look up session: %w", err)
	}

	if now.After(row.ExpiresAt) || now.After(row.CreatedAt.Add(s.maxTTL)) || row.DisabledAt.Valid {
		_ = s.q.DeleteSession(ctx, hashed)
		return Session{}, ErrNoSession
	}

	newExpiry := s.expiryFrom(row.CreatedAt, now)
	session := Session{
		UserID:      row.UserID,
		DisplayName: row.DisplayName,
		Role:        row.Role,
		ExpiresAt:   newExpiry,
	}

	if newExpiry.Sub(row.ExpiresAt) >= 1*time.Minute {
		err := s.q.TouchSession(ctx, db.TouchSessionParams{
			ExpiresAt:  newExpiry,
			LastSeenAt: now,
			TokenHash:  hashed,
		})
		if err != nil {
			return Session{}, fmt.Errorf("auth: refresh session: %w", err)
		}
	}

	return session, nil
}

func (s *SessionStore) Revoke(ctx context.Context, token string) error {
	if err := s.q.DeleteSession(ctx, hashToken(token)); err != nil {
		return fmt.Errorf("auth: revoke session: %w", err)
	}
	return nil
}

func (s *SessionStore) RevokeAllForUser(ctx context.Context, userID int64) error {
	if err := s.q.DeleteSessionsForUser(ctx, userID); err != nil {
		return fmt.Errorf("auth: revoke user sessions: %w", err)
	}
	return nil
}

func (s *SessionStore) DeleteExpired(ctx context.Context) (int64, error) {
	now := time.Now().UTC()

	deleted, err := s.q.DeleteExpiredSessions(ctx, db.DeleteExpiredSessionsParams{
		Now:    now,
		Cutoff: now.Add(-s.maxTTL),
	})
	if err != nil {
		return 0, fmt.Errorf("auth: delete expired sessions: %w", err)
	}

	return deleted, nil
}

func (s *SessionStore) expiryFrom(created, now time.Time) time.Time {
	expiry := now.Add(s.idleTTL)
	limit := created.Add(s.maxTTL)

	if expiry.After(limit) {
		expiry = limit
	}
	return expiry
}

func newToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("auth: generate session token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}
