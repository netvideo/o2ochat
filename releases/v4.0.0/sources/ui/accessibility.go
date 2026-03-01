package ui

import (
	"sync"
)

type AccessibilityRole string

const (
	RoleButton        AccessibilityRole = "button"
	RoleCheckbox      AccessibilityRole = "checkbox"
	RoleCombobox     AccessibilityRole = "combobox"
	RoleDialog        AccessibilityRole = "dialog"
	RoleGrid          AccessibilityRole = "grid"
	RoleImage         AccessibilityRole = "image"
	RoleLink          AccessibilityRole = "link"
	RoleList          AccessibilityRole = "list"
	RoleMenu          AccessibilityRole = "menu"
	RoleMenuItem      AccessibilityRole = "menuitem"
	RoleOption        AccessibilityRole = "option"
	RoleProgressBar   AccessibilityRole = "progressbar"
	RoleRadio         AccessibilityRole = "radio"
	RoleSlider        AccessibilityRole = "slider"
	RoleTab           AccessibilityRole = "tab"
	RoleTable         AccessibilityRole = "table"
	RoleTextbox       AccessibilityRole = "textbox"
	RoleTooltip       AccessibilityRole = "tooltip"
	RoleTree          AccessibilityRole = "tree"
	RoleTreeItem      AccessibilityRole = "treeitem"
	RoleWindow        AccessibilityRole = "window"
	RoleApplication   AccessibilityRole = "application"
)

type AccessibilityState struct {
	Disabled    bool
	Expanded   bool
	Selected   bool
	Checked    bool
	Busy       bool
	ReadOnly   bool
	Required   bool
	Hidden     bool
	Invalid    bool
	Pressed    bool
	Hovered    bool
	Focused    bool
}

type AccessibilityDescription struct {
	Label       string
	Description string
	Placeholder string
	HelperText string
	ErrorText  string
	HintText   string
}

type AccessibilityProperties struct {
	Role          AccessibilityRole
	Label        string
	Description  string
	Shortcut     string
	Value        string
	MinValue     float64
	MaxValue     float64
	ValueNow     float64
	Level        int
	ItemsCount   int
	ItemsAfter   int
	ItemsBefore int
	SetSize      int
	PosInSet     int
}

type AccessibilityAction struct {
	Name  string
	Label string
}

type AccessibilityComponent struct {
	mu           sync.RWMutex
	id           string
	properties  AccessibilityProperties
	state       AccessibilityState
	description AccessibilityDescription
	actions     []*AccessibilityAction
	parent      string
	children    []string
	onFocus     func()
	onBlur      func()
	onAction    func(action string)
}

func NewAccessibilityComponent(id string, role AccessibilityRole) *AccessibilityComponent {
	return &AccessibilityComponent{
		id:          id,
		properties:  AccessibilityProperties{Role: role},
		state:       AccessibilityState{},
		description: AccessibilityDescription{},
		actions:     make([]*AccessibilityAction, 0),
		children:    make([]string, 0),
	}
}

func (ac *AccessibilityComponent) GetID() string {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.id
}

func (ac *AccessibilityComponent) SetRole(role AccessibilityRole) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.properties.Role = role
}

func (ac *AccessibilityComponent) GetRole() AccessibilityRole {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.properties.Role
}

func (ac *AccessibilityComponent) SetLabel(label string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.properties.Label = label
}

func (ac *AccessibilityComponent) GetLabel() string {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.properties.Label
}

func (ac *AccessibilityComponent) SetDescription(description string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.description.Description = description
}

func (ac *AccessibilityComponent) GetDescription() string {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.description.Description
}

func (ac *AccessibilityComponent) SetShortcut(shortcut string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.properties.Shortcut = shortcut
}

func (ac *AccessibilityComponent) GetShortcut() string {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.properties.Shortcut
}

func (ac *AccessibilityComponent) SetValue(value string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.properties.Value = value
}

func (ac *AccessibilityComponent) GetValue() string {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.properties.Value
}

func (ac *AccessibilityComponent) SetState(state AccessibilityState) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.state = state
}

func (ac *AccessibilityComponent) GetState() AccessibilityState {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.state
}

func (ac *AccessibilityComponent) SetDisabled(disabled bool) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.state.Disabled = disabled
}

func (ac *AccessibilityComponent) IsDisabled() bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.state.Disabled
}

func (ac *AccessibilityComponent) SetSelected(selected bool) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.state.Selected = selected
}

func (ac *AccessibilityComponent) IsSelected() bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.state.Selected
}

func (ac *AccessibilityComponent) SetFocused(focused bool) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.state.Focused = focused

	if focused && ac.onFocus != nil {
		ac.onFocus()
	} else if !focused && ac.onBlur != nil {
		ac.onBlur()
	}
}

func (ac *AccessibilityComponent) IsFocused() bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.state.Focused
}

func (ac *AccessibilityComponent) SetExpanded(expanded bool) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.state.Expanded = expanded
}

func (ac *AccessibilityComponent) IsExpanded() bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.state.Expanded
}

func (ac *AccessibilityComponent) SetChecked(checked bool) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.state.Checked = checked
}

func (ac *AccessibilityComponent) IsChecked() bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.state.Checked
}

func (ac *AccessibilityComponent) SetHidden(hidden bool) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.state.Hidden = hidden
}

func (ac *AccessibilityComponent) IsHidden() bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.state.Hidden
}

func (ac *AccessibilityComponent) SetErrorText(errorText string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.description.ErrorText = errorText
}

func (ac *AccessibilityComponent) GetErrorText() string {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.description.ErrorText
}

func (ac *AccessibilityComponent) AddAction(name, label string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.actions = append(ac.actions, &AccessibilityAction{Name: name, Label: label})
}

func (ac *AccessibilityComponent) GetActions() []*AccessibilityAction {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.actions
}

func (ac *AccessibilityComponent) PerformAction(actionName string) {
	ac.mu.RLock()
	disabled := ac.state.Disabled
	ac.mu.RUnlock()

	if disabled {
		return
	}

	if ac.onAction != nil {
		ac.onAction(actionName)
	}
}

func (ac *AccessibilityComponent) SetOnFocus(callback func()) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.onFocus = callback
}

func (ac *AccessibilityComponent) SetOnBlur(callback func()) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.onBlur = callback
}

func (ac *AccessibilityComponent) SetOnAction(callback func(action string)) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.onAction = callback
}

func (ac *AccessibilityComponent) GetText() string {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	text := ac.properties.Label
	if ac.description.Description != "" {
		text += ". " + ac.description.Description
	}
	if ac.description.ErrorText != "" {
		text += ". Error: " + ac.description.ErrorText
	}
	if ac.properties.Shortcut != "" {
		text += ". Shortcut: " + ac.properties.Shortcut
	}

	return text
}

type FocusManager struct {
	mu          sync.RWMutex
	components  map[string]*AccessibilityComponent
	focusedID   string
	order       []string
	cycleFocus  bool
	onFocusChange func(from, to string)
}

func NewFocusManager() *FocusManager {
	return &FocusManager{
		components: make(map[string]*AccessibilityComponent),
		order:      make([]string, 0),
		cycleFocus: true,
	}
}

func (fm *FocusManager) Register(component *AccessibilityComponent) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	id := component.GetID()
	fm.components[id] = component
	fm.order = append(fm.order, id)
}

func (fm *FocusManager) Unregister(id string) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	delete(fm.components, id)
	for i, oid := range fm.order {
		if oid == id {
			fm.order = append(fm.order[:i], fm.order[i+1:]...)
			break
		}
	}
}

func (fm *FocusManager) Focus(id string) bool {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	component, ok := fm.components[id]
	if !ok {
		return false
	}

	if component.IsDisabled() {
		return false
	}

	oldID := fm.focusedID
	fm.focusedID = id

	if component, ok := fm.components[id]; ok {
		component.SetFocused(true)
	}

	if oldID != "" && oldID != id {
		if oldComponent, ok := fm.components[oldID]; ok {
			oldComponent.SetFocused(false)
		}
	}

	if fm.onFocusChange != nil {
		fm.onFocusChange(oldID, id)
	}

	return true
}

func (fm *FocusManager) GetFocusedID() string {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.focusedID
}

func (fm *FocusManager) FocusNext() bool {
	fm.mu.Lock()
	order := make([]string, len(fm.order))
	copy(order, fm.order)
	current := fm.focusedID
	fm.mu.Unlock()

	if len(order) == 0 {
		return false
	}

	currentIndex := -1
	for i, id := range order {
		if id == current {
			currentIndex = i
			break
		}
	}

	nextIndex := currentIndex + 1
	if nextIndex >= len(order) {
		if fm.cycleFocus {
			nextIndex = 0
		} else {
			return false
		}
	}

	return fm.Focus(order[nextIndex])
}

func (fm *FocusManager) FocusPrevious() bool {
	fm.mu.Lock()
	order := make([]string, len(fm.order))
	copy(order, fm.order)
	current := fm.focusedID
	fm.mu.Unlock()

	if len(order) == 0 {
		return false
	}

	currentIndex := -1
	for i, id := range order {
		if id == current {
			currentIndex = i
			break
		}
	}

	prevIndex := currentIndex - 1
	if prevIndex < 0 {
		if fm.cycleFocus {
			prevIndex = len(order) - 1
		} else {
			return false
		}
	}

	return fm.Focus(order[prevIndex])
}

func (fm *FocusManager) ClearFocus() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if fm.focusedID != "" {
		if component, ok := fm.components[fm.focusedID]; ok {
			component.SetFocused(false)
		}
		fm.focusedID = ""
	}
}

func (fm *FocusManager) SetOnFocusChange(callback func(from, to string)) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.onFocusChange = callback
}

func (fm *FocusManager) SetCycleFocus(cycle bool) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.cycleFocus = cycle
}

func (fm *FocusManager) GetComponent(id string) (*AccessibilityComponent, bool) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	component, ok := fm.components[id]
	return component, ok
}

type ScreenReader struct {
	mu          sync.RWMutex
	enabled     bool
	announcements []string
	callback    func(string)
	focusManager *FocusManager
}

func NewScreenReader() *ScreenReader {
	return &ScreenReader{
		enabled:      false,
		announcements: make([]string, 0),
		focusManager: NewFocusManager(),
	}
}

func (sr *ScreenReader) Enable() {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.enabled = true
}

func (sr *ScreenReader) Disable() {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.enabled = false
}

func (sr *ScreenReader) IsEnabled() bool {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	return sr.enabled
}

func (sr *ScreenReader) Announce(message string) {
	sr.mu.Lock()
	enabled := sr.enabled
	sr.mu.RUnlock()

	if !enabled {
		return
	}

	sr.mu.Lock()
	sr.announcements = append(sr.announcements, message)
	if len(sr.announcements) > 10 {
		sr.announcements = sr.announcements[1:]
	}
	sr.mu.Unlock()

	if sr.callback != nil {
		sr.callback(message)
	}
}

func (sr *ScreenReader) AnnounceFocused() {
	sr.mu.RLock()
	focusManager := sr.focusManager
	sr.mu.RUnlock()

	focusedID := focusManager.GetFocusedID()
	if focusedID == "" {
		return
	}

	component, ok := focusManager.GetComponent(focusedID)
	if !ok {
		return
	}

	text := component.GetText()
	sr.Announce(text)
}

func (sr *ScreenReader) SetCallback(callback func(string)) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.callback = callback
}

func (sr *ScreenReader) GetFocusManager() *FocusManager {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	return sr.focusManager
}
