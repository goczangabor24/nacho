package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gaborgoczan/nacho/internal/auth"
)

const sessionCookie = "session"

type SessionLookup interface {
	Lookup(ctx context.Context, token string) (auth.Session, error)
}

type ctxKey int

const sessionKey ctxKey = 0

func withSession(ctx context.Context, s auth.Session) context.Context {
	return context.WithValue(ctx, sessionKey, s)
}

func userFromContext(ctx context.Context) (auth.Session, bool) {
	s, ok := ctx.Value(sessionKey).(auth.Session)
	return s, ok
}

func requireAuth(sessions SessionLookup, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookie)
		if err != nil {
			denyUnauthenticated(w, r)
			return
		}

		sess, err := sessions.Lookup(r.Context(), cookie.Value)
		if errors.Is(err, auth.ErrNoSession) {
			denyUnauthenticated(w, r)
			return
		}
		if err != nil {
			// A database outage is not a failed login; do not report it as 401.
			log.Printf("requireAuth: lookup session: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(w, r.WithContext(withSession(r.Context(), sess)))
	})
}

func requireRole(role string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, ok := userFromContext(r.Context())
		if !ok {
			log.Print("requireRole: no session in context; route is missing requireAuth")
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if sess.Role != role {
			denyForbidden(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func denyUnauthenticated(w http.ResponseWriter, r *http.Request) {
	if isAPI(r) {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	target := r.URL.RequestURI()
	if target == "" {
		target = "/"
	}
	http.Redirect(w, r, "/login?next="+url.QueryEscape(target), http.StatusSeeOther)
}

func denyForbidden(w http.ResponseWriter, r *http.Request) {
	if isAPI(r) {
		writeJSONError(w, http.StatusForbidden, "forbidden")
		return
	}
	http.Error(w, "forbidden", http.StatusForbidden)
}

func isAPI(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, "/api/")
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"error":"` + message + `"}`))
}

func safeNext(raw string) string {
	if raw == "" ||
		!strings.HasPrefix(raw, "/") ||
		strings.HasPrefix(raw, "//") ||
		strings.ContainsRune(raw, '\\') {
		return "/"
	}
	for _, c := range raw {
		if c < 0x20 || c == 0x7f {
			return "/"
		}
	}
	return raw
}
