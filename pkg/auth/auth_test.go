package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHashPassword(t *testing.T) {
	password := "testpassword123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == "" {
		t.Error("Hash is empty")
	}
	if hash == password {
		t.Error("Hash should not equal password")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "testpassword123"
	hash, _ := HashPassword(password)

	if !CheckPasswordHash(password, hash) {
		t.Error("Valid password check failed")
	}

	if CheckPasswordHash("wrongpassword", hash) {
		t.Error("Invalid password check should fail")
	}
}

func TestCreateSession(t *testing.T) {
	auth := NewAuthenticator()
	token, err := auth.CreateSession()
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	if token == "" {
		t.Error("Token is empty")
	}

	// Validate the created session
	if !auth.ValidateSession(token) {
		t.Error("Created session should be valid")
	}
}

func TestValidateSession(t *testing.T) {
	auth := NewAuthenticator()

	// Test invalid token
	if auth.ValidateSession("invalid-token") {
		t.Error("Invalid token should not validate")
	}

	// Test valid token
	token, _ := auth.CreateSession()
	if !auth.ValidateSession(token) {
		t.Error("Valid token should validate")
	}

	// Test expired token
	auth.sessions[token] = Session{
		Token:     token,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	if auth.ValidateSession(token) {
		t.Error("Expired token should not validate")
	}
}

func TestMiddleware(t *testing.T) {
	auth := NewAuthenticator()
	token, _ := auth.CreateSession()

	handler := auth.Middleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test without cookie
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected %d, got %d", http.StatusUnauthorized, w.Code)
	}

	// Test with valid cookie
	req = httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	w = httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, w.Code)
	}

	// Test with invalid cookie
	req = httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "invalid"})
	w = httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected %d, got %d", http.StatusUnauthorized, w.Code)
	}
}
