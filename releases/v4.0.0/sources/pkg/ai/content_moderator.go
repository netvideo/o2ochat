package ai

import (
	"context"
	"regexp"
	"strings"
	"time"
)

// ContentModerator represents AI content moderation system
type ContentModerator struct {
	rules         []ModerationRule
	userReports   map[string][]UserReport
	autoBanList   []string
	warningCounts map[string]int
}

// ModerationRule represents a moderation rule
type ModerationRule struct {
	ID          string
	Name        string
	Pattern     *regexp.Regexp
	Action      string // "warn", "mute", "ban"
	Severity    int    // 1-5
	Description string
}

// UserReport represents a user report
type UserReport struct {
	ID           string
	ReporterID   string
	ReportedID   string
	Content      string
	Reason       string
	Timestamp    time.Time
	Status       string // "pending", "reviewed", "resolved"
	ModeratorID  string
}

// ModerationResult represents moderation result
type ModerationResult struct {
	IsAppropriate bool
	Score         float64
	Flags         []string
	Action        string
	Reason        string
}

// NewContentModerator creates a new content moderator
func NewContentModerator() *ContentModerator {
	return &ContentModerator{
		rules:         make([]ModerationRule, 0),
		userReports:   make(map[string][]UserReport),
		autoBanList:   []string{},
		warningCounts: make(map[string]int),
	}
}

// SetupDefaultRules sets up default moderation rules
func (cm *ContentModerator) SetupDefaultRules() {
	// Spam detection
	cm.rules = append(cm.rules, ModerationRule{
		ID:          "spam_1",
		Name:        "Spam Detection",
		Pattern:     regexp.MustCompile(`(?i)(buy now|limited offer|click here|free money)`),
		Action:      "warn",
		Severity:    2,
		Description: "Detects spam-like content",
	})

	// Profanity filter
	cm.rules = append(cm.rules, ModerationRule{
		ID:          "profanity_1",
		Name:        "Profanity Filter",
		Pattern:     regexp.MustCompile(`(?i)(bad_word_1|bad_word_2)`),
		Action:      "mute",
		Severity:    3,
		Description: "Filters profanity",
	})

	// Hate speech detection
	cm.rules = append(cm.rules, ModerationRule{
		ID:          "hate_1",
		Name:        "Hate Speech Detection",
		Pattern:     regexp.MustCompile(`(?i)(hate|racist|discriminate)`),
		Action:      "ban",
		Severity:    5,
		Description: "Detects hate speech",
	})

	// Personal information detection
	cm.rules = append(cm.rules, ModerationRule{
		ID:          "pii_1",
		Name:        "PII Detection",
		Pattern:     regexp.MustCompile(`\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`), // Phone numbers
		Action:      "warn",
		Severity:    2,
		Description: "Detects personal information",
	})

	// URL spam
	cm.rules = append(cm.rules, ModerationRule{
		ID:          "url_spam_1",
		Name:        "URL Spam",
		Pattern:     regexp.MustCompile(`(http|https):\/\/[^\s]+`),
		Action:      "warn",
		Severity:    1,
		Description: "Detects URLs",
	})
}

// ModerateContent moderates content
func (cm *ContentModerator) ModerateContent(ctx context.Context, userID, content string) (*ModerationResult, error) {
	result := &ModerationResult{
		IsAppropriate: true,
		Score:         1.0,
		Flags:         make([]string, 0),
		Action:        "approve",
		Reason:        "Content is appropriate",
	}

	// Check against auto-ban list
	for _, banned := range cm.autoBanList {
		if strings.Contains(content, banned) {
			result.IsAppropriate = false
			result.Score = 0.0
			result.Action = "ban"
			result.Reason = "Banned content detected"
			return result, nil
		}
	}

	// Check against rules
	maxSeverity := 0
	for _, rule := range cm.rules {
		if rule.Pattern.MatchString(content) {
			result.Flags = append(result.Flags, rule.Name)
			result.IsAppropriate = false
			result.Score -= 0.2

			if rule.Severity > maxSeverity {
				maxSeverity = rule.Severity
				result.Action = rule.Action
				result.Reason = rule.Description
			}
		}
	}

	// Adjust score based on severity
	if maxSeverity >= 4 {
		result.Score = 0.0
	} else if maxSeverity >= 3 {
		result.Score = 0.3
	} else if maxSeverity >= 2 {
		result.Score = 0.6
	}

	// Check warning count
	if cm.warningCounts[userID] >= 3 {
		result.Action = "mute"
		result.Reason = "Multiple warnings"
	}

	return result, nil
}

// ReportContent allows users to report content
func (cm *ContentModerator) ReportContent(ctx context.Context, reporterID, reportedID, content, reason string) (string, error) {
	report := UserReport{
		ID:          generateReportID(),
		ReporterID:  reporterID,
		ReportedID:  reportedID,
		Content:     content,
		Reason:      reason,
		Timestamp:   time.Now(),
		Status:      "pending",
	}

	cm.userReports[reportedID] = append(cm.userReports[reportedID], report)

	return report.ID, nil
}

// GetReports gets reports for a user
func (cm *ContentModerator) GetReports(userID string) []UserReport {
	return cm.userReports[userID]
}

// ReviewReport reviews a report
func (cm *ContentModerator) ReviewReport(reportID, moderatorID, decision string) error {
	// Find and update report
	for userID, reports := range cm.userReports {
		for i, report := range reports {
			if report.ID == reportID {
				cm.userReports[userID][i].Status = "resolved"
				cm.userReports[userID][i].ModeratorID = moderatorID

				// Apply action based on decision
				if decision == "approve" {
					// No action
				} else if decision == "warn" {
					cm.warningCounts[userID]++
				} else if decision == "mute" {
					cm.warningCounts[userID] = 10 // Auto-mute
				} else if decision == "ban" {
					cm.autoBanList = append(cm.autoBanList, userID)
				}

				return nil
			}
		}
	}

	return nil
}

// AddRule adds a new moderation rule
func (cm *ContentModerator) AddRule(rule ModerationRule) {
	cm.rules = append(cm.rules, rule)
}

// RemoveRule removes a moderation rule
func (cm *ContentModerator) RemoveRule(ruleID string) {
	for i, rule := range cm.rules {
		if rule.ID == ruleID {
			cm.rules = append(cm.rules[:i], cm.rules[i+1:]...)
			break
		}
	}
}

// GetRules gets all moderation rules
func (cm *ContentModerator) GetRules() []ModerationRule {
	return cm.rules
}

// ClearWarnings clears warnings for a user
func (cm *ContentModerator) ClearWarnings(userID string) {
	cm.warningCounts[userID] = 0
}

// GetModerationStats gets moderation statistics
func (cm *ContentModerator) GetModerationStats() map[string]interface{} {
	totalReports := 0
	pendingReports := 0
	for _, reports := range cm.userReports {
		totalReports += len(reports)
		for _, report := range reports {
			if report.Status == "pending" {
				pendingReports++
			}
		}
	}

	return map[string]interface{}{
		"total_rules":      len(cm.rules),
		"total_reports":    totalReports,
		"pending_reports":  pendingReports,
		"banned_users":     len(cm.autoBanList),
		"warned_users":     len(cm.warningCounts),
	}
}

// generateReportID generates a unique report ID
func generateReportID() string {
	return "report-" + time.Now().Format("20060102150405")
}
