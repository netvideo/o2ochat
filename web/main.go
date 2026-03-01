package main

import (
	"fmt"
	"syscall/js"
)

// WebRTC API for browser
type WebRTCConnection struct {
	pc          *js.Value
	dataChannel *js.Value
	onMessage   func([]byte)
	onOpen      func()
	onClose     func()
}

// NewWebRTCConnection creates a new WebRTC connection
func NewWebRTCConnection(config map[string]interface{}) (*WebRTCConnection, error) {
	// Get RTCPeerConnection constructor
	window := js.Global()
	rtcp := window.Get("RTCPeerConnection")
	if rtcp.IsUndefined() {
		return nil, fmt.Errorf("WebRTC not supported")
	}

	// Create peer connection
	configObj := js.ValueOf(config)
	pc := rtcp.New(configObj)

	conn := &WebRTCConnection{
		pc: &pc,
	}

	// Setup event handlers
	conn.setupEventHandlers()

	return conn, nil
}

// setupEventHandlers sets up WebRTC event handlers
func (c *WebRTCConnection) setupEventHandlers() {
	// ondatachannel
	onDataChannel := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		dc := args[0]
		c.dataChannel = &dc
		c.setupDataChannel()
		return nil
	})
	(*c.pc).Set("ondatachannel", onDataChannel)
}

// setupDataChannel sets up data channel handlers
func (c *WebRTCConnection) setupDataChannel() {
	if c.dataChannel == nil {
		return
	}

	// onopen
	onOpen := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if c.onOpen != nil {
			c.onOpen()
		}
		return nil
	})
	(*c.dataChannel).Set("onopen", onOpen)

	// onmessage
	onMessage := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		data := args[0].Get("data")
		if c.onMessage != nil {
			c.onMessage([]byte(data.String()))
		}
		return nil
	})
	(*c.dataChannel).Set("onmessage", onMessage)

	// onclose
	onClose := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if c.onClose != nil {
			c.onClose()
		}
		return nil
	})
	(*c.dataChannel).Set("onclose", onClose)
}

// CreateDataChannel creates a data channel
func (c *WebRTCConnection) CreateDataChannel(label string) error {
	dc := (*c.pc).Call("createDataChannel", label)
	c.dataChannel = &dc
	c.setupDataChannel()
	return nil
}

// Send sends data through data channel
func (c *WebRTCConnection) Send(data []byte) error {
	if c.dataChannel == nil {
		return fmt.Errorf("data channel not ready")
	}
	(*c.dataChannel).Call("send", string(data))
	return nil
}

// Close closes the connection
func (c *WebRTCConnection) Close() {
	(*c.pc).Call("close")
}

// CreateOffer creates an SDP offer
func (c *WebRTCConnection) CreateOffer() (string, error) {
	// Implementation for creating SDP offer
	promise := (*c.pc).Call("createOffer")
	// Handle promise asynchronously
	return "", nil
}

// SetLocalDescription sets local description
func (c *WebRTCConnection) SetLocalDescription(sdp string) error {
	desc := js.ValueOf(map[string]interface{}{
		"type": "offer",
		"sdp":  sdp,
	})
	(*c.pc).Call("setLocalDescription", desc)
	return nil
}

// SetRemoteDescription sets remote description
func (c *WebRTCConnection) SetRemoteDescription(sdp string, sdpType string) error {
	desc := js.ValueOf(map[string]interface{}{
		"type": sdpType,
		"sdp":  sdp,
	})
	(*c.pc).Call("setRemoteDescription", desc)
	return nil
}

// AddICECandidate adds an ICE candidate
func (c *WebRTCConnection) AddICECandidate(candidate string) error {
	candidateObj := js.ValueOf(map[string]interface{}{
		"candidate":  candidate,
		"sdpMid":     "0",
		"sdpMLineIndex": 0,
	})
	(*c.pc).Call("addIceCandidate", candidateObj)
	return nil
}

// Main entry point for WebAssembly
func main() {
	fmt.Println("O2OChat Web Client initializing...")

	// Expose Go functions to JavaScript
	js.Global().Set("O2OChat", map[string]interface{}{
		"init": func(config map[string]interface{}) interface{} {
			conn, err := NewWebRTCConnection(config)
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"error":   err.Error(),
				}
			}
			return map[string]interface{}{
				"success": true,
				"conn":    conn,
			}
		},
		"connect": func(peerID string) interface{} {
			// Implementation for connecting to peer
			return map[string]interface{}{
				"success": true,
			}
		},
		"send": func(data string) interface{} {
			// Implementation for sending message
			return map[string]interface{}{
				"success": true,
			}
		},
	})

	// Keep the Go program running
	select {}
}
