package auth

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Session holds session data.
type Session struct {
	Token     string
	ExpiresAt time.Time
}

// Authenticator handles auth.
type Authenticator struct {
	mu       sync.RWMutex
	sessions map[string]Session
}

// NewAuthenticator creates a new authenticator.
func NewAuthenticator() *Authenticator {
	return &Authenticator{
		sessions: make(map[string]Session),
	}
}

// HashPassword hashes a password.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash checks password.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// CreateSession creates a session.
func (a *Authenticator) CreateSession() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := base64.URLEncoding.EncodeToString(b)

	a.mu.Lock()
	defer a.mu.Unlock()
	a.sessions[token] = Session{
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	return token, nil
}

// ValidateSession validates a session.
func (a *Authenticator) ValidateSession(token string) bool {
	a.mu.RLock()
	session, ok := a.sessions[token]
	a.mu.RUnlock()

	if !ok {
		return false
	}
	if time.Now().After(session.ExpiresAt) {
		a.mu.Lock()
		delete(a.sessions, token)
		a.mu.Unlock()
		return false
	}
	return true
}

// Middleware protects routes.
func (a *Authenticator) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("session_token")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if !a.ValidateSession(c.Value) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

// CleanupExpiredSessions periodically removes expired sessions.
func (a *Authenticator) CleanupExpiredSessions() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		a.mu.Lock()
		now := time.Now()
		for token, session := range a.sessions {
			if now.After(session.ExpiresAt) {
				delete(a.sessions, token)
			}
		}
		a.mu.Unlock()
	}
}
