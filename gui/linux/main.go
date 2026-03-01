package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// O2OChatLinuxGUI represents the Linux GUI application
type O2OChatLinuxGUI struct {
	app       fyne.App
	window    fyne.MainWindow
	contacts  []string
	messages  []*widget.RichText
	statusBar *widget.Label
}

// NewO2OChatLinuxGUI creates a new Linux GUI
func NewO2OChatLinuxGUI() *O2OChatLinuxGUI {
	return &O2OChatLinuxGUI{
		contacts: []string{"Contact 1", "Contact 2", "Contact 3"},
		messages: make([]*widget.RichText, 0),
	}
}

// Initialize initializes the application
func (gui *O2OChatLinuxGUI) Initialize() {
	gui.app = app.New()
	gui.window = gui.app.NewWindow("O2OChat v3.0.0-beta - Linux")
	gui.window.Resize(fyne.NewSize(800, 600))
}

// createMenuBar creates the menu bar
func (gui *O2OChatLinuxGUI) createMenuBar() *fyne.MainMenu {
	newChat := fyne.NewMenuItem("New Chat", func() {
		gui.statusBar.SetText("Creating new chat...")
	})
	
	settings := fyne.NewMenuItem("Settings", func() {
		gui.showSettings()
	})
	
	about := fyne.NewMenuItem("About", func() {
		dialog.ShowInformation("About", "O2OChat v3.0.0-beta\nLinux Desktop Application", gui.window)
	})
	
	fileMenu := fyne.NewMenu("File", newChat, settings)
	helpMenu := fyne.NewMenu("Help", about)
	
	return fyne.NewMainMenu(fileMenu, helpMenu)
}

// createToolbar creates the toolbar
func (gui *O2OChatLinuxGUI) createToolbar() *widget.Toolbar {
	return widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			gui.showAddContact()
		}),
		widget.NewToolbarAction(theme.MailComposeIcon(), func() {
			gui.statusBar.SetText("New message...")
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.SettingsIcon(), func() {
			gui.showSettings()
		}),
		widget.NewToolbarAction(theme.HelpIcon(), func() {
			dialog.ShowInformation("Help", "O2OChat Help\n\nType a message and press Send", gui.window)
		}),
	)
}

// createContactList creates the contact list sidebar
func (gui *O2OChatLinuxGUI) createContactList() fyne.CanvasObject {
	contactList := widget.NewList(
		func() int {
			return len(gui.contacts)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template Contact")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(gui.contacts[id])
		},
	)
	
	contactList.OnSelected = func(id widget.ListItemID) {
		gui.statusBar.SetText("Selected: " + gui.contacts[id])
	}
	
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search contacts...")
	searchEntry.OnChanged = func(text string) {
		contactList.Refresh()
	}
	
	addBtn := widget.NewButton("+ Add Contact", func() {
		gui.showAddContact()
	})
	
	return container.NewVBox(
		widget.NewLabelWithStyle("Contacts", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		searchEntry,
		contactList,
		addBtn,
	)
}

// createChatArea creates the main chat area
func (gui *O2OChatLinuxGUI) createChatArea() fyne.CanvasObject {
	// Messages area
	messagesBox := container.NewVBox()
	
	messageScroll := container.NewScroll(messagesBox)
	
	// Message input
	messageEntry := widget.NewMultiLineEntry()
	messageEntry.SetPlaceHolder("Type your message...")
	messageEntry.SetMinRowsVisible(3)
	
	sendBtn := widget.NewButtonWithIcon("Send", theme.MailSendIcon(), func() {
		if messageEntry.Text != "" {
			// Add message
			messageLabel := widget.NewLabel("You: " + messageEntry.Text)
			messagesBox.Add(messageLabel)
			messageEntry.SetText("")
			gui.statusBar.SetText("Message sent!")
			
			// Scroll to bottom
			messageScroll.ScrollToBottom()
		}
	})
	
	attachBtn := widget.NewButtonWithIcon("Attach", theme.FileIcon(), func() {
		gui.statusBar.SetText("Attach file...")
	})
	
	micBtn := widget.NewButtonWithIcon("Voice", theme.MediaMusicIcon(), func() {
		gui.statusBar.SetText("Voice message...")
	})
	
	toolbar := container.NewHBox(
		attachBtn,
		micBtn,
		widget.NewSpacer(),
		sendBtn,
	)
	
	return container.NewBorder(nil, toolbar, nil, nil, messageScroll)
}

// showSettings shows settings dialog
func (gui *O2OChatLinuxGUI) showSettings() {
	usernameEntry := widget.NewEntry()
	usernameEntry.SetText("User")
	usernameEntry.SetPlaceHolder("Username")
	
	themeSelect := widget.NewSelect([]string{"Light", "Dark", "System"}, func(s string) {
		gui.statusBar.SetText("Theme: " + s)
	})
	themeSelect.SetSelected("System")
	
	notificationsCheck := widget.NewCheck("Enable notifications", func(checked bool) {
		if checked {
			gui.statusBar.SetText("Notifications enabled")
		} else {
			gui.statusBar.SetText("Notifications disabled")
		}
	})
	notificationsCheck.SetChecked(true)
	
	form := dialog.NewForm("Settings", "Save", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Username", usernameEntry),
			widget.NewFormItem("Theme", themeSelect),
			widget.NewFormItem("Notifications", notificationsCheck),
		},
		func(ok bool) {
			if ok {
				gui.statusBar.SetText("Settings saved!")
			}
		},
		gui.window,
	)
	
	form.Resize(fyne.NewSize(400, 300))
	form.Show()
}

// showAddContact shows add contact dialog
func (gui *O2OChatLinuxGUI) showAddContact() {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Contact name")
	
	idEntry := widget.NewEntry()
	idEntry.SetPlaceHolder("Peer ID")
	
	form := dialog.NewForm("Add Contact", "Add", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Name", nameEntry),
			widget.NewFormItem("Peer ID", idEntry),
		},
		func(ok bool) {
			if ok && nameEntry.Text != "" && idEntry.Text != "" {
				gui.contacts = append(gui.contacts, nameEntry.Text)
				gui.statusBar.SetText("Contact added: " + nameEntry.Text)
			}
		},
		gui.window,
	)
	
	form.Resize(fyne.NewSize(400, 250))
	form.Show()
}

// createStatusBar creates the status bar
func (gui *O2OChatLinuxGUI) createStatusBar() fyne.CanvasObject {
	gui.statusBar = widget.NewLabel("Ready - v3.0.0-beta")
	gui.statusBar.TextStyle = fyne.TextStyle{Italic: true}
	
	versionLabel := widget.NewLabel("v3.0.0-beta")
	versionLabel.TextStyle = fyne.TextStyle{Italic: true}
	
	return container.NewHBox(gui.statusBar, widget.NewSpacer(), versionLabel)
}

// BuildUI builds the complete UI
func (gui *O2OChatLinuxGUI) BuildUI() fyne.CanvasObject {
	// Create components
	contactList := gui.createContactList()
	chatArea := gui.createChatArea()
	statusBar := gui.createStatusBar()
	
	// Create main layout
	split := container.NewHSplit(contactList, chatArea)
	split.SetOffset(0.3)
	
	content := container.NewBorder(
		gui.createToolbar(), // Top
		statusBar,           // Bottom
		nil,                 // Left
		nil,                 // Right
		split,               // Center
	)
	
	return content
}

// Run runs the application
func (gui *O2OChatLinuxGUI) Run() {
	gui.Initialize()
	gui.window.SetMainMenu(gui.createMenuBar())
	gui.window.SetContent(gui.BuildUI())
	gui.window.ShowAndRun()
}

// Main function for Linux GUI
func main() {
	gui := NewO2OChatLinuxGUI()
	gui.Run()
}
