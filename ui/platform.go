package ui

import (
	"runtime"
	"sync"
)

type Platform string

const (
	PlatformWindows Platform = "windows"
	PlatformMac     Platform = "darwin"
	PlatformLinux   Platform = "linux"
	PlatformWeb     Platform = "web"
	PlatformMobile  Platform = "mobile"
	PlatformUnknown Platform = "unknown"
)

type Architecture string

const (
	ArchAMD64 Architecture = "amd64"
	ArchARM64 Architecture = "arm64"
	Arch386   Architecture = "386"
	ArchARM   Architecture = "arm"
)

type DeviceType string

const (
	DeviceTypeDesktop DeviceType = "desktop"
	DeviceTypeTablet  DeviceType = "tablet"
	DeviceTypePhone   DeviceType = "phone"
	DeviceTypeWeb    DeviceType = "web"
)

type PlatformInfo struct {
	Platform     Platform
	Architecture Architecture
	DeviceType   DeviceType
	OSVersion    string
	AppVersion   string
	HighDPI      bool
	TouchEnabled bool
}

type PlatformDetector struct {
	mu      sync.RWMutex
	info    *PlatformInfo
	once    sync.Once
}

var (
	detector *PlatformDetector
	detectorOnce sync.Once
)

func GetPlatformDetector() *PlatformDetector {
	detectorOnce.Do(func() {
		detector = &PlatformDetector{
			info: detectPlatform(),
		}
	})
	return detector
}

func detectPlatform() *PlatformInfo {
	info := &PlatformInfo{
		OSVersion:   runtime.GOOS,
		AppVersion:  "1.0.0",
		HighDPI:    true,
		TouchEnabled: false,
	}

	switch runtime.GOOS {
	case "windows":
		info.Platform = PlatformWindows
		info.DeviceType = DeviceTypeDesktop
	case "darwin":
		info.Platform = PlatformMac
		info.DeviceType = DeviceTypeDesktop
	case "linux":
		info.Platform = PlatformLinux
		info.DeviceType = DeviceTypeDesktop
	default:
		info.Platform = PlatformUnknown
		info.DeviceType = DeviceTypeDesktop
	}

	switch runtime.GOARCH {
	case "amd64":
		info.Architecture = ArchAMD64
	case "arm64":
		info.Architecture = ArchARM64
	case "386":
		info.Architecture = Arch386
	case "arm":
		info.Architecture = ArchARM
	}

	return info
}

func (pd *PlatformDetector) GetInfo() *PlatformInfo {
	pd.mu.RLock()
	defer pd.mu.RUnlock()
	return pd.info
}

func (pd *PlatformDetector) IsMobile() bool {
	pd.mu.RLock()
	defer pd.mu.RUnlock()
	return pd.info.DeviceType == DeviceTypePhone || pd.info.DeviceType == DeviceTypeTablet
}

func (pd *PlatformDetector) IsDesktop() bool {
	pd.mu.RLock()
	defer pd.mu.RUnlock()
	return pd.info.DeviceType == DeviceTypeDesktop
}

func (pd *PlatformDetector) IsWindows() bool {
	pd.mu.RLock()
	defer pd.mu.RUnlock()
	return pd.info.Platform == PlatformWindows
}

func (pd *PlatformDetector) IsMac() bool {
	pd.mu.RLock()
	defer pd.mu.RUnlock()
	return pd.info.Platform == PlatformMac
}

func (pd *PlatformDetector) IsLinux() bool {
	pd.mu.RLock()
	defer pd.mu.RUnlock()
	return pd.info.Platform == PlatformLinux
}

type PlatformFeature struct {
	Name        string
	Supported   bool
	Required    bool
	Description string
}

type FeatureDetector struct {
	mu        sync.RWMutex
	features  map[string]*PlatformFeature
	platform  Platform
}

func NewFeatureDetector(platform Platform) *FeatureDetector {
	fd := &FeatureDetector{
		platform: platform,
		features: make(map[string]*PlatformFeature),
	}

	fd.initDefaultFeatures()
	return fd
}

func (fd *FeatureDetector) initDefaultFeatures() {
	features := []*PlatformFeature{
		{Name: "tray_icon", Supported: true, Required: false, Description: "System tray icon support"},
		{Name: "notifications", Supported: true, Required: false, Description: "Desktop notifications"},
		{Name: "autostart", Supported: false, Required: false, Description: "Auto start on login"},
		{Name: "background", Supported: true, Required: false, Description: "Run in background"},
		{Name: "shortcuts", Supported: true, Required: false, Description: "Global shortcuts"},
		{Name: "drag_drop", Supported: true, Required: false, Description: "Drag and drop files"},
		{Name: "clipboard", Supported: true, Required: false, Description: "Clipboard access"},
		{Name: "camera", Supported: true, Required: false, Description: "Camera access"},
		{Name: "microphone", Supported: true, Required: false, Description: "Microphone access"},
		{Name: "filesystem", Supported: true, Required: false, Description: "Filesystem access"},
		{Name: "push_notification", Supported: false, Required: false, Description: "Push notifications"},
		{Name: "widgets", Supported: false, Required: false, Description: "Desktop widgets"},
		{Name: "a11y", Supported: true, Required: false, Description: "Accessibility support"},
	}

	for _, f := range features {
		fd.features[f.Name] = f
	}
}

func (fd *FeatureDetector) IsSupported(featureName string) bool {
	fd.mu.RLock()
	defer fd.mu.RUnlock()

	if feature, ok := fd.features[featureName]; ok {
		return feature.Supported
	}
	return false
}

func (fd *FeatureDetector) GetFeature(featureName string) (*PlatformFeature, bool) {
	fd.mu.RLock()
	defer fd.mu.RUnlock()
	feature, ok := fd.features[featureName]
	return feature, ok
}

func (fd *FeatureDetector) SetSupported(featureName string, supported bool) {
	fd.mu.Lock()
	defer fd.mu.Unlock()

	if feature, ok := fd.features[featureName]; ok {
		feature.Supported = supported
	}
}

func (fd *FeatureDetector) GetAllFeatures() []*PlatformFeature {
	fd.mu.RLock()
	defer fd.mu.RUnlock()

	features := make([]*PlatformFeature, 0, len(fd.features))
	for _, f := range fd.features {
		features = append(features, f)
	}
	return features
}

func (fd *FeatureDetector) GetRequiredFeatures() []*PlatformFeature {
	fd.mu.RLock()
	defer fd.mu.RUnlock()

	var required []*PlatformFeature
	for _, f := range fd.features {
		if f.Required {
			required = append(required, f)
		}
	}
	return required
}

type PlatformAdapter struct {
	mu             sync.RWMutex
	platformInfo   *PlatformInfo
	featureDetector *FeatureDetector
	theme          *PlatformTheme
	shortcuts      *ShortcutManager
	a11y           *AccessibilityManager
}

type PlatformTheme struct {
	PrimaryColor    string
	SecondaryColor  string
	BackgroundColor string
	TextColor      string
	FontFamily     string
	FontSize       int
	BorderRadius   int
	Spacing        int
}

func NewPlatformAdapter() *PlatformAdapter {
	detector := GetPlatformDetector()
	info := detector.GetInfo()

	pa := &PlatformAdapter{
		platformInfo:   info,
		featureDetector: NewFeatureDetector(info.Platform),
		theme:          NewPlatformTheme(info.Platform),
		shortcuts:      NewShortcutManager(info.Platform),
		a11y:           NewAccessibilityManager(),
	}

	return pa
}

func NewPlatformTheme(platform Platform) *PlatformTheme {
	theme := &PlatformTheme{
		FontSize:     14,
		BorderRadius: 4,
		Spacing:      8,
	}

	switch platform {
	case PlatformWindows:
		theme.PrimaryColor = "#0078D4"
		theme.FontFamily = "Segoe UI"
	case PlatformMac:
		theme.PrimaryColor = "#007AFF"
		theme.FontFamily = "SF Pro"
	case PlatformLinux:
		theme.PrimaryColor = "#3584E4"
		theme.FontFamily = "Ubuntu"
	default:
		theme.PrimaryColor = "#0078D4"
		theme.FontFamily = "System"
	}

	theme.SecondaryColor = theme.PrimaryColor
	theme.BackgroundColor = "#FFFFFF"
	theme.TextColor = "#000000"

	return theme
}

func (pa *PlatformAdapter) GetPlatformInfo() *PlatformInfo {
	pa.mu.RLock()
	defer pa.mu.RUnlock()
	return pa.platformInfo
}

func (pa *PlatformAdapter) GetTheme() *PlatformTheme {
	pa.mu.RLock()
	defer pa.mu.RUnlock()
	return pa.theme
}

func (pa *PlatformAdapter) GetFeatureDetector() *FeatureDetector {
	pa.mu.RLock()
	defer pa.mu.RUnlock()
	return pa.featureDetector
}

func (pa *PlatformAdapter) GetShortcutManager() *ShortcutManager {
	pa.mu.RLock()
	defer pa.mu.RUnlock()
	return pa.shortcuts
}

func (pa *PlatformAdapter) GetAccessibilityManager() *AccessibilityManager {
	pa.mu.RLock()
	defer pa.mu.RUnlock()
	return pa.a11y
}

type ShortcutManager struct {
	mu        sync.RWMutex
	platform  Platform
	shortcuts map[string]*Shortcut
}

type Shortcut struct {
	ID          string
	Keys        []string
	Description string
	Action      func()
}

func NewShortcutManager(platform Platform) *ShortcutManager {
	sm := &ShortcutManager{
		platform:  platform,
		shortcuts: make(map[string]*Shortcut),
	}

	sm.registerDefaultShortcuts()
	return sm
}

func (sm *ShortcutManager) registerDefaultShortcuts() {
	shortcuts := []*Shortcut{
		{ID: "send_message", Keys: []string{"Enter"}, Description: "Send message"},
		{ID: "new_line", Keys: []string{"Shift", "Enter"}, Description: "Insert new line"},
		{ID: "search", Keys: []string{"Ctrl", "F"}, Description: "Search"},
		{ID: "settings", Keys: []string{"Ctrl", ","}, Description: "Open settings"},
		{ID: "quit", Keys: []string{"Ctrl", "Q"}, Description: "Quit application"},
		{ID: "call_voice", Keys: []string{"Ctrl", "C"}, Description: "Start voice call"},
		{ID: "call_video", Keys: []string{"Ctrl", "V"}, Description: "Start video call"},
		{ID: "mute", Keys: []string{"Ctrl", "M"}, Description: "Toggle mute"},
		{ID: "screenshot", Keys: []string{"Ctrl", "Shift", "S"}, Description: "Take screenshot"},
	}

	for _, s := range shortcuts {
		sm.shortcuts[s.ID] = s
	}
}

func (sm *ShortcutManager) Register(shortcut *Shortcut) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.shortcuts[shortcut.ID] = shortcut
}

func (sm *ShortcutManager) GetShortcut(id string) (*Shortcut, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	s, ok := sm.shortcuts[id]
	return s, ok
}

func (sm *ShortcutManager) GetAllShortcuts() []*Shortcut {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	shortcuts := make([]*Shortcut, 0, len(sm.shortcuts))
	for _, s := range sm.shortcuts {
		shortcuts = append(shortcuts, s)
	}
	return shortcuts
}

func (sm *ShortcutManager) SetAction(id string, action func()) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if s, ok := sm.shortcuts[id]; ok {
		s.Action = action
	}
}

type AccessibilityManager struct {
	mu              sync.RWMutex
	enabled         bool
	screenReader    bool
	highContrast    bool
	largeText       bool
	focusIndicators bool
	announcements   chan string
}

func NewAccessibilityManager() *AccessibilityManager {
	return &AccessibilityManager{
		enabled:          true,
		screenReader:     false,
		highContrast:    false,
		largeText:        false,
		focusIndicators: true,
		announcements:    make(chan string, 10),
	}
}

func (am *AccessibilityManager) IsEnabled() bool {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.enabled
}

func (am *AccessibilityManager) SetEnabled(enabled bool) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.enabled = enabled
}

func (am *AccessibilityManager) IsScreenReaderEnabled() bool {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.screenReader
}

func (am *AccessibilityManager) SetScreenReaderEnabled(enabled bool) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.screenReader = enabled
}

func (am *AccessibilityManager) Announce(message string) {
	am.mu.RLock()
	enabled := am.enabled
	am.mu.RUnlock()

	if enabled {
		select {
		case am.announcements <- message:
		default:
		}
	}
}

func (am *AccessibilityManager) GetAnnouncements() <-chan string {
	return am.announcements
}
