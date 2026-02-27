# O2OChat

🌍 **[English](README_EN.md)** | **[中文](README.md)** | **[Español](README_ES.md)** | **[Français](README_FR.md)** | **[Deutsch](README_DE.md)** | **[日本語](README_JA.md)** | **[한국어](README_KO.md)** | **[Русский](README_RU.md)** | **[العربية](README_AR.md)** | **[עברית](README_HE.md)** | **[Bahasa Melayu](README_MS.md)**

## תוכנת מסרים מיידיים P2P טהורה

O2OChat היא תוכנת מסרים מיידיים עמית-לעמית (P2P) טהורה שאינה תלויה בשרתים מרכזיים לאחסון הודעות. כל התקשורת מתבצעת ישירות בין המשתמשים.

### תכונות ליבה

- 🔒 **הצפנה מקצה לקצה** - כל ההודעות משתמשות בהצפנה AES-256-GCM
- 🌐 **ארכיטקטורת P2P טהורה** - אין שרת מרכזי, תקשורת ישירה
- 📱 **תמיכה במספר פלטפורמות** - Android, iOS, Windows, Linux, macOS, HarmonyOS
- 📁 **העברת קבצים** - המשך העברה מופסקת, הורדה ממקורות מרובים, אימות עץ מרקל
- 🌍 **16 שפות** - סינית, אנגלית, יפנית, קוריאנית, גרמנית, צרפתית, ספרדית, רוסית, מלאית, עברית, ערבית, טיבטית, מונגולית, אויגורית, סינית (מסורתית)

### התחלה מהירה

```bash
# שכפול הפרויקט
git clone https://github.com/yourusername/o2ochat.git
cd o2ochat

# בנייה
go build -o o2ochat ./cmd/o2ochat

# הפעלה
./o2ochat
```

### מבנה הפרויקט

```
o2ochat/
├── cmd/              # נקודות כניסה
├── pkg/              # ספריות ליבה
│   ├── identity/     # ניהול זהויות
│   ├── transport/    # תחבורת רשת
│   ├── signaling/    # שירות איתות
│   ├── crypto/       # מודול הצפנה
│   ├── storage/      # אחסון נתונים
│   ├── filetransfer/ # העברת קבצים
│   └── media/        # עיבוד אודיו/וידאו
├── ui/               # ממשק משתמש
├── cli/              # כלי שורת פקודה
├── tests/            # בדיקות
├── docs/             # תיעוד
└── scripts/          # סקריפטים לבנייה
```

### מחסנית טכנולוגיות

- **Go 1.21+** - ליבת בק-אנד
- **Protocol Buffers** - סידור
- **QUIC/WebRTC** - תחבורת P2P
- **SQLite** - אחסון מקומי
- **Fyne** - GUI לשולחן עבודה
- **Jetpack Compose** - UI ל-Android
- **SwiftUI** - UI ל-iOS
- **ArkTS** - UI ל-HarmonyOS

### תרומה

תרומות יתקבלו בברכה! אנא קרא את [מדריך התרומה](CONTRIBUTING.md).

### רישיון

רישיון MIT - פרטים נוספים בקובץ [LICENSE](LICENSE).

### צור קשר

- דף הבית של הפרויקט: https://o2ochat.io
- מעקב בעיות: https://github.com/yourusername/o2ochat/issues
- דוא"ל: support@o2ochat.io

---

<p align="center">
  <b>P2P טהור | הצפנה מקצה לקצה | תקשורת חופשית</b>
</p>
