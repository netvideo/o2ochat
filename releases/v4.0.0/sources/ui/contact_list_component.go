package ui

import (
	"sort"
	"sync"
)

type ContactListComponent struct {
	mu               sync.RWMutex
	contacts         map[string]*ContactInfo
	filteredContacts []*ContactInfo
	searchQuery      string
	groupFilter      string
	showOnlineOnly   bool
	sortBy          string
	onContactSelect  func(peerID string)
	onContactAdd     func()
	onContactDelete  func(peerID string)
	onContactUpdate func(contact *ContactInfo)
}

func NewContactListComponent() *ContactListComponent {
	return &ContactListComponent{
		contacts: make(map[string]*ContactInfo),
		sortBy:   "name",
	}
}

func (cl *ContactListComponent) AddContact(contact *ContactInfo) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	cl.contacts[contact.PeerID] = contact
	cl.updateFilteredContacts()

	if cl.onContactUpdate != nil {
		cl.onContactUpdate(contact)
	}
}

func (cl *ContactListComponent) RemoveContact(peerID string) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	delete(cl.contacts, peerID)
	cl.updateFilteredContacts()
}

func (cl *ContactListComponent) UpdateContact(contact *ContactInfo) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	cl.contacts[contact.PeerID] = contact
	cl.updateFilteredContacts()

	if cl.onContactUpdate != nil {
		cl.onContactUpdate(contact)
	}
}

func (cl *ContactListComponent) GetContact(peerID string) (*ContactInfo, bool) {
	cl.mu.RLock()
	defer cl.mu.RUnlock()
	c, ok := cl.contacts[peerID]
	return c, ok
}

func (cl *ContactListComponent) GetAllContacts() []*ContactInfo {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	result := make([]*ContactInfo, 0, len(cl.contacts))
	for _, c := range cl.contacts {
		result = append(result, c)
	}
	return result
}

func (cl *ContactListComponent) GetFilteredContacts() []*ContactInfo {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	result := make([]*ContactInfo, len(cl.filteredContacts))
	copy(result, cl.filteredContacts)
	return result
}

func (cl *ContactListComponent) Search(query string) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	cl.searchQuery = query
	cl.updateFilteredContacts()
}

func (cl *ContactListComponent) SetGroupFilter(group string) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	cl.groupFilter = group
	cl.updateFilteredContacts()
}

func (cl *ContactListComponent) SetShowOnlineOnly(onlineOnly bool) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	cl.showOnlineOnly = onlineOnly
	cl.updateFilteredContacts()
}

func (cl *ContactListComponent) SetSortBy(field string) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	cl.sortBy = field
	cl.updateFilteredContacts()
}

func (cl *ContactListComponent) updateFilteredContacts() {
	cl.filteredContacts = make([]*ContactInfo, 0)

	for _, c := range cl.contacts {
		if cl.showOnlineOnly && !c.Online {
			continue
		}

		if cl.groupFilter != "" {
			hasGroup := false
			for _, g := range c.Groups {
				if g == cl.groupFilter {
					hasGroup = true
					break
				}
			}
			if !hasGroup {
				continue
			}
		}

		if cl.searchQuery != "" {
			if !containsIgnoreCase(c.Name, cl.searchQuery) && !containsIgnoreCase(c.PeerID, cl.searchQuery) {
				continue
			}
		}

		cl.filteredContacts = append(cl.filteredContacts, c)
	}

	sort.Slice(cl.filteredContacts, func(i, j int) bool {
		switch cl.sortBy {
		case "name":
			return cl.filteredContacts[i].Name < cl.filteredContacts[j].Name
		case "lastSeen":
			return cl.filteredContacts[i].LastSeen.After(cl.filteredContacts[j].LastSeen)
		case "unread":
			return cl.filteredContacts[i].UnreadCount > cl.filteredContacts[j].UnreadCount
		default:
			return cl.filteredContacts[i].Name < cl.filteredContacts[j].Name
		}
	})
}

func (cl *ContactListComponent) SetOnContactSelect(callback func(peerID string)) {
	cl.onContactSelect = callback
}

func (cl *ContactListComponent) SetOnContactAdd(callback func()) {
	cl.onContactAdd = callback
}

func (cl *ContactListComponent) SetOnContactDelete(callback func(peerID string)) {
	cl.onContactDelete = callback
}

func (cl *ContactListComponent) SetOnContactUpdate(callback func(contact *ContactInfo)) {
	cl.onContactUpdate = callback
}

func (cl *ContactListComponent) GetOnlineCount() int {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	count := 0
	for _, c := range cl.contacts {
		if c.Online {
			count++
		}
	}
	return count
}

func (cl *ContactListComponent) GetTotalCount() int {
	cl.mu.RLock()
	defer cl.mu.RUnlock()
	return len(cl.contacts)
}

func (cl *ContactListComponent) GetTotalUnreadCount() int {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	count := 0
	for _, c := range cl.contacts {
		count += c.UnreadCount
	}
	return count
}

func (cl *ContactListComponent) GetGroups() []string {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	groupSet := make(map[string]bool)
	for _, c := range cl.contacts {
		for _, g := range c.Groups {
			groupSet[g] = true
		}
	}

	groups := make([]string, 0, len(groupSet))
	for g := range groupSet {
		groups = append(groups, g)
	}
	sort.Strings(groups)
	return groups
}

func containsIgnoreCase(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}

	s = toLower(s)
	substr = toLower(substr)

	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}
