package ui

import (
	"testing"
	"time"
)

func TestNewContactUI(t *testing.T) {
	contact := NewContactUI()
	if contact == nil {
		t.Error("expected non-nil ContactUI")
	}

	defaultContact, ok := contact.(*DefaultContactUI)
	if !ok {
		t.Error("expected DefaultContactUI type")
	}

	if defaultContact.contacts == nil {
		t.Error("expected contacts map to be initialized")
	}
}

func TestContactUIAddContact(t *testing.T) {
	contact := NewContactUI()

	c := &ContactInfo{
		PeerID:      "QmPeer123",
		Name:        "张三",
		Avatar:      "avatar.png",
		LastSeen:    time.Now(),
		Online:      true,
		UnreadCount: 3,
	}

	err := contact.AddContact(c)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = contact.AddContact(nil)
	if err != ErrInvalidParameter {
		t.Errorf("expected ErrInvalidParameter, got %v", err)
	}

	err = contact.AddContact(&ContactInfo{PeerID: ""})
	if err != ErrInvalidParameter {
		t.Errorf("expected ErrInvalidParameter, got %v", err)
	}
}

func TestContactUIRemoveContact(t *testing.T) {
	contact := NewContactUI()

	err := contact.RemoveContact("QmPeer123")
	if err != ErrContactNotFound {
		t.Errorf("expected ErrContactNotFound, got %v", err)
	}

	c := &ContactInfo{PeerID: "QmPeer123", Name: "张三"}
	contact.AddContact(c)

	err = contact.RemoveContact("QmPeer123")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestContactUIUpdateContact(t *testing.T) {
	contact := NewContactUI()

	c := &ContactInfo{PeerID: "QmPeer123", Name: "张三"}
	contact.AddContact(c)

	updated := &ContactInfo{PeerID: "QmPeer123", Name: "李四", Online: true}
	err := contact.UpdateContact(updated)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	contacts, _ := contact.GetAllContacts()
	if contacts[0].Name != "李四" {
		t.Errorf("expected name to be updated to 李四, got %s", contacts[0].Name)
	}

	err = contact.UpdateContact(nil)
	if err != ErrInvalidParameter {
		t.Errorf("expected ErrInvalidParameter, got %v", err)
	}
}

func TestContactUISearchContacts(t *testing.T) {
	contact := NewContactUI()

	contacts := []*ContactInfo{
		{PeerID: "QmPeer1", Name: "张三"},
		{PeerID: "QmPeer2", Name: "李四"},
		{PeerID: "QmPeer3", Name: "王五"},
	}
	for _, c := range contacts {
		contact.AddContact(c)
	}

	results, err := contact.SearchContacts("张三")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}

	results, err = contact.SearchContacts("Qm")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}

func TestContactUIGetAllContacts(t *testing.T) {
	contact := NewContactUI()

	contacts, err := contact.GetAllContacts()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(contacts) != 0 {
		t.Errorf("expected 0 contacts, got %d", len(contacts))
	}

	contact.AddContact(&ContactInfo{PeerID: "QmPeer1", Name: "张三"})
	contact.AddContact(&ContactInfo{PeerID: "QmPeer2", Name: "李四"})

	contacts, err = contact.GetAllContacts()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(contacts) != 2 {
		t.Errorf("expected 2 contacts, got %d", len(contacts))
	}
}

func TestContactUIGetOnlineContacts(t *testing.T) {
	contact := NewContactUI()

	contact.AddContact(&ContactInfo{PeerID: "QmPeer1", Name: "张三", Online: true})
	contact.AddContact(&ContactInfo{PeerID: "QmPeer2", Name: "李四", Online: false})
	contact.AddContact(&ContactInfo{PeerID: "QmPeer3", Name: "王五", Online: true})

	online, err := contact.GetOnlineContacts()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(online) != 2 {
		t.Errorf("expected 2 online contacts, got %d", len(online))
	}
}

func TestContactUISetContactSelectCallback(t *testing.T) {
	contact := NewContactUI()

	callbackCalled := false
	callback := func(peerID string) {
		callbackCalled = true
	}

	err := contact.SetContactSelectCallback(callback)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	defaultContact := contact.(*DefaultContactUI)
	defaultContact.contactSelectCallback("QmPeer123")
	if !callbackCalled {
		t.Error("expected callback to be called")
	}
}

func TestContactUISetAddContactCallback(t *testing.T) {
	contact := NewContactUI()

	callbackCalled := false
	callback := func(peerID, name string) {
		callbackCalled = true
	}

	err := contact.SetAddContactCallback(callback)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	defaultContact := contact.(*DefaultContactUI)
	defaultContact.addContactCallback("QmPeer123", "张三")
	if !callbackCalled {
		t.Error("expected callback to be called")
	}
}
