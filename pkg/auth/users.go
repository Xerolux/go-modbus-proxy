package auth

import (
	"errors"
	"sync"
	"time"
)

// UserRole defines user permission levels.
type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleEditor   UserRole = "editor"
	RoleViewer   UserRole = "viewer"
	RoleReadOnly UserRole = "readonly"
)

// User represents a user in the system.
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // Never returned in JSON
	Role      UserRole  `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// UserStore manages users.
type UserStore struct {
	mu    sync.RWMutex
	users map[string]*User
}

// NewUserStore creates a new user store.
func NewUserStore() *UserStore {
	return &UserStore{
		users: make(map[string]*User),
	}
}

// AddUser adds a new user.
func (s *UserStore) AddUser(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[user.ID]; exists {
		return errors.New("user already exists")
	}

	s.users[user.ID] = user
	return nil
}

// RemoveUser removes a user.
func (s *UserStore) RemoveUser(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.users, id)
}

// GetUser retrieves a user by ID.
func (s *UserStore) GetUser(id string) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	return user, exists
}

// GetByUsername retrieves a user by username.
func (s *UserStore) GetByUsername(username string) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.Username == username {
			return user, true
		}
	}
	return nil, false
}

// GetAllUsers returns all users.
func (s *UserStore) GetAllUsers() []User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, *user)
	}
	return users
}

// UpdateUser updates a user.
func (s *UserStore) UpdateUser(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[user.ID]; !exists {
		return errors.New("user not found")
	}

	s.users[user.ID] = user
	return nil
}

// HasPermission checks if a user has a specific permission.
func (u *User) HasPermission(requiredRole UserRole) bool {
	if u.Role == RoleAdmin {
		return true
	}

	roleHierarchy := map[UserRole]int{
		RoleAdmin:    4,
		RoleEditor:   3,
		RoleViewer:   2,
		RoleReadOnly: 1,
	}

	return roleHierarchy[u.Role] >= roleHierarchy[requiredRole]
}
