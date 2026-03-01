package i18n

import (
	"encoding/json"
	"fmt"
	"sync"
)

var (
	defaultLocale = "en"
	translations  map[string]map[string]string
	mu            sync.RWMutex
	currentLocale string
	initialized   bool
)

func init() {
	translations = make(map[string]map[string]string)
	LoadTranslations()
}

func LoadTranslations() {
	mu.Lock()
	defer mu.Unlock()

	translations["en"] = map[string]string{
		"app_name":           "O2OChat",
		"settings":           "Settings",
		"add_contact":        "Add Contact",
		"enter_peer_id":      "Enter Peer ID",
		"enter_display_name": "Enter Display Name",
		"send":               "Send",
		"cancel":             "Cancel",
		"ok":                 "OK",
		"message_hint":       "Type a message...",
		"voice_call":         "Voice Call",
		"video_call":         "Video Call",
		"send_file":          "Send File",
		"no_contacts":        "No contacts yet",
		"online":             "Online",
		"offline":            "Offline",
		"connecting":         "Connecting...",
		"call_ended":         "Call Ended",
		"incoming_call":      "Incoming Call",
		"outgoing_call":      "Calling...",
		"general":            "General",
		"chat":               "Chat",
		"notifications":      "Notifications",
		"security":           "Security",
		"about":              "About",
		"language":           "Language",
		"theme":              "Theme",
		"dark_mode":          "Dark Mode",
		"light_mode":         "Light Mode",
		"auto":               "Auto",
		"save":               "Save",
		"delete":             "Delete",
		"edit":               "Edit",
		"confirm":            "Confirm",
		"error":              "Error",
		"success":            "Success",
		"loading":            "Loading...",
		"no_messages":        "No messages yet",
		"type_message":       "Type a message...",
		"search_contacts":    "Search contacts...",
		"new_message":        "New Message",
		"create_group":       "Create Group",
		"group_name":         "Group Name",
		"add_members":        "Add Members",
		"remove_member":      "Remove Member",
		"leave_group":        "Leave Group",
		"delete_group":       "Delete Group",
		"group_info":         "Group Info",
		"member_count":       "%d members",
		"today":              "Today",
		"yesterday":          "Yesterday",
		"typing":             "typing...",
		"seen":               "Seen",
		"delivered":          "Delivered",
		"sent":               "Sent",
		"failed":             "Failed",
		"retry":              "Retry",
		"copy":               "Copy",
		"forward":            "Forward",
		"reply":              "Reply",
		"delete_message":     "Delete Message",
		"call_in_progress":   "Call in progress",
		"call_duration":      "Duration: %s",
		"mute":               "Mute",
		"unmute":             "Unmute",
		"enable_video":       "Enable Video",
		"disable_video":      "Disable Video",
		"share_screen":       "Share Screen",
		"stop_sharing":       "Stop Sharing",
		"end_call":           "End Call",
		"accept":             "Accept",
		"decline":            "Decline",
		"missed_call":        "Missed Call",
		"peer_id":            "Peer ID",
		"copy_peer_id":       "Copy Peer ID",
		"share_peer_id":      "Share Peer ID",
		"my_peer_id":         "My Peer ID",
		"scan_qr":            "Scan QR Code",
		"show_qr":            "Show QR Code",
		"enter_password":     "Enter Password",
		"set_password":       "Set Password",
		"change_password":    "Change Password",
		"remove_password":    "Remove Password",
		"encryption":         "Encryption",
		"end_to_end":         "End-to-End Encrypted",
		"verify_device":      "Verify Device",
		"version":            "Version",
		"check_updates":      "Check for Updates",
		"no_updates":         "You are up to date",
		"update_available":   "Update Available",
		"download_update":    "Download Update",
		"install_update":     "Install Update",
		"restart_now":        "Restart Now",
		"connection_lost":    "Connection Lost",
		"reconnecting":       "Reconnecting...",
		"connected":          "Connected",
		"disconnected":       "Disconnected",
		"welcome":            "Welcome to O2OChat",
		"get_started":        "Get Started",
		"privacy_policy":     "Privacy Policy",
		"terms_of_service":   "Terms of Service",
		"accept_terms":       "I accept the Terms of Service",
	}

	translations["zh"] = map[string]string{
		"app_name":           "O2OChat",
		"settings":           "设置",
		"add_contact":        "添加联系人",
		"enter_peer_id":      "输入 Peer ID",
		"enter_display_name": "输入显示名称",
		"send":               "发送",
		"cancel":             "取消",
		"ok":                 "确定",
		"message_hint":       "输入消息...",
		"voice_call":         "语音通话",
		"video_call":         "视频通话",
		"send_file":          "发送文件",
		"no_contacts":        "暂无联系人",
		"online":             "在线",
		"offline":            "离线",
		"connecting":         "连接中...",
		"call_ended":         "通话结束",
		"incoming_call":      "来电",
		"outgoing_call":      "呼叫中...",
		"general":            "通用",
		"chat":               "聊天",
		"notifications":      "通知",
		"security":           "安全",
		"about":              "关于",
		"language":           "语言",
		"theme":              "主题",
		"dark_mode":          "深色模式",
		"light_mode":         "浅色模式",
		"auto":               "自动",
		"save":               "保存",
		"delete":             "删除",
		"edit":               "编辑",
		"confirm":            "确认",
		"error":              "错误",
		"success":            "成功",
		"loading":            "加载中...",
		"no_messages":        "暂无消息",
		"type_message":       "输入消息...",
		"search_contacts":    "搜索联系人...",
		"new_message":        "新消息",
		"create_group":       "创建群组",
		"group_name":         "群名称",
		"add_members":        "添加成员",
		"remove_member":      "移除成员",
		"leave_group":        "退出群组",
		"delete_group":       "删除群组",
		"group_info":         "群信息",
		"member_count":       "%d 位成员",
		"today":              "今天",
		"yesterday":          "昨天",
		"typing":             "正在输入...",
		"seen":               "已读",
		"delivered":          "已送达",
		"sent":               "已发送",
		"failed":             "发送失败",
		"retry":              "重试",
		"copy":               "复制",
		"forward":            "转发",
		"reply":              "回复",
		"delete_message":     "删除消息",
		"call_in_progress":   "通话中",
		"call_duration":      "时长: %s",
		"mute":               "静音",
		"unmute":             "取消静音",
		"enable_video":       "开启视频",
		"disable_video":      "关闭视频",
		"share_screen":       "共享屏幕",
		"stop_sharing":       "停止共享",
		"end_call":           "结束通话",
		"accept":             "接听",
		"decline":            "拒绝",
		"missed_call":        "未接来电",
		"peer_id":            "Peer ID",
		"copy_peer_id":       "复制 Peer ID",
		"share_peer_id":      "分享 Peer ID",
		"my_peer_id":         "我的 Peer ID",
		"scan_qr":            "扫描二维码",
		"show_qr":            "显示二维码",
		"enter_password":     "输入密码",
		"set_password":       "设置密码",
		"change_password":    "修改密码",
		"remove_password":    "移除密码",
		"encryption":         "加密",
		"end_to_end":         "端到端加密",
		"verify_device":      "验证设备",
		"version":            "版本",
		"check_updates":      "检查更新",
		"no_updates":         "已是最新版本",
		"update_available":   "有可用更新",
		"download_update":    "下载更新",
		"install_update":     "安装更新",
		"restart_now":        "立即重启",
		"connection_lost":    "连接已断开",
		"reconnecting":       "重新连接中...",
		"connected":          "已连接",
		"disconnected":       "已断开",
		"welcome":            "欢迎使用 O2OChat",
		"get_started":        "开始使用",
		"privacy_policy":     "隐私政策",
		"terms_of_service":   "服务条款",
		"accept_terms":       "我同意服务条款",
	}

	translations["zh-CN"] = translations["zh"]
	translations["en-US"] = translations["en"]
	translations["zh-TW"] = map[string]string{
		"app_name":       "O2OChat",
		"settings":       "設定",
		"add_contact":    "新增聯絡人",
		"enter_peer_id":  "輸入 Peer ID",
		"enter_nickname": "輸入暱稱",
		"send":           "傳送",
		"cancel":         "取消",
		"ok":             "確定",
		"message_hint":   "輸入訊息...",
		"voice_call":     "語音通話",
		"video_call":     "視訊通話",
		"send_file":      "傳送檔案",
		"no_contacts":    "尚無聯絡人",
		"online":         "線上",
		"offline":        "離線",
		"connecting":     "連線中...",
		"call_ended":     "通話結束",
		"incoming_call":  "來電",
		"outgoing_call":  "呼叫中...",
	}
	translations["bo"] = map[string]string{
		"app_name":       "O2OChat",
		"settings":       "སྒ໲ད་སྤྱོད",
		"add_contact":    "འབྲེལ་བ་ཁ་བསྡུ་བ",
		"enter_peer_id":  "Peer ID བསྒ໲ད་བྱོས",
		"enter_nickname": "མིང་འགྱུར་བསྒ໲ད་བྱོས",
		"send":           "བསྲུང་བ",
		"cancel":         "འདོེར་བ",
		"ok":             "ངེས་པ",
		"message_hint":   "འཕྲིན་བསྒ໲ད་བྱོས...",
		"voice_call":     "སྤྱི་ནོར་ཁ་བསྡུ་བ",
		"video_call":     "བརྙན་འཕྲིན་ཁ་བསྡུ་བ",
		"send_file":      "ཡིག་ཆ་བསྲུང་བ",
		"no_contacts":    "འབྲེལ་བ་མེད་པ",
		"online":         "ཡོད་པ",
		"offline":        "མེད་པ",
		"connecting":     "འབྲེལ་བར་བྱེད་པ",
		"call_ended":     "ཁ་བསྡུ་མཚམས་འདུག",
		"incoming_call":  "ཁ་བྱུང་བ",
		"outgoing_call":  "ཁ་འབོད་པ",
	}
	translations["mn"] = map[string]string{
		"app_name":       "O2OChat",
		"settings":       "Тохиргоо",
		"add_contact":    "Холбоо нэмэх",
		"enter_peer_id":  "Peer ID оруулна уу",
		"enter_nickname": "Нэрээ оруулна уу",
		"send":           "Илгээх",
		"cancel":         "Цуцлах",
		"ok":             "Тийм",
		"message_hint":   "Зурвас бичнэ үү...",
		"voice_call":     "Дууны яриа",
		"video_call":     "Видео яриа",
		"send_file":      "Файл илгээх",
		"no_contacts":    "Холбоо байхгүй",
		"online":         "Онлайн",
		"offline":        "Офлайн",
		"connecting":     "Холбогдож байна...",
		"call_ended":     "Яриа дууслаа",
		"incoming_call":  "Ирж буй дуудлага",
		"outgoing_call":  "Ярьж байна...",
	}
	translations["ug"] = map[string]string{
		"app_name":       "O2OChat",
		"settings":       "تەڭشەك",
		"add_contact":    "ئالاقىداشنى قوشۇش",
		"enter_peer_id":  "Peer ID نى كىرگۈزۈڭ",
		"enter_nickname": "ئىسمىنى كىرگۈزۈڭ",
		"send":           "يوللاش",
		"cancel":         "بىكار قىلىش",
		"ok":             "ئىناۋەتلىك",
		"message_hint":   "خەت يېزىڭ...",
		"voice_call":     "ئاۋازلىق سۆزلىشىش",
		"video_call":     "ۋىدېيولۇق سۆزلىشىش",
		"send_file":      "ھۆجەت يوللاش",
		"no_contacts":    "ئالاقىداش يوق",
		"online":         "ئىشلەپ تۇرۇۋاتىدۇ",
		"offline":        "ئىشلەپ تۇرمايدۇ",
		"connecting":     "ئۇلانماقتا...",
		"call_ended":     "سۆزلىشىش ئاخىرلاشتى",
		"incoming_call":  "كىرىۋاتقان چاقىرىش",
		"outgoing_call":  "چاقىرىۋاتىدۇ...",
	}
	translations["de"] = map[string]string{
		"app_name":       "O2OChat",
		"settings":       "Einstellungen",
		"add_contact":    "Kontakt hinzufügen",
		"enter_peer_id":  "Peer-ID eingeben",
		"enter_nickname": "Anzeigename eingeben",
		"send":           "Senden",
		"cancel":         "Abbrechen",
		"ok":             "OK",
		"message_hint":   "Nachricht eingeben...",
		"voice_call":     "Sprachanruf",
		"video_call":     "Videoanruf",
		"send_file":      "Datei senden",
		"no_contacts":    "Keine Kontakte",
		"online":         "Online",
		"offline":        "Offline",
		"connecting":     "Verbinden...",
		"call_ended":     "Anruf beendet",
		"incoming_call":  "Eingehender Anruf",
		"outgoing_call":  "Anrufen...",
	}
	translations["fr"] = map[string]string{
		"app_name":       "O2OChat",
		"settings":       "Paramètres",
		"add_contact":    "Ajouter un contact",
		"enter_peer_id":  "Entrer l'ID Peer",
		"enter_nickname": "Entrer le nom",
		"send":           "Envoyer",
		"cancel":         "Annuler",
		"ok":             "OK",
		"message_hint":   "Tapez un message...",
		"voice_call":     "Appel vocal",
		"video_call":     "Appel vidéo",
		"send_file":      "Envoyer un fichier",
		"no_contacts":    "Aucun contact",
		"online":         "En ligne",
		"offline":        "Hors ligne",
		"connecting":     "Connexion...",
		"call_ended":     "Appel terminé",
		"incoming_call":  "Appel entrant",
		"outgoing_call":  "Appel en cours...",
	}
	translations["es"] = map[string]string{
		"app_name":       "O2OChat",
		"settings":       "Ajustes",
		"add_contact":    "Agregar contacto",
		"enter_peer_id":  "Ingrese ID de Peer",
		"enter_nickname": "Ingrese nombre",
		"send":           "Enviar",
		"cancel":         "Cancelar",
		"ok":             "Aceptar",
		"message_hint":   "Escribe un mensaje...",
		"voice_call":     "Llamada de voz",
		"video_call":     "Videollamada",
		"send_file":      "Enviar archivo",
		"no_contacts":    "Sin contactos",
		"online":         "En línea",
		"offline":        "Desconectado",
		"connecting":     "Conectando...",
		"call_ended":     "Llamada finalizada",
		"incoming_call":  "Llamada entrante",
		"outgoing_call":  "Llamando...",
	}
	translations["ja"] = map[string]string{
		"app_name":       "O2OChat",
		"settings":       "設定",
		"add_contact":    "連絡先を追加",
		"enter_peer_id":  "Peer IDを入力",
		"enter_nickname": "名前を入力",
		"send":           "送信",
		"cancel":         "キャンセル",
		"ok":             "OK",
		"message_hint":   "メッセージを入力...",
		"voice_call":     "音声通話",
		"video_call":     "ビデオ通話",
		"send_file":      "ファイルを送信",
		"no_contacts":    "連絡先がありません",
		"online":         "オンライン",
		"offline":        "オフライン",
		"connecting":     "接続中...",
		"call_ended":     "通話終了",
		"incoming_call":  "着信",
		"outgoing_call":  "発信中...",
	}
	translations["ko"] = map[string]string{
		"app_name":       "O2OChat",
		"settings":       "설정",
		"add_contact":    "연락처 추가",
		"enter_peer_id":  "Peer ID 입력",
		"enter_nickname": "이름 입력",
		"send":           "보내기",
		"cancel":         "취소",
		"ok":             "확인",
		"message_hint":   "메시지 입력...",
		"voice_call":     "음성 통화",
		"video_call":     "영상 통화",
		"send_file":      "파일 보내기",
		"no_contacts":    "연락처 없음",
		"online":         "온라인",
		"offline":        "오프라인",
		"connecting":     "연결 중...",
		"call_ended":     "통화 종료",
		"incoming_call":  "수신 전화",
		"outgoing_call":  "통화 중...",
	}
	translations["ru"] = map[string]string{
		"app_name":       "O2OChat",
		"settings":       "Настройки",
		"add_contact":    "Добавить контакт",
		"enter_peer_id":  "Введите Peer ID",
		"enter_nickname": "Введите имя",
		"send":           "Отправить",
		"cancel":         "Отмена",
		"ok":             "ОК",
		"message_hint":   "Введите сообщение...",
		"voice_call":     "Голосовой вызов",
		"video_call":     "Видеозвонок",
		"send_file":      "Отправить файл",
		"no_contacts":    "Нет контактов",
		"online":         "Онлайн",
		"offline":        "Офлайн",
		"connecting":     "Подключение...",
		"call_ended":     "Вызов завершён",
		"incoming_call":  "Входящий вызов",
		"outgoing_call":  "Вызов...",
	}
	translations["ms"] = map[string]string{
		"app_name":       "O2OChat",
		"settings":       "Tetapan",
		"add_contact":    "Tambah kenalan",
		"enter_peer_id":  "Masukkan ID Peer",
		"enter_nickname": "Masukkan nama",
		"send":           "Hantar",
		"cancel":         "Batal",
		"ok":             "OK",
		"message_hint":   "Taip mesej...",
		"voice_call":     "Panggilan suara",
		"video_call":     "Panggilan video",
		"send_file":      "Hantar fail",
		"no_contacts":    "Tiada kenalan",
		"online":         "Online",
		"offline":        "Offline",
		"connecting":     "Menyambu...",
		"call_ended":     "Panggilan tamat",
		"incoming_call":  "Panggilan masuk",
		"outgoing_call":  "Memanggil...",
	}
	translations["he"] = map[string]string{
		"app_name":       "O2OChat",
		"settings":       "הגדרות",
		"add_contact":    "הוסף קשר",
		"enter_peer_id":  "הזן Peer ID",
		"enter_nickname": "הזן שם",
		"send":           "שלח",
		"cancel":         "ביטול",
		"ok":             "אישור",
		"message_hint":   "הקלד הודעה...",
		"voice_call":     "שיחה קולית",
		"video_call":     "שיחת וידאו",
		"send_file":      "שלח קובץ",
		"no_contacts":    "אין קשרים",
		"online":         "מחובר",
		"offline":        "מנותק",
		"connecting":     "מתחבר...",
		"call_ended":     "השיחה הסתיימה",
		"incoming_call":  "שיחה נכנסת",
		"outgoing_call":  "מתקשר...",
	}
	translations["ar"] = map[string]string{
		"app_name":       "O2OChat",
		"settings":       "الإعدادات",
		"add_contact":    "إضافة جهة اتصال",
		"enter_peer_id":  "أدخل Peer ID",
		"enter_nickname": "أدخل الاسم",
		"send":           "إرسال",
		"cancel":         "إلغاء",
		"ok":             "موافق",
		"message_hint":   "اكتب رسالة...",
		"voice_call":     "مكالمة صوتية",
		"video_call":     "مكلة فيديو",
		"send_file":      "إرسال ملف",
		"no_contacts":    "لا توجد جهات اتصال",
		"online":         "متصل",
		"offline":        "غير متصل",
		"connecting":     "جاري الاتصال...",
		"call_ended":     "انتهت المكالمة",
		"incoming_call":  "مكالمة واردة",
		"outgoing_call":  "جاري الاتصال...",
	}

	initialized = true
}

func SetLocale(locale string) {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := translations[locale]; ok {
		currentLocale = locale
	} else if len(locale) >= 2 {
		lang := locale[:2]
		if _, ok := translations[lang]; ok {
			currentLocale = lang
		} else {
			currentLocale = defaultLocale
		}
	} else {
		currentLocale = defaultLocale
	}
}

func GetLocale() string {
	mu.RLock()
	defer mu.RUnlock()
	if currentLocale == "" {
		return defaultLocale
	}
	return currentLocale
}

func T(key string) string {
	mu.RLock()
	defer mu.RUnlock()

	locale := currentLocale
	if locale == "" {
		locale = defaultLocale
	}

	if t, ok := translations[locale]; ok {
		if val, ok := t[key]; ok {
			return val
		}
	}

	if t, ok := translations[defaultLocale]; ok {
		if val, ok := t[key]; ok {
			return val
		}
	}

	return key
}

func TF(key string, args ...interface{}) string {
	return fmt.Sprintf(T(key), args...)
}

func GetAvailableLocales() []string {
	mu.RLock()
	defer mu.RUnlock()

	locales := make([]string, 0, len(translations))
	for locale := range translations {
		locales = append(locales, locale)
	}
	return locales
}

func AddTranslation(locale, key, value string) {
	mu.Lock()
	defer mu.Unlock()

	if translations[locale] == nil {
		translations[locale] = make(map[string]string)
	}
	translations[locale][key] = value
}

func ExportTranslations() map[string]map[string]string {
	mu.RLock()
	defer mu.RUnlock()

	result := make(map[string]map[string]string)
	for locale, trans := range translations {
		result[locale] = make(map[string]string)
		for k, v := range trans {
			result[locale][k] = v
		}
	}
	return result
}

func ImportTranslations(data map[string]map[string]string) {
	mu.Lock()
	defer mu.Unlock()

	for locale, trans := range data {
		if translations[locale] == nil {
			translations[locale] = make(map[string]string)
		}
		for k, v := range trans {
			translations[locale][k] = v
		}
	}
}

func MarshalJSON() ([]byte, error) {
	return json.Marshal(ExportTranslations())
}

func UnmarshalJSON(data []byte) error {
	var trans map[string]map[string]string
	if err := json.Unmarshal(data, &trans); err != nil {
		return err
	}
	ImportTranslations(trans)
	return nil
}
