package ui

import (
	"errors"
	"sync"
)

type DefaultContactUI struct {
	mu                    sync.RWMutex
	contacts              map[string]*ContactInfo
	contactSelectCallback func(peerID string)
	addContactCallback    func(peerID, name string)
}

func NewContactUI() ContactUI {
	return &DefaultContactUI{
		contacts: make(map[string]*ContactInfo),
	}
}

func (c *DefaultContactUI) AddContact(contact *ContactInfo) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if contact == nil || contact.PeerID == "" {
		return ErrInvalidParameter
	}

	c.contacts[contact.PeerID] = contact
	return nil
}

func (c *DefaultContactUI) RemoveContact(peerID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.contacts[peerID]; !ok {
		return ErrContactNotFound
	}

	delete(c.contacts, peerID)
	return nil
}

func (c *DefaultContactUI) UpdateContact(contact *ContactInfo) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if contact == nil || contact.PeerID == "" {
		return ErrInvalidParameter
	}

	c.contacts[contact.PeerID] = contact
	return nil
}

func (c *DefaultContactUI) SearchContacts(query string) ([]*ContactInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var results []*ContactInfo
	for _, contact := range c.contacts {
		if contains(contact.Name, query) || contains(contact.PeerID, query) {
			results = append(results, contact)
		}
	}

	return results, nil
}

func (c *DefaultContactUI) GetAllContacts() ([]*ContactInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	contacts := make([]*ContactInfo, 0, len(c.contacts))
	for _, contact := range c.contacts {
		contacts = append(contacts, contact)
	}

	return contacts, nil
}

func (c *DefaultContactUI) GetOnlineContacts() ([]*ContactInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var onlineContacts []*ContactInfo
	for _, contact := range c.contacts {
		if contact.Online {
			onlineContacts = append(onlineContacts, contact)
		}
	}

	return onlineContacts, nil
}

func (c *DefaultContactUI) SetContactSelectCallback(callback func(peerID string)) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.contactSelectCallback = callback
	return nil
}

func (c *DefaultContactUI) SetAddContactCallback(callback func(peerID, name string)) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.addContactCallback = callback
	return nil
}

var ErrContactDuplicate = errors.New("ui: contact already exists")
