//
//  MainView.swift
//  O2OChat macOS GUI
//
//  Created by O2OChat AI Team
//  Copyright © 2026 O2OChat. All rights reserved.
//

import SwiftUI

/// Main view for O2OChat macOS application
struct MainView: View {
    @State private var selectedContact: String?
    @State private var searchText = ""
    @State private var messageText = ""
    @State private var showingSettings = false
    @State private var showingAddContact = false
    @State private var statusMessage = "Ready - v3.0.0-beta"
    
    // Sample contacts
    let contacts = [
        "Contact 1",
        "Contact 2",
        "Contact 3",
        "Contact 4",
        "Contact 5"
    ]
    
    // Sample messages
    @State private var messages: [Message] = [
        Message(sender: "Contact 1", text: "Hello!", time: Date()),
        Message(sender: "You", text: "Hi there!", time: Date()),
        Message(sender: "Contact 1", text: "How are you?", time: Date())
    ]
    
    var body: some View {
        NavigationView {
            // Sidebar
            contactList
            
            // Main content
            chatView
        }
        .frame(minWidth: 800, minHeight: 600)
        .navigationTitle("O2OChat v3.0.0-beta")
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                Button(action: { showingAddContact = true }) {
                    Label("Add Contact", systemImage: "plus")
                }
            }
            
            ToolbarItem(placement: .secondaryAction) {
                Button(action: { showingSettings = true }) {
                    Label("Settings", systemImage: "gear")
                }
            }
        }
        .sheet(isPresented: $showingSettings) {
            SettingsView()
        }
        .sheet(isPresented: $showingAddContact) {
            AddContactView()
        }
    }
    
    // MARK: - Contact List
    
    private var contactList: some View {
        List(selection: $selectedContact) {
            Section(header: Text("Contacts")) {
                ForEach(filteredContacts, id: \.self) { contact in
                    Text(contact)
                        .tag(contact as String?)
                }
            }
        }
        .listStyle(SidebarListStyle())
        .frame(minWidth: 200)
        .searchable(text: $searchText, prompt: "Search contacts...")
    }
    
    private var filteredContacts: [String] {
        if searchText.isEmpty {
            return contacts
        }
        return contacts.filter { $0.localizedCaseInsensitiveContains(searchText) }
    }
    
    // MARK: - Chat View
    
    private var chatView: some View {
        VStack(spacing: 0) {
            // Messages
            ScrollViewReader { proxy in
                ScrollView {
                    VStack(alignment: .leading, spacing: 8) {
                        ForEach(messages) { message in
                            MessageRow(message: message)
                        }
                    }
                    .padding()
                }
                .onChange(of: messages.count) { _ in
                    withAnimation {
                        proxy.scrollTo(messages.count - 1, anchor: .bottom)
                    }
                }
            }
            
            Divider()
            
            // Message input
            HStack(spacing: 12) {
                // Toolbar buttons
                HStack(spacing: 8) {
                    Button(action: {
                        statusMessage = "Attach file..."
                    }) {
                        Image(systemName: "paperclip")
                            .foregroundColor(.blue)
                    }
                    
                    Button(action: {
                        statusMessage = "Voice message..."
                    }) {
                        Image(systemName: "mic")
                            .foregroundColor(.blue)
                    }
                    
                    Button(action: {
                        statusMessage = "Send image..."
                    }) {
                        Image(systemName: "camera")
                            .foregroundColor(.blue)
                    }
                }
                
                Spacer()
                
                // Send button
                Button(action: sendMessage) {
                    Image(systemName: "paperplane.fill")
                        .foregroundColor(.white)
                        .padding(8)
                        .background(messageText.isEmpty ? Color.gray : Color.blue)
                        .cornerRadius(8)
                }
                .disabled(messageText.isEmpty)
            }
            .padding()
            
            // Text input
            TextEditor(text: $messageText)
                .frame(minHeight: 60)
                .padding(8)
                .background(Color(NSColor.textBackgroundColor))
                .cornerRadius(8)
                .padding([.leading, .trailing, .bottom])
            
            // Status bar
            HStack {
                Text(statusMessage)
                    .font(.caption)
                    .foregroundColor(.secondary)
                    .italic()
                
                Spacer()
                
                Text("v3.0.0-beta")
                    .font(.caption)
                    .foregroundColor(.secondary)
            }
            .padding(4)
            .background(Color(NSColor.controlBackgroundColor))
        }
    }
    
    // MARK: - Actions
    
    private func sendMessage() {
        guard !messageText.isEmpty else { return }
        
        let newMessage = Message(sender: "You", text: messageText, time: Date())
        messages.append(newMessage)
        messageText = ""
        statusMessage = "Message sent!"
        
        // Simulate reply
        DispatchQueue.main.asyncAfter(deadline: .now() + 1.0) {
            let reply = Message(sender: "Contact 1", text: "Thanks!", time: Date())
            messages.append(reply)
            statusMessage = "New message received"
        }
    }
}

// MARK: - Message Model

struct Message: Identifiable {
    let id = UUID()
    let sender: String
    let text: String
    let time: Date
}

// MARK: - Message Row

struct MessageRow: View {
    let message: Message
    
    var body: some View {
        HStack {
            if message.sender == "You" {
                Spacer()
            }
            
            VStack(alignment: message.sender == "You" ? .trailing : .leading, spacing: 4) {
                Text(message.sender)
                    .font(.caption)
                    .foregroundColor(.secondary)
                
                Text(message.text)
                    .padding(10)
                    .background(message.sender == "You" ? Color.blue : Color.gray.opacity(0.2))
                    .foregroundColor(message.sender == "You" ? .white : .primary)
                    .cornerRadius(10)
                
                Text(message.time, style: .time)
                    .font(.caption2)
                    .foregroundColor(.secondary)
            }
            
            if message.sender != "You" {
                Spacer()
            }
        }
        .id(messages.count - 1)
    }
}

// MARK: - Settings View

struct SettingsView: View {
    @Environment(\.dismiss) private var dismiss
    @State private var username = "User"
    @State private var theme = "System"
    @State private var notifications = true
    
    var body: some View {
        NavigationView {
            Form {
                Section(header: Text("Account")) {
                    TextField("Username", text: $username)
                }
                
                Section(header: Text("Appearance")) {
                    Picker("Theme", selection: $theme) {
                        Text("Light").tag("Light")
                        Text("Dark").tag("Dark")
                        Text("System").tag("System")
                    }
                }
                
                Section(header: Text("Notifications")) {
                    Toggle("Enable notifications", isOn: $notifications)
                }
            }
            .frame(width: 400, height: 300)
            .navigationTitle("Settings")
            .toolbar {
                ToolbarItem(placement: .confirmationAction) {
                    Button("Save") {
                        dismiss()
                    }
                }
                
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") {
                        dismiss()
                    }
                }
            }
        }
    }
}

// MARK: - Add Contact View

struct AddContactView: View {
    @Environment(\.dismiss) private var dismiss
    @State private var name = ""
    @State private var peerID = ""
    
    var body: some View {
        NavigationView {
            Form {
                Section(header: Text("Contact Information")) {
                    TextField("Name", text: $name)
                    TextField("Peer ID", text: $peerID)
                }
            }
            .frame(width: 400, height: 250)
            .navigationTitle("Add Contact")
            .toolbar {
                ToolbarItem(placement: .confirmationAction) {
                    Button("Add") {
                        if !name.isEmpty && !peerID.isEmpty {
                            dismiss()
                        }
                    }
                    .disabled(name.isEmpty || peerID.isEmpty)
                }
                
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") {
                        dismiss()
                    }
                }
            }
        }
    }
}

// MARK: - Preview

struct MainView_Previews: PreviewProvider {
    static var previews: some View {
        MainView()
    }
}

// MARK: - Main Entry Point

import SwiftUI

@main
struct O2OChatApp: App {
    var body: some Scene {
        WindowGroup {
            MainView()
        }
        .windowStyle(.automatic)
        .commands {
            CommandGroup(replacing: .newItem) {
                Button("New Chat") {
                    // New chat action
                }
                .keyboardShortcut("n", modifiers: .command)
            }
        }
    }
}
