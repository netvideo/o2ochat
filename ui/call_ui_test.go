package ui

import (
	"testing"
	"time"
)

func TestNewCallUI(t *testing.T) {
	call := NewCallUI()
	if call == nil {
		t.Error("expected non-nil CallUI")
	}

	defaultCall, ok := call.(*DefaultCallUI)
	if !ok {
		t.Error("expected DefaultCallUI type")
	}

	if defaultCall.activeCalls == nil {
		t.Error("expected activeCalls map to be initialized")
	}
}

func TestCallUIShowIncomingCall(t *testing.T) {
	call := NewCallUI()

	info := &CallInfo{
		SessionID:  "call1",
		PeerID:     "QmPeer456",
		PeerName:   "张三",
		HasVideo:   true,
		IsIncoming: true,
	}

	err := call.ShowIncomingCall(info)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = call.ShowIncomingCall(nil)
	if err != ErrInvalidParameter {
		t.Errorf("expected ErrInvalidParameter, got %v", err)
	}

	err = call.ShowIncomingCall(&CallInfo{SessionID: ""})
	if err != ErrInvalidParameter {
		t.Errorf("expected ErrInvalidParameter, got %v", err)
	}
}

func TestCallUIShowOutgoingCall(t *testing.T) {
	call := NewCallUI()

	info := &CallInfo{
		SessionID:  "call1",
		PeerID:     "QmPeer456",
		PeerName:   "张三",
		HasVideo:   true,
		IsIncoming: false,
	}

	err := call.ShowOutgoingCall(info)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCallUIUpdateCallState(t *testing.T) {
	call := NewCallUI()

	info := &CallInfo{
		SessionID:  "call1",
		PeerID:     "QmPeer456",
		PeerName:   "张三",
		HasVideo:   true,
		IsIncoming: true,
	}
	call.ShowIncomingCall(info)

	state := &CallUIState{
		SessionID: "call1",
		IsMuted:   true,
		Duration:  60 * time.Second,
	}

	err := call.UpdateCallState(state)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = call.UpdateCallState(nil)
	if err != ErrInvalidParameter {
		t.Errorf("expected ErrInvalidParameter, got %v", err)
	}

	err = call.UpdateCallState(&CallUIState{SessionID: ""})
	if err != ErrInvalidParameter {
		t.Errorf("expected ErrInvalidParameter, got %v", err)
	}
}

func TestCallUIEndCall(t *testing.T) {
	call := NewCallUI()

	info := &CallInfo{
		SessionID: "call1",
		PeerID:    "QmPeer456",
		PeerName:  "张三",
	}
	call.ShowIncomingCall(info)

	err := call.EndCall("call1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = call.EndCall("nonexistent")
	if err != ErrCallNotFound {
		t.Errorf("expected ErrCallNotFound, got %v", err)
	}
}

func TestCallUISetVideoFrameCallback(t *testing.T) {
	call := NewCallUI()

	callbackCalled := false
	callback := func(frame []byte, width, height int) {
		callbackCalled = true
	}

	err := call.SetVideoFrameCallback(callback)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	defaultCall := call.(*DefaultCallUI)
	defaultCall.videoFrameCallback([]byte{0x00}, 1920, 1080)
	if !callbackCalled {
		t.Error("expected callback to be called")
	}
}

func TestCallUISetAudioDataCallback(t *testing.T) {
	call := NewCallUI()

	callbackCalled := false
	callback := func(data []byte, sampleRate int) {
		callbackCalled = true
	}

	err := call.SetAudioDataCallback(callback)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	defaultCall := call.(*DefaultCallUI)
	defaultCall.audioDataCallback([]byte{0x00}, 48000)
	if !callbackCalled {
		t.Error("expected callback to be called")
	}
}

func TestCallUISetCallControlCallback(t *testing.T) {
	call := NewCallUI()

	callbackCalled := false
	callback := func(action CallAction) {
		callbackCalled = true
	}

	err := call.SetCallControlCallback(callback)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	defaultCall := call.(*DefaultCallUI)
	defaultCall.callControlCallback(CallActionMute)
	if !callbackCalled {
		t.Error("expected callback to be called")
	}
}
