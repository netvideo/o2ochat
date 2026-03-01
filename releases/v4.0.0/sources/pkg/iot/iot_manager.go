package iot

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DeviceType represents IoT device type
type DeviceType string

const (
	DeviceTypeLight       DeviceType = "light"
	DeviceTypeThermostat  DeviceType = "thermostat"
	DeviceTypeLock        DeviceType = "lock"
	DeviceTypeCamera      DeviceType = "camera"
	DeviceTypeSpeaker     DeviceType = "speaker"
	DeviceTypeSensor      DeviceType = "sensor"
	DeviceTypeAppliance   DeviceType = "appliance"
)

// DeviceState represents device state
type DeviceState string

const (
	DeviceStateOnline   DeviceState = "online"
	DeviceStateOffline  DeviceState = "offline"
	DeviceStateError    DeviceState = "error"
)

// IoTDevice represents an IoT device
type IoTDevice struct {
	ID           string
	Name         string
	Type         DeviceType
	State        DeviceState
	IPAddress    string
	Port         int
	IsConnected  bool
	LastSeen     time.Time
	Capabilities []string
	Metadata     map[string]interface{}
	mu           sync.RWMutex
}

// LightDevice represents a smart light
type LightDevice struct {
	IoTDevice
	Brightness int    // 0-100
	Color      string // hex color
	IsOn       bool
	ColorTemp  int    // Kelvin
}

// ThermostatDevice represents a smart thermostat
type ThermostatDevice struct {
	IoTDevice
	CurrentTemp   float64
	TargetTemp    float64
	Mode          string // "heat", "cool", "auto", "off"
	FanMode       string // "auto", "on", "circulate"
	Humidity      int
}

// LockDevice represents a smart lock
type LockDevice struct {
	IoTDevice
	IsLocked    bool
	LockType    string // "deadbolt", "knob"
	BatteryLevel int
	AutoLockTime int // seconds
}

// CameraDevice represents a smart camera
type CameraDevice struct {
	IoTDevice
	IsRecording   bool
	MotionDetected bool
	NightVision   bool
	Resolution    string
	StorageUsed   int // MB
}

// IoTManager manages IoT devices
type IoTManager struct {
	devices     map[string]*IoTDevice
	scenes      map[string]*Scene
	automations []*Automation
	mu          sync.RWMutex
}

// Scene represents a device scene
type Scene struct {
	ID        string
	Name      string
	Devices   map[string]interface{} // deviceID -> state
	IsActive  bool
	CreatedAt time.Time
}

// Automation represents a device automation
type Automation struct {
	ID          string
	Name        string
	Trigger     Trigger
	Conditions  []Condition
	Actions     []Action
	IsActive    bool
	LastTriggered time.Time
}

// Trigger represents an automation trigger
type Trigger struct {
	Type       string // "time", "device", "location", "voice"
	DeviceID   string
	Time       time.Time
	Location   string
	VoiceCommand string
}

// Condition represents an automation condition
type Condition struct {
	Type       string // "device_state", "time_range", "location"
	DeviceID   string
	State      interface{}
	StartTime  time.Time
	EndTime    time.Time
}

// Action represents an automation action
type Action struct {
	Type       string // "device_control", "notification", "scene"
	DeviceID   string
	SceneID    string
	Command    string
	Parameters map[string]interface{}
}

// NewIoTManager creates a new IoT manager
func NewIoTManager() *IoTManager {
	return &IoTManager{
		devices:     make(map[string]*IoTDevice),
		scenes:      make(map[string]*Scene),
		automations: make([]*Automation, 0),
	}
}

// RegisterDevice registers an IoT device
func (im *IoTManager) RegisterDevice(device *IoTDevice) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	if _, exists := im.devices[device.ID]; exists {
		return fmt.Errorf("device already registered")
	}

	im.devices[device.ID] = device
	return nil
}

// UnregisterDevice unregisters an IoT device
func (im *IoTManager) UnregisterDevice(deviceID string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	if _, exists := im.devices[deviceID]; !exists {
		return fmt.Errorf("device not found")
	}

	delete(im.devices, deviceID)
	return nil
}

// GetDevice gets a device by ID
func (im *IoTManager) GetDevice(deviceID string) (*IoTDevice, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	device, exists := im.devices[deviceID]
	if !exists {
		return nil, fmt.Errorf("device not found")
	}

	return device, nil
}

// GetAllDevices gets all devices
func (im *IoTManager) GetAllDevices() []*IoTDevice {
	im.mu.RLock()
	defer im.mu.RUnlock()

	devices := make([]*IoTDevice, 0, len(im.devices))
	for _, device := range im.devices {
		devices = append(devices, device)
	}
	return devices
}

// GetDevicesByType gets devices by type
func (im *IoTManager) GetDevicesByType(deviceType DeviceType) []*IoTDevice {
	im.mu.RLock()
	defer im.mu.RUnlock()

	devices := make([]*IoTDevice, 0)
	for _, device := range im.devices {
		if device.Type == deviceType {
			devices = append(devices, device)
		}
	}
	return devices
}

// ControlDevice controls a device
func (im *IoTManager) ControlDevice(ctx context.Context, deviceID, command string, params map[string]interface{}) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	device, exists := im.devices[deviceID]
	if !exists {
		return fmt.Errorf("device not found")
	}

	if !device.IsConnected {
		return fmt.Errorf("device is offline")
	}

	// Execute command based on device type
	switch device.Type {
	case DeviceTypeLight:
		return im.controlLight(device, command, params)
	case DeviceTypeThermostat:
		return im.controlThermostat(device, command, params)
	case DeviceTypeLock:
		return im.controlLock(device, command, params)
	case DeviceTypeCamera:
		return im.controlCamera(device, command, params)
	default:
		return fmt.Errorf("unsupported device type")
	}
}

// controlLight controls a light device
func (im *IoTManager) controlLight(device *IoTDevice, command string, params map[string]interface{}) error {
	light := device // In production, would cast to LightDevice

	switch command {
	case "on":
		light.Metadata["is_on"] = true
	case "off":
		light.Metadata["is_on"] = false
	case "brightness":
		if brightness, ok := params["brightness"].(int); ok {
			light.Metadata["brightness"] = brightness
		}
	case "color":
		if color, ok := params["color"].(string); ok {
			light.Metadata["color"] = color
		}
	}

	device.LastSeen = time.Now()
	return nil
}

// controlThermostat controls a thermostat device
func (im *IoTManager) controlThermostat(device *IoTDevice, command string, params map[string]interface{}) error {
	switch command {
	case "set_temp":
		if temp, ok := params["temperature"].(float64); ok {
			device.Metadata["target_temp"] = temp
		}
	case "set_mode":
		if mode, ok := params["mode"].(string); ok {
			device.Metadata["mode"] = mode
		}
	}

	device.LastSeen = time.Now()
	return nil
}

// controlLock controls a lock device
func (im *IoTManager) controlLock(device *IoTDevice, command string, params map[string]interface{}) error {
	switch command {
	case "lock":
		device.Metadata["is_locked"] = true
	case "unlock":
		device.Metadata["is_locked"] = false
	}

	device.LastSeen = time.Now()
	return nil
}

// controlCamera controls a camera device
func (im *IoTManager) controlCamera(device *IoTDevice, command string, params map[string]interface{}) error {
	switch command {
	case "start_recording":
		device.Metadata["is_recording"] = true
	case "stop_recording":
		device.Metadata["is_recording"] = false
	case "snapshot":
		device.Metadata["snapshot_taken"] = time.Now()
	}

	device.LastSeen = time.Now()
	return nil
}

// CreateScene creates a device scene
func (im *IoTManager) CreateScene(name string, devices map[string]interface{}) (*Scene, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	scene := &Scene{
		ID:        fmt.Sprintf("scene-%d", time.Now().UnixNano()),
		Name:      name,
		Devices:   devices,
		IsActive:  false,
		CreatedAt: time.Now(),
	}

	im.scenes[scene.ID] = scene
	return scene, nil
}

// ActivateScene activates a scene
func (im *IoTManager) ActivateScene(ctx context.Context, sceneID string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	scene, exists := im.scenes[sceneID]
	if !exists {
		return fmt.Errorf("scene not found")
	}

	// Deactivate other scenes
	for _, s := range im.scenes {
		s.IsActive = false
	}

	// Activate this scene
	scene.IsActive = true

	// Apply scene settings
	for deviceID, state := range scene.Devices {
		if device, exists := im.devices[deviceID]; exists {
			// Apply state to device
			device.Metadata["scene_state"] = state
		}
	}

	return nil
}

// CreateAutomation creates an automation
func (im *IoTManager) CreateAutomation(name string, trigger Trigger, conditions []Condition, actions []Action) (*Automation, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	automation := &Automation{
		ID:         fmt.Sprintf("automation-%d", time.Now().UnixNano()),
		Name:       name,
		Trigger:    trigger,
		Conditions: conditions,
		Actions:    actions,
		IsActive:   true,
	}

	im.automations = append(im.automations, automation)
	return automation, nil
}

// RemoveAutomation removes an automation
func (im *IoTManager) RemoveAutomation(automationID string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	for i, automation := range im.automations {
		if automation.ID == automationID {
			im.automations = append(im.automations[:i], im.automations[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("automation not found")
}

// CheckAutomations checks and triggers automations
func (im *IoTManager) CheckAutomations(ctx context.Context) {
	im.mu.RLock()
	automations := make([]*Automation, len(im.automations))
	copy(automations, im.automations)
	im.mu.RUnlock()

	now := time.Now()

	for _, automation := range automations {
		if !automation.IsActive {
			continue
		}

		// Check trigger
		triggered := false
		switch automation.Trigger.Type {
		case "time":
			if now.Hour() == automation.Trigger.Time.Hour() && now.Minute() == automation.Trigger.Time.Minute() {
				triggered = true
			}
		case "device":
			if device, exists := im.devices[automation.Trigger.DeviceID]; exists {
				if !device.IsConnected {
					triggered = true
				}
			}
		}

		if triggered {
			// Check conditions
			conditionsMet := true
			for _, condition := range automation.Conditions {
				if !im.checkCondition(condition) {
					conditionsMet = false
					break
				}
			}

			if conditionsMet {
				// Execute actions
				for _, action := range automation.Actions {
					im.executeAction(ctx, action)
				}

				automation.LastTriggered = now
			}
		}
	}
}

// checkCondition checks if a condition is met
func (im *IoTManager) checkCondition(condition Condition) bool {
	switch condition.Type {
	case "time_range":
		now := time.Now()
		return now.After(condition.StartTime) && now.Before(condition.EndTime)
	case "device_state":
		if device, exists := im.devices[condition.DeviceID]; exists {
			return device.Metadata["state"] == condition.State
		}
	}
	return false
}

// executeAction executes an automation action
func (im *IoTManager) executeAction(ctx context.Context, action Action) {
	switch action.Type {
	case "device_control":
		im.ControlDevice(ctx, action.DeviceID, action.Command, action.Parameters)
	case "scene":
		im.ActivateScene(ctx, action.SceneID)
	}
}

// GetIoTStats gets IoT statistics
func (im *IoTManager) GetIoTStats() map[string]interface{} {
	im.mu.RLock()
	defer im.mu.RUnlock()

	totalDevices := len(im.devices)
	onlineDevices := 0
	offlineDevices := 0

	for _, device := range im.devices {
		if device.IsConnected {
			onlineDevices++
		} else {
			offlineDevices++
		}
	}

	return map[string]interface{}{
		"total_devices":     totalDevices,
		"online_devices":    onlineDevices,
		"offline_devices":   offlineDevices,
		"total_scenes":      len(im.scenes),
		"total_automations": len(im.automations),
		"device_types": map[string]int{
			"lights":       len(im.GetDevicesByType(DeviceTypeLight)),
			"thermostats":  len(im.GetDevicesByType(DeviceTypeThermostat)),
			"locks":        len(im.GetDevicesByType(DeviceTypeLock)),
			"cameras":      len(im.GetDevicesByType(DeviceTypeCamera)),
		},
	}
}
