package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// O2OChatGUI represents the main GUI window
type O2OChatGUI struct {
	app       fyne.App
	window    fyne.Window
	statusBar *widget.Label
}

// NewO2OChatGUI creates a new O2OChat GUI
func NewO2OChatGUI() *O2OChatGUI {
	return &O2OChatGUI{}
}

// Initialize initializes the GUI
func (gui *O2OChatGUI) Initialize() {
	gui.app = app.New()
	gui.window = gui.app.NewWindow("O2OChat v3.0.0-beta")
	
	// Set minimum window size
	gui.window.Resize(fyne.NewSize(800, 600))
	gui.window.SetFixedSize(false)
}

// BuildUI builds the main UI
func (gui *O2OChatGUI) BuildUI() fyne.CanvasObject {
	// Create header
	header := gui.createHeader()
	
	// Create sidebar
	sidebar := gui.createSidebar()
	
	// Create main content
	mainContent := gui.createMainContent()
	
	// Create status bar
	gui.statusBar = widget.NewLabel("Ready - v3.0.0-beta")
	gui.statusBar.TextStyle = widget.TextStyle{Italic: true}
	
	// Create main layout
	content := container.NewBorder(
		header,      // Top
		gui.statusBar, // Bottom
		sidebar,     // Left
		nil,         // Right
		mainContent, // Center
	)
	
	return content
}

// createHeader creates the header bar
func (gui *O2OChatGUI) createHeader() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("O2OChat v3.0.0-beta", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	
	newChatBtn := widget.NewButton("New Chat", func() {
		gui.statusBar.SetText("Creating new chat...")
	})
	
	settingsBtn := widget.NewButton("Settings", func() {
		gui.showSettings()
	})
	
	header := container.NewHBox(
		title,
		widget.NewSeparator(),
		widget.NewSpacer(),
		newChatBtn,
		settingsBtn,
	)
	
	return header
}

// createSidebar creates the sidebar with contacts
func (gui *O2OChatGUI) createSidebar() fyne.CanvasObject {
	contacts := []string{
		"Contact 1",
		"Contact 2",
		"Contact 3",
		"Contact 4",
		"Contact 5",
	}
	
	contactList := widget.NewList(
		func() int {
			return len(contacts)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(contacts[id])
		},
	)
	
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search contacts...")
	
	addBtn := widget.NewButton("+ Add", func() {
		gui.showAddContact()
	})
	
	sidebar := container.NewVBox(
		widget.NewLabelWithStyle("Contacts", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		searchEntry,
		contactList,
		addBtn,
	)
	
	return sidebar
}

// createMainContent creates the main content area
func (gui *O2OChatGUI) createMainContent() fyne.CanvasObject {
	// Create message list
	messages := widget.NewVBox(
		widget.NewLabel("Message 1: Hello!"),
		widget.NewLabel("Message 2: Hi there!"),
		widget.NewLabel("Message 3: How are you?"),
	)
	
	scrollContainer := container.NewScroll(messages)
	
	// Create message input
	messageEntry := widget.NewMultiLineEntry()
	messageEntry.SetPlaceHolder("Type your message...")
	messageEntry.SetMinRowsVisible(3)
	
	sendBtn := widget.NewButton("Send", func() {
		if messageEntry.Text != "" {
			// Add message to list
			messages.Append(widget.NewLabel("You: " + messageEntry.Text))
			messageEntry.SetText("")
			gui.statusBar.SetText("Message sent!")
			
			// Scroll to bottom
			scrollContainer.Offset.Y = 9999
			scrollContainer.Refresh()
		}
	})
	
	// Create toolbar
	toolbar := container.NewHBox(
		widget.NewButton("📁", func() {
			gui.statusBar.SetText("Attach file...")
		}),
		widget.NewButton("🎤", func() {
			gui.statusBar.SetText("Voice message...")
		}),
		widget.NewButton("📷", func() {
			gui.statusBar.SetText("Send image...")
		}),
		widget.NewSpacer(),
		sendBtn,
	)
	
	content := container.NewBorder(
		nil,       // Top
		toolbar,   // Bottom
		nil,       // Left
		nil,       // Right
		scrollContainer,
	)
	
	return content
}

// showSettings shows settings window
func (gui *O2OChatGUI) showSettings() {
	settingsWindow := gui.app.NewWindow("Settings")
	
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username")
	usernameEntry.SetText("User")
	
	themeSelect := widget.NewSelect(
		[]string{"Light", "Dark", "System"},
		func(s string) {
			gui.statusBar.SetText("Theme changed to: " + s)
		},
	)
	themeSelect.SetSelected("System")
	
	notificationsCheck := widget.NewCheck("Enable notifications", func(checked bool) {
		gui.statusBar.SetText("Notifications: " + map[bool]string{true: "Enabled", false: "Disabled"}[checked])
	})
	notificationsCheck.SetChecked(true)
	
	saveBtn := widget.NewButton("Save", func() {
		gui.statusBar.SetText("Settings saved!")
		settingsWindow.Close()
	})
	
	cancelBtn := widget.NewButton("Cancel", func() {
		settingsWindow.Close()
	})
	
	content := container.NewVBox(
		widget.NewLabelWithStyle("Settings", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel("Username:"),
		usernameEntry,
		widget.NewLabel("Theme:"),
		themeSelect,
		widget.NewSeparator(),
		notificationsCheck,
		widget.NewSeparator(),
		container.NewHBox(saveBtn, cancelBtn),
	)
	
	settingsWindow.SetContent(content)
	settingsWindow.Resize(fyne.NewSize(400, 300))
	settingsWindow.Show()
}

// showAddContact shows add contact dialog
func (gui *O2OChatGUI) showAddContact() {
	addWindow := gui.app.NewWindow("Add Contact")
	
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Contact name")
	
	idEntry := widget.NewEntry()
	idEntry.SetPlaceHolder("Peer ID")
	
	addBtn := widget.NewButton("Add", func() {
		if nameEntry.Text != "" && idEntry.Text != "" {
			gui.statusBar.SetText("Contact added: " + nameEntry.Text)
			addWindow.Close()
		}
	})
	
	cancelBtn := widget.NewButton("Cancel", func() {
		addWindow.Close()
	})
	
	content := container.NewVBox(
		widget.NewLabelWithStyle("Add Contact", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel("Name:"),
		nameEntry,
		widget.NewLabel("Peer ID:"),
		idEntry,
		widget.NewSeparator(),
		container.NewHBox(addBtn, cancelBtn),
	)
	
	addWindow.SetContent(content)
	addWindow.Resize(fyne.NewSize(400, 250))
	addWindow.Show()
}

// Run runs the GUI
func (gui *O2OChatGUI) Run() {
	gui.Initialize()
	gui.window.SetContent(gui.BuildUI())
	gui.window.ShowAndRun()
}

// Main function for Windows GUI
func main() {
	gui := NewO2OChatGUI()
	gui.Run()
}
