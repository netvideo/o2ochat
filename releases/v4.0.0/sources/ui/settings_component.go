package ui

import (
	"sync"
)

type SettingSection struct {
	Name        string
	Description string
	Settings    map[string]*Setting
}

type Setting struct {
	Key         string
	Value       interface{}
	DefaultValue interface{}
	Type        string
	Options     []string
	Min         float64
	Max         float64
	Description string
}

type SettingsComponent struct {
	mu           sync.RWMutex
	sections     map[string]*SettingSection
	config       *UIConfig
	onChange     func(section, key string, value interface{})
	onSave       func(config *UIConfig)
	onReset      func()
}

func NewSettingsComponent() *SettingsComponent {
	sc := &SettingsComponent{
		sections: make(map[string]*SettingSection),
		config: &UIConfig{
			Theme:          ThemeDark,
			Language:       "zh-CN",
			FontSize:       14,
			ShowAvatars:    true,
			ShowTimestamps: true,
			NotifySounds:   true,
			NotifyDesktop:  true,
			AutoStart:      false,
			MinimizeToTray: true,
		},
	}

	sc.initDefaultSections()
	return sc
}

func (sc *SettingsComponent) initDefaultSections() {
	sc.sections["general"] = &SettingSection{
		Name:        "通用",
		Description: "应用程序通用设置",
		Settings: map[string]*Setting{
			"theme": {
				Key:         "theme",
				Value:       "dark",
				DefaultValue: "dark",
				Type:        "select",
				Options:     []string{"light", "dark", "auto"},
				Description: "应用程序主题",
			},
			"language": {
				Key:         "language",
				Value:       "zh-CN",
				DefaultValue: "zh-CN",
				Type:        "select",
				Options:     []string{"zh-CN", "en-US"},
				Description: "界面语言",
			},
			"font_size": {
				Key:         "font_size",
				Value:       14,
				DefaultValue: 14,
				Type:        "number",
				Min:         10,
				Max:         24,
				Description: "字体大小",
			},
		},
	}

	sc.sections["chat"] = &SettingSection{
		Name:        "聊天",
		Description: "聊天相关设置",
		Settings: map[string]*Setting{
			"show_avatars": {
				Key:         "show_avatars",
				Value:       true,
				DefaultValue: true,
				Type:        "bool",
				Description: "显示头像",
			},
			"show_timestamps": {
				Key:         "show_timestamps",
				Value:       true,
				DefaultValue: true,
				Type:        "bool",
				Description: "显示时间戳",
			},
			"enter_to_send": {
				Key:         "enter_to_send",
				Value:       true,
				DefaultValue: true,
				Type:        "bool",
				Description: "按回车键发送消息",
			},
		},
	}

	sc.sections["notification"] = &SettingSection{
		Name:        "通知",
		Description: "通知相关设置",
		Settings: map[string]*Setting{
			"notify_sounds": {
				Key:         "notify_sounds",
				Value:       true,
				DefaultValue: true,
				Type:        "bool",
				Description: "通知声音",
			},
			"notify_desktop": {
				Key:         "notify_desktop",
				Value:       true,
				DefaultValue: true,
				Type:        "bool",
				Description: "桌面通知",
			},
			"notify_preview": {
				Key:         "notify_preview",
				Value:       true,
				DefaultValue: true,
				Type:        "bool",
				Description: "通知预览",
			},
		},
	}

	sc.sections["system"] = &SettingSection{
		Name:        "系统",
		Description: "系统相关设置",
		Settings: map[string]*Setting{
			"auto_start": {
				Key:         "auto_start",
				Value:       false,
				DefaultValue: false,
				Type:        "bool",
				Description: "开机自启动",
			},
			"minimize_to_tray": {
				Key:         "minimize_to_tray",
				Value:       true,
				DefaultValue: true,
				Type:        "bool",
				Description: "最小化到托盘",
			},
			"close_to_tray": {
				Key:         "close_to_tray",
				Value:       true,
				DefaultValue: true,
				Type:        "bool",
				Description: "关闭到托盘",
			},
		},
	}

	sc.sections["network"] = &SettingSection{
		Name:        "网络",
		Description: "网络相关设置",
		Settings: map[string]*Setting{
			"connection_timeout": {
				Key:         "connection_timeout",
				Value:       30,
				DefaultValue: 30,
				Type:        "number",
				Min:         5,
				Max:         120,
				Description: "连接超时（秒）",
			},
			"enable_stun": {
				Key:         "enable_stun",
				Value:       true,
				DefaultValue: true,
				Type:        "bool",
				Description: "启用STUN",
			},
			"enable_turn": {
				Key:         "enable_turn",
				Value:       true,
				DefaultValue: true,
				Type:        "bool",
				Description: "启用TURN",
			},
		},
	}

	sc.sections["storage"] = &SettingSection{
		Name:        "存储",
		Description: "存储相关设置",
		Settings: map[string]*Setting{
			"download_path": {
				Key:         "download_path",
				Value:       "",
				DefaultValue: "",
				Type:        "string",
				Description: "下载保存路径",
			},
			"auto_accept_file": {
				Key:         "auto_accept_file",
				Value:       false,
				DefaultValue: false,
				Type:        "bool",
				Description: "自动接收文件",
			},
			"max_cache_size": {
				Key:         "max_cache_size",
				Value:       1024,
				DefaultValue: 1024,
				Type:        "number",
				Min:         100,
				Max:         10240,
				Description: "最大缓存大小（MB）",
			},
		},
	}
}

func (sc *SettingsComponent) GetSection(name string) (*SettingSection, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	section, ok := sc.sections[name]
	return section, ok
}

func (sc *SettingsComponent) GetAllSections() []*SettingSection {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	sections := make([]*SettingSection, 0, len(sc.sections))
	for _, s := range sc.sections {
		sections = append(sections, s)
	}
	return sections
}

func (sc *SettingsComponent) GetSetting(section, key string) (*Setting, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if s, ok := sc.sections[section]; ok {
		setting, ok := s.Settings[key]
		return setting, ok
	}
	return nil, false
}

func (sc *SettingsComponent) SetSetting(section, key string, value interface{}) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if s, ok := sc.sections[section]; ok {
		if setting, ok := s.Settings[key]; ok {
			setting.Value = value

			sc.applySetting(section, key, value)

			if sc.onChange != nil {
				sc.onChange(section, key, value)
			}
			return nil
		}
	}
	return ErrResourceNotFound
}

func (sc *SettingsComponent) ResetSetting(section, key string) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if s, ok := sc.sections[section]; ok {
		if setting, ok := s.Settings[key]; ok {
			setting.Value = setting.DefaultValue

			sc.applySetting(section, key, setting.DefaultValue)
			return nil
		}
	}
	return ErrResourceNotFound
}

func (sc *SettingsComponent) ResetSection(section string) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if s, ok := sc.sections[section]; ok {
		for key, setting := range s.Settings {
			setting.Value = setting.DefaultValue
			sc.applySetting(section, key, setting.DefaultValue)
		}
		return nil
	}
	return ErrResourceNotFound
}

func (sc *SettingsComponent) ResetAll() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	for sectionName, s := range sc.sections {
		for key, setting := range s.Settings {
			setting.Value = setting.DefaultValue
			sc.applySetting(sectionName, key, setting.DefaultValue)
		}
	}

	if sc.onReset != nil {
		sc.onReset()
	}
	return nil
}

func (sc *SettingsComponent) applySetting(section, key string, value interface{}) {
	switch section {
	case "general":
		switch key {
		case "theme":
			if v, ok := value.(string); ok {
				sc.config.Theme = UITheme(v)
			}
		case "language":
			if v, ok := value.(string); ok {
				sc.config.Language = v
			}
		case "font_size":
			if v, ok := value.(int); ok {
				sc.config.FontSize = v
			}
		}
	case "chat":
		switch key {
		case "show_avatars":
			if v, ok := value.(bool); ok {
				sc.config.ShowAvatars = v
			}
		case "show_timestamps":
			if v, ok := value.(bool); ok {
				sc.config.ShowTimestamps = v
			}
		}
	case "notification":
		switch key {
		case "notify_sounds":
			if v, ok := value.(bool); ok {
				sc.config.NotifySounds = v
			}
		case "notify_desktop":
			if v, ok := value.(bool); ok {
				sc.config.NotifyDesktop = v
			}
		}
	case "system":
		switch key {
		case "auto_start":
			if v, ok := value.(bool); ok {
				sc.config.AutoStart = v
			}
		case "minimize_to_tray":
			if v, ok := value.(bool); ok {
				sc.config.MinimizeToTray = v
			}
		}
	}
}

func (sc *SettingsComponent) GetConfig() *UIConfig {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.config
}

func (sc *SettingsComponent) Save() error {
	if sc.onSave != nil {
		sc.onSave(sc.config)
	}
	return nil
}

func (sc *SettingsComponent) SetOnChange(callback func(section, key string, value interface{})) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.onChange = callback
}

func (sc *SettingsComponent) SetOnSave(callback func(config *UIConfig)) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.onSave = callback
}

func (sc *SettingsComponent) SetOnReset(callback func()) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.onReset = callback
}

func (sc *SettingsComponent) ImportConfig(config *UIConfig) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.config = config

	sc.sections["general"].Settings["theme"].Value = string(config.Theme)
	sc.sections["general"].Settings["language"].Value = config.Language
	sc.sections["general"].Settings["font_size"].Value = config.FontSize

	sc.sections["chat"].Settings["show_avatars"].Value = config.ShowAvatars
	sc.sections["chat"].Settings["show_timestamps"].Value = config.ShowTimestamps

	sc.sections["notification"].Settings["notify_sounds"].Value = config.NotifySounds
	sc.sections["notification"].Settings["notify_desktop"].Value = config.NotifyDesktop

	sc.sections["system"].Settings["auto_start"].Value = config.AutoStart
	sc.sections["system"].Settings["minimize_to_tray"].Value = config.MinimizeToTray

	return nil
}
