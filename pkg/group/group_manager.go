package group

import (
	"sync"
	"time"
)

// Group represents a chat group
type Group struct {
	ID          string
	Name        string
	Description string
	OwnerID     string
	Members     map[string]*GroupMember
	MaxMembers  int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	mu          sync.RWMutex
}

// GroupMember represents a group member
type GroupMember struct {
	UserID     string
	Nickname   string
	Role       MemberRole
	JoinedAt   time.Time
	LastSeenAt time.Time
	mu         sync.RWMutex
}

// MemberRole represents member role in group
type MemberRole string

const (
	MemberRoleOwner   MemberRole = "owner"
	MemberRoleAdmin   MemberRole = "admin"
	MemberRoleMember  MemberRole = "member"
)

// GroupManager manages chat groups
type GroupManager struct {
	groups     map[string]*Group
	userGroups map[string][]string // userID -> groupIDs
	mu         sync.RWMutex
	stats      GroupStats
}

// GroupStats represents group statistics
type GroupStats struct {
	TotalGroups   int
	TotalMembers  int
	ActiveGroups  int
}

// NewGroupManager creates a new group manager
func NewGroupManager() *GroupManager {
	return &GroupManager{
		groups:     make(map[string]*Group),
		userGroups: make(map[string][]string),
	}
}

// CreateGroup creates a new group
func (gm *GroupManager) CreateGroup(ownerID, name, description string) (*Group, error) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	group := &Group{
		ID:          generateGroupID(),
		Name:        name,
		Description: description,
		OwnerID:     ownerID,
		Members:     make(map[string]*GroupMember),
		MaxMembers:  100,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Add owner as member
	ownerMember := &GroupMember{
		UserID:   ownerID,
		Nickname: "Owner",
		Role:     MemberRoleOwner,
		JoinedAt: time.Now(),
	}
	group.Members[ownerID] = ownerMember

	// Store group
	gm.groups[group.ID] = group
	gm.userGroups[ownerID] = append(gm.userGroups[ownerID], group.ID)

	gm.stats.TotalGroups++
	gm.stats.TotalMembers++

	return group, nil
}

// GetGroup gets a group by ID
func (gm *GroupManager) GetGroup(groupID string) (*Group, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return nil, ErrGroupNotFound
	}

	return group, nil
}

// AddMember adds a member to group
func (gm *GroupManager) AddMember(groupID, userID, nickname string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return ErrGroupNotFound
	}

	// Check if already member
	if _, exists := group.Members[userID]; exists {
		return ErrMemberAlreadyExists
	}

	// Check max members
	if len(group.Members) >= group.MaxMembers {
		return ErrGroupFull
	}

	// Add member
	member := &GroupMember{
		UserID:   userID,
		Nickname: nickname,
		Role:     MemberRoleMember,
		JoinedAt: time.Now(),
	}
	group.Members[userID] = member
	group.UpdatedAt = time.Now()

	// Update user groups
	gm.userGroups[userID] = append(gm.userGroups[userID], group.ID)

	gm.stats.TotalMembers++

	return nil
}

// RemoveMember removes a member from group
func (gm *GroupManager) RemoveMember(groupID, userID string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return ErrGroupNotFound
	}

	// Check if member exists
	if _, exists := group.Members[userID]; !exists {
		return ErrMemberNotFound
	}

	// Remove member
	delete(group.Members, userID)
	group.UpdatedAt = time.Now()

	// Update user groups
	userGroupIDs := gm.userGroups[userID]
	for i, gid := range userGroupIDs {
		if gid == groupID {
			gm.userGroups[userID] = append(userGroupIDs[:i], userGroupIDs[i+1:]...)
			break
		}
	}

	gm.stats.TotalMembers--

	return nil
}

// SetMemberRole sets member role
func (gm *GroupManager) SetMemberRole(groupID, userID string, role MemberRole) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return ErrGroupNotFound
	}

	member, exists := group.Members[userID]
	if !exists {
		return ErrMemberNotFound
	}

	member.mu.Lock()
	member.Role = role
	member.mu.Unlock()

	group.UpdatedAt = time.Now()

	return nil
}

// GetMember gets a member from group
func (gm *GroupManager) GetMember(groupID, userID string) (*GroupMember, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return nil, ErrGroupNotFound
	}

	member, exists := group.Members[userID]
	if !exists {
		return nil, ErrMemberNotFound
	}

	return member, nil
}

// GetMembers gets all members from group
func (gm *GroupManager) GetMembers(groupID string) ([]*GroupMember, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return nil, ErrGroupNotFound
	}

	members := make([]*GroupMember, 0, len(group.Members))
	for _, member := range group.Members {
		members = append(members, member)
	}

	return members, nil
}

// GetUserGroups gets all groups for a user
func (gm *GroupManager) GetUserGroups(userID string) ([]*Group, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	groupIDs, exists := gm.userGroups[userID]
	if !exists {
		return []*Group{}, nil
	}

	groups := make([]*Group, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		if group, exists := gm.groups[groupID]; exists {
			groups = append(groups, group)
		}
	}

	return groups, nil
}

// UpdateGroup updates group info
func (gm *GroupManager) UpdateGroup(groupID, name, description string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return ErrGroupNotFound
	}

	group.Name = name
	group.Description = description
	group.UpdatedAt = time.Now()

	return nil
}

// DeleteGroup deletes a group
func (gm *GroupManager) DeleteGroup(groupID string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return ErrGroupNotFound
	}

	// Remove from user groups
	for userID := range group.Members {
		userGroupIDs := gm.userGroups[userID]
		for i, gid := range userGroupIDs {
			if gid == groupID {
				gm.userGroups[userID] = append(userGroupIDs[:i], userGroupIDs[i+1:]...)
				break
			}
		}
	}

	// Delete group
	delete(gm.groups, groupID)
	gm.stats.TotalGroups--
	gm.stats.TotalMembers -= len(group.Members)

	return nil
}

// GetStats gets group statistics
func (gm *GroupManager) GetStats() GroupStats {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	return gm.stats
}

// UpdateMemberLastSeen updates member last seen time
func (gm *GroupManager) UpdateMemberLastSeen(groupID, userID string) error {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return ErrGroupNotFound
	}

	member, exists := group.Members[userID]
	if !exists {
		return ErrMemberNotFound
	}

	member.mu.Lock()
	member.LastSeenAt = time.Now()
	member.mu.Unlock()

	return nil
}

// generateGroupID generates a unique group ID
func generateGroupID() string {
	return "group-" + time.Now().Format("20060102150405")
}

// Group errors
var (
	ErrGroupNotFound      = "group not found"
	ErrMemberNotFound     = "member not found"
	ErrMemberAlreadyExists = "member already exists"
	ErrGroupFull          = "group is full"
)
