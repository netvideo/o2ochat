package ui

import (
	"sync"
)

type DefaultCallUI struct {
	mu                    sync.RWMutex
	activeCalls           map[string]*CallUIState
	videoFrameCallback    func(frame []byte, width, height int)
	audioDataCallback     func(data []byte, sampleRate int)
	callControlCallback   func(action CallAction)
}

func NewCallUI() CallUI {
	return &DefaultCallUI{
		activeCalls: make(map[string]*CallUIState),
	}
}

func (c *DefaultCallUI) ShowIncomingCall(callInfo *CallInfo) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if callInfo == nil || callInfo.SessionID == "" {
		return ErrInvalidParameter
	}

	state := &CallUIState{
		SessionID:  callInfo.SessionID,
		PeerID:     callInfo.PeerID,
		PeerName:   callInfo.PeerName,
		IsIncoming: true,
		HasVideo:   callInfo.HasVideo,
	}

	c.activeCalls[callInfo.SessionID] = state
	return nil
}

func (c *DefaultCallUI) ShowOutgoingCall(callInfo *CallInfo) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if callInfo == nil || callInfo.SessionID == "" {
		return ErrInvalidParameter
	}

	state := &CallUIState{
		SessionID:  callInfo.SessionID,
		PeerID:     callInfo.PeerID,
		PeerName:   callInfo.PeerName,
		IsIncoming: false,
		HasVideo:   callInfo.HasVideo,
	}

	c.activeCalls[callInfo.SessionID] = state
	return nil
}

func (c *DefaultCallUI) UpdateCallState(state *CallUIState) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if state == nil || state.SessionID == "" {
		return ErrInvalidParameter
	}

	c.activeCalls[state.SessionID] = state
	return nil
}

func (c *DefaultCallUI) EndCall(sessionID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.activeCalls[sessionID]; !ok {
		return ErrCallNotFound
	}

	delete(c.activeCalls, sessionID)
	return nil
}

func (c *DefaultCallUI) SetVideoFrameCallback(callback func(frame []byte, width, height int)) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.videoFrameCallback = callback
	return nil
}

func (c *DefaultCallUI) SetAudioDataCallback(callback func(data []byte, sampleRate int)) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.audioDataCallback = callback
	return nil
}

func (c *DefaultCallUI) SetCallControlCallback(callback func(action CallAction)) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.callControlCallback = callback
	return nil
}
