package ui

import (
	"sync"
	"time"
)

type ThemeColor struct {
	Name        string
	LightValue  string
	DarkValue   string
	Description string
}

type ThemeManager struct {
	mu       sync.RWMutex
	current  UITheme
	colors   map[string]*ThemeColor
	callback func(UITheme)
}

func NewThemeManager() *ThemeManager {
	tm := &ThemeManager{
		current: UITheme(ThemeDark),
		colors:  make(map[string]*ThemeColor),
	}

	tm.initDefaultColors()
	return tm
}

func (tm *ThemeManager) initDefaultColors() {
	colors := []*ThemeColor{
		{Name: "primary", LightValue: "#007AFF", DarkValue: "#0A84FF", Description: "Primary brand color"},
		{Name: "secondary", LightValue: "#5856D6", DarkValue: "#5E5CE6", Description: "Secondary brand color"},
		{Name: "background", LightValue: "#FFFFFF", DarkValue: "#1C1C1E", Description: "Background color"},
		{Name: "surface", LightValue: "#F2F2F7", DarkValue: "#2C2C2E", Description: "Surface color"},
		{Name: "text", LightValue: "#000000", DarkValue: "#FFFFFF", Description: "Primary text color"},
		{Name: "textSecondary", LightValue: "#8E8E93", DarkValue: "#8E8E93", Description: "Secondary text color"},
		{Name: "border", LightValue: "#C6C6C8", DarkValue: "#38383A", Description: "Border color"},
		{Name: "success", LightValue: "#34C759", DarkValue: "#30D158", Description: "Success color"},
		{Name: "warning", LightValue: "#FF9500", DarkValue: "#FF9F0A", Description: "Warning color"},
		{Name: "error", LightValue: "#FF3B30", DarkValue: "#FF453A", Description: "Error color"},
		{Name: "info", LightValue: "#5AC8FA", DarkValue: "#64D2FF", Description: "Info color"},
	}

	for _, c := range colors {
		tm.colors[c.Name] = c
	}
}

func (tm *ThemeManager) SetTheme(theme UITheme) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.current = theme

	if tm.callback != nil {
		tm.callback(theme)
	}
}

func (tm *ThemeManager) GetTheme() UITheme {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.current
}

func (tm *ThemeManager) GetColor(name string) string {
	tm.mu.RLock()
	theme := tm.current
	color, ok := tm.colors[name]
	tm.mu.RUnlock()

	if !ok {
		return "#000000"
	}

	if theme == UITheme(ThemeDark) {
		return color.DarkValue
	}
	return color.LightValue
}

func (tm *ThemeManager) GetAllColors() map[string]string {
	tm.mu.RLock()
	theme := tm.current
	colors := tm.colors
	tm.mu.RUnlock()

	result := make(map[string]string)
	for name, color := range colors {
		if theme == UITheme(ThemeDark) {
			result[name] = color.DarkValue
		} else {
			result[name] = color.LightValue
		}
	}
	return result
}

func (tm *ThemeManager) SetOnChange(callback func(UITheme)) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.callback = callback
}

func (tm *ThemeManager) AddColor(color *ThemeColor) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.colors[color.Name] = color
}

type Animation struct {
	ID        string
	Duration  time.Duration
	Easing    string
	Delay     time.Duration
	OnComplete func()
}

type AnimationManager struct {
	mu          sync.RWMutex
	animations  map[string]*Animation
	enabled     bool
	speed       float64
}

func NewAnimationManager() *AnimationManager {
	return &AnimationManager{
		animations: make(map[string]*Animation),
		enabled:    true,
		speed:      1.0,
	}
}

func (am *AnimationManager) Add(anim *Animation) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.animations[anim.ID] = anim
}

func (am *AnimationManager) Remove(id string) {
	am.mu.Lock()
	defer am.mu.Unlock()
	delete(am.animations, id)
}

func (am *AnimationManager) Get(id string) (*Animation, bool) {
	am.mu.RLock()
	defer am.mu.RUnlock()
	anim, ok := am.animations[id]
	return anim, ok
}

func (am *AnimationManager) SetEnabled(enabled bool) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.enabled = enabled
}

func (am *AnimationManager) IsEnabled() bool {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.enabled
}

func (am *AnimationManager) SetSpeed(speed float64) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.speed = speed
}

func (am *AnimationManager) GetSpeed() float64 {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.speed
}

func (am *AnimationManager) GetDuration(id string) time.Duration {
	am.mu.RLock()
	enabled := am.enabled
	speed := am.speed
	anim, ok := am.animations[id]
	am.mu.RUnlock()

	if !ok || !enabled {
		return 0
	}

	duration := time.Duration(float64(anim.Duration) / speed)
	return duration
}

type FeedbackManager struct {
	mu          sync.RWMutex
	haptic      bool
	sound       bool
	visual      bool
	feedbacks   map[string]*Feedback
}

type Feedback struct {
	ID       string
	Type     string
	Haptic   bool
	Sound    string
	Visual   string
	Duration time.Duration
}

func NewFeedbackManager() *FeedbackManager {
	fm := &FeedbackManager{
		haptic:   true,
		sound:    true,
		visual:   true,
		feedbacks: make(map[string]*Feedback),
	}

	fm.initDefaultFeedbacks()
	return fm
}

func (fm *FeedbackManager) initDefaultFeedbacks() {
	feedbacks := []*Feedback{
		{ID: "message_sent", Type: "success", Sound: "send", Visual: "fade"},
		{ID: "message_received", Type: "info", Sound: "notification", Visual: "highlight"},
		{ID: "call_incoming", Type: "warning", Sound: "ring", Visual: "pulse"},
		{ID: "call_connected", Type: "success", Sound: "connect", Visual: "fade"},
		{ID: "file_received", Type: "success", Sound: "complete", Visual: "slide"},
		{ID: "error", Type: "error", Sound: "error", Visual: "shake"},
		{ID: "button_click", Type: "light", Sound: "click", Visual: "ripple"},
		{ID: "success", Type: "success", Sound: "success", Visual: "checkmark"},
	}

	for _, f := range feedbacks {
		fm.feedbacks[f.ID] = f
	}
}

func (fm *FeedbackManager) Trigger(feedbackID string) {
	fm.mu.RLock()
	enabled := fm.haptic && fm.sound && fm.visual
	_, ok := fm.feedbacks[feedbackID]
	fm.mu.RUnlock()

	if !ok || !enabled {
		return
	}

	// Trigger haptic feedback
	// Trigger sound
	// Trigger visual feedback
}

func (fm *FeedbackManager) SetHaptic(enabled bool) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.haptic = enabled
}

func (fm *FeedbackManager) SetSound(enabled bool) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.sound = enabled
}

func (fm *FeedbackManager) SetVisual(enabled bool) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.visual = enabled
}

func (fm *FeedbackManager) IsHapticEnabled() bool {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.haptic
}

func (fm *FeedbackManager) IsSoundEnabled() bool {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.sound
}

func (fm *FeedbackManager) IsVisualEnabled() bool {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.visual
}

type ToastManager struct {
	mu       sync.RWMutex
	toasts   []*Toast
	maxCount int
	duration time.Duration
	callback func(*Toast)
}

type Toast struct {
	ID        string
	Message   string
	Type      string
	Duration  time.Duration
	Action    *ToastAction
	OnDismiss func()
}

type ToastAction struct {
	Label string
	Action func()
}

func NewToastManager() *ToastManager {
	return &ToastManager{
		toasts:   make([]*Toast, 0),
		maxCount: 3,
		duration: 3 * time.Second,
	}
}

func (tm *ToastManager) Show(message string, toastType string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	toast := &Toast{
		ID:       generateToastID(),
		Message:  message,
		Type:     toastType,
		Duration: tm.duration,
	}

	tm.toasts = append(tm.toasts, toast)

	if len(tm.toasts) > tm.maxCount {
		tm.toasts = tm.toasts[1:]
	}

	if tm.callback != nil {
		tm.callback(toast)
	}
}

func (tm *ToastManager) ShowWithAction(message string, toastType string, label string, action func()) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	toast := &Toast{
		ID:       generateToastID(),
		Message:  message,
		Type:     toastType,
		Duration: tm.duration,
		Action: &ToastAction{
			Label: label,
			Action: action,
		},
	}

	tm.toasts = append(tm.toasts, toast)

	if len(tm.toasts) > tm.maxCount {
		tm.toasts = tm.toasts[1:]
	}

	if tm.callback != nil {
		tm.callback(toast)
	}
}

func (tm *ToastManager) Dismiss(id string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for i, t := range tm.toasts {
		if t.ID == id {
			if t.OnDismiss != nil {
				t.OnDismiss()
			}
			tm.toasts = append(tm.toasts[:i], tm.toasts[i+1:]...)
			return
		}
	}
}

func (tm *ToastManager) Clear() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.toasts = make([]*Toast, 0)
}

func (tm *ToastManager) GetAll() []*Toast {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	result := make([]*Toast, len(tm.toasts))
	copy(result, tm.toasts)
	return result
}

func (tm *ToastManager) SetMaxCount(count int) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.maxCount = count
}

func (tm *ToastManager) SetDuration(duration time.Duration) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.duration = duration
}

func (tm *ToastManager) SetCallback(callback func(*Toast)) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.callback = callback
}

func generateToastID() string {
	return "toast_" + time.Now().Format("20060102150405")
}

type LoadingManager struct {
	mu       sync.RWMutex
	loadings map[string]*Loading
}

type Loading struct {
	ID          string
	Message     string
	Progress    float64
	Indeterminate bool
	Cancelable  bool
	OnCancel    func()
}

func NewLoadingManager() *LoadingManager {
	return &LoadingManager{
		loadings: make(map[string]*Loading),
	}
}

func (lm *LoadingManager) Show(message string) string {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	id := generateLoadingID()
	loading := &Loading{
		ID:      id,
		Message: message,
	}

	lm.loadings[id] = loading
	return id
}

func (lm *LoadingManager) ShowWithProgress(message string, progress float64) string {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	id := generateLoadingID()
	loading := &Loading{
		ID:          id,
		Message:     message,
		Progress:    progress,
		Indeterminate: progress < 0,
	}

	lm.loadings[id] = loading
	return id
}

func (lm *LoadingManager) UpdateProgress(id string, progress float64) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if loading, ok := lm.loadings[id]; ok {
		loading.Progress = progress
		if progress >= 100 {
			delete(lm.loadings, id)
		}
	}
}

func (lm *LoadingManager) Dismiss(id string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	delete(lm.loadings, id)
}

func (lm *LoadingManager) GetAll() []*Loading {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	loadings := make([]*Loading, 0, len(lm.loadings))
	for _, l := range lm.loadings {
		loadings = append(loadings, l)
	}
	return loadings
}

func generateLoadingID() string {
	return "loading_" + time.Now().Format("20060102150405")
}
