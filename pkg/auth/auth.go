package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode"

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
	// Validate password strength before hashing
	if err := ValidatePasswordStrength(password); err != nil {
		return "", err
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// ValidatePasswordStrength checks if a password meets security requirements
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	if len(password) > 128 {
		return errors.New("password must not exceed 128 characters")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// Require at least 3 of 4 character types
	typeCount := 0
	if hasUpper {
		typeCount++
	}
	if hasLower {
		typeCount++
	}
	if hasNumber {
		typeCount++
	}
	if hasSpecial {
		typeCount++
	}

	if typeCount < 3 {
		return errors.New("password must contain at least 3 of: uppercase, lowercase, numbers, special characters")
	}

	// Check for common weak passwords
	weakPasswords := []string{
		"password", "Password1!", "12345678", "qwerty",
		"admin", "letmein", "welcome", "monkey",
	}

	lowerPassword := strings.ToLower(password)
	for _, weak := range weakPasswords {
		if strings.Contains(lowerPassword, weak) {
			return errors.New("password is too common or weak")
		}
	}

	return nil
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
