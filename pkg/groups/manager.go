package groups

import (
	"errors"
	"sync"
)

// Manager manages proxy groups and tags.
type Manager struct {
	mu     sync.RWMutex
	groups map[string]*Group         // group name -> Group
	tags   map[string]map[string]bool // proxy ID -> tags
}

// Group represents a logical grouping of proxies.
type Group struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	Metadata    map[string]string `json:"metadata"`
	ProxyIDs    []string          `json:"proxy_ids"`
}

// NewManager creates a new groups manager.
func NewManager() *Manager {
	return &Manager{
		groups: make(map[string]*Group),
		tags:   make(map[string]map[string]bool),
	}
}

// CreateGroup creates a new proxy group.
func (m *Manager) CreateGroup(name, description string, tags []string) error {
	if name == "" {
		return errors.New("group name cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.groups[name]; exists {
		return errors.New("group already exists")
	}

	m.groups[name] = &Group{
		Name:        name,
		Description: description,
		Tags:        tags,
		Metadata:    make(map[string]string),
		ProxyIDs:    make([]string, 0),
	}

	return nil
}

// DeleteGroup deletes a proxy group.
func (m *Manager) DeleteGroup(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.groups[name]; !exists {
		return errors.New("group not found")
	}

	delete(m.groups, name)
	return nil
}

// GetGroup retrieves a proxy group.
func (m *Manager) GetGroup(name string) (*Group, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	group, exists := m.groups[name]
	if !exists {
		return nil, errors.New("group not found")
	}

	// Return a copy to avoid race conditions
	return &Group{
		Name:        group.Name,
		Description: group.Description,
		Tags:        append([]string{}, group.Tags...),
		Metadata:    copyMap(group.Metadata),
		ProxyIDs:    append([]string{}, group.ProxyIDs...),
	}, nil
}

// ListGroups returns all proxy groups.
func (m *Manager) ListGroups() []Group {
	m.mu.RLock()
	defer m.mu.RUnlock()

	groups := make([]Group, 0, len(m.groups))
	for _, group := range m.groups {
		groups = append(groups, Group{
			Name:        group.Name,
			Description: group.Description,
			Tags:        append([]string{}, group.Tags...),
			Metadata:    copyMap(group.Metadata),
			ProxyIDs:    append([]string{}, group.ProxyIDs...),
		})
	}

	return groups
}

// AddProxyToGroup adds a proxy to a group.
func (m *Manager) AddProxyToGroup(groupName, proxyID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	group, exists := m.groups[groupName]
	if !exists {
		return errors.New("group not found")
	}

	// Check if already in group
	for _, id := range group.ProxyIDs {
		if id == proxyID {
			return nil // Already in group
		}
	}

	group.ProxyIDs = append(group.ProxyIDs, proxyID)
	return nil
}

// RemoveProxyFromGroup removes a proxy from a group.
func (m *Manager) RemoveProxyFromGroup(groupName, proxyID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	group, exists := m.groups[groupName]
	if !exists {
		return errors.New("group not found")
	}

	for i, id := range group.ProxyIDs {
		if id == proxyID {
			group.ProxyIDs = append(group.ProxyIDs[:i], group.ProxyIDs[i+1:]...)
			return nil
		}
	}

	return errors.New("proxy not in group")
}

// AddTag adds a tag to a proxy.
func (m *Manager) AddTag(proxyID, tag string) error {
	if tag == "" {
		return errors.New("tag cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.tags[proxyID] == nil {
		m.tags[proxyID] = make(map[string]bool)
	}

	m.tags[proxyID][tag] = true
	return nil
}

// RemoveTag removes a tag from a proxy.
func (m *Manager) RemoveTag(proxyID, tag string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.tags[proxyID] == nil {
		return errors.New("proxy has no tags")
	}

	delete(m.tags[proxyID], tag)

	// Clean up empty map
	if len(m.tags[proxyID]) == 0 {
		delete(m.tags, proxyID)
	}

	return nil
}

// GetTags returns all tags for a proxy.
func (m *Manager) GetTags(proxyID string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tags := make([]string, 0)
	if m.tags[proxyID] != nil {
		for tag := range m.tags[proxyID] {
			tags = append(tags, tag)
		}
	}

	return tags
}

// FindProxiesByTag returns all proxies with the given tag.
func (m *Manager) FindProxiesByTag(tag string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	proxies := make([]string, 0)
	for proxyID, tags := range m.tags {
		if tags[tag] {
			proxies = append(proxies, proxyID)
		}
	}

	return proxies
}

// FindProxiesByGroup returns all proxies in the given group.
func (m *Manager) FindProxiesByGroup(groupName string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	group, exists := m.groups[groupName]
	if !exists {
		return nil, errors.New("group not found")
	}

	return append([]string{}, group.ProxyIDs...), nil
}

// HasTag checks if a proxy has a specific tag.
func (m *Manager) HasTag(proxyID, tag string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.tags[proxyID] == nil {
		return false
	}

	return m.tags[proxyID][tag]
}

// SetGroupMetadata sets metadata for a group.
func (m *Manager) SetGroupMetadata(groupName, key, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	group, exists := m.groups[groupName]
	if !exists {
		return errors.New("group not found")
	}

	group.Metadata[key] = value
	return nil
}

// GetGroupMetadata gets metadata for a group.
func (m *Manager) GetGroupMetadata(groupName, key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	group, exists := m.groups[groupName]
	if !exists {
		return "", errors.New("group not found")
	}

	value, ok := group.Metadata[key]
	if !ok {
		return "", errors.New("metadata key not found")
	}

	return value, nil
}

// copyMap creates a copy of a string map.
func copyMap(m map[string]string) map[string]string {
	copy := make(map[string]string, len(m))
	for k, v := range m {
		copy[k] = v
	}
	return copy
}

// GlobalGroupManager is the global group manager instance.
var GlobalGroupManager = NewManager()
