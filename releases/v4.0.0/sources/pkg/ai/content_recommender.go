package ai

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// ContentRecommender represents AI content recommendation system
type ContentRecommender struct {
	userProfiles  map[string]*UserProfile
	itemCatalog   map[string]*Item
	interactions  []UserInteraction
	modelWeights  ModelWeights
}

// UserProfile represents a user's profile for recommendations
type UserProfile struct {
	UserID           string
	Interests        []string
	InteractionCount int
	LastActive       time.Time
	Preferences      map[string]float64
}

// Item represents a recommendable item (message, contact, group, etc.)
type Item struct {
	ID        string
	Type      string // "message", "contact", "group", "file"
	Category  string
	Tags      []string
	Content   string
	Popularity float64
	Timestamp time.Time
}

// UserInteraction represents a user-item interaction
type UserInteraction struct {
	UserID    string
	ItemID    string
	Action    string // "view", "like", "share", "reply"
	Timestamp time.Time
	Score     float64
}

// ModelWeights represents recommendation model weights
type ModelWeights struct {
	PopularityWeight float64
	RecencyWeight    float64
	SimilarityWeight float64
	DiversityWeight  float64
}

// Recommendation represents a recommendation result
type Recommendation struct {
	ItemID     string
	Score      float64
	Reason     string
	ItemType   string
	ItemContent string
}

// NewContentRecommender creates a new content recommender
func NewContentRecommender() *ContentRecommender {
	return &ContentRecommender{
		userProfiles: make(map[string]*UserProfile),
		itemCatalog:  make(map[string]*Item),
		interactions: make([]UserInteraction, 0),
		modelWeights: ModelWeights{
			PopularityWeight: 0.3,
			RecencyWeight:    0.3,
			SimilarityWeight: 0.3,
			DiversityWeight:  0.1,
		},
	}
}

// CreateUserProfile creates a user profile
func (cr *ContentRecommender) CreateUserProfile(userID string, interests []string) *UserProfile {
	profile := &UserProfile{
		UserID:           userID,
		Interests:        interests,
		InteractionCount: 0,
		LastActive:       time.Now(),
		Preferences:      make(map[string]float64),
	}

	cr.userProfiles[userID] = profile
	return profile
}

// AddItem adds an item to catalog
func (cr *ContentRecommender) AddItem(item *Item) {
	cr.itemCatalog[item.ID] = item
}

// RecordInteraction records a user-item interaction
func (cr *ContentRecommender) RecordInteraction(userID, itemID, action string) {
	interaction := UserInteraction{
		UserID:    userID,
		ItemID:    itemID,
		Action:    action,
		Timestamp: time.Now(),
		Score:     cr.getActionScore(action),
	}

	cr.interactions = append(cr.interactions, interaction)

	// Update user profile
	if profile, exists := cr.userProfiles[userID]; exists {
		profile.InteractionCount++
		profile.LastActive = time.Now()

		// Update preferences based on item category
		if item, exists := cr.itemCatalog[itemID]; exists {
			currentPref := profile.Preferences[item.Category]
			profile.Preferences[item.Category] = currentPref + interaction.Score*0.1
		}
	}
}

// getActionScore gets score for an action
func (cr *ContentRecommender) getActionScore(action string) float64 {
	scores := map[string]float64{
		"view":   0.1,
		"like":   0.5,
		"share":  0.8,
		"reply":  0.6,
		"ignore": -0.2,
	}

	if score, exists := scores[action]; exists {
		return score
	}
	return 0.0
}

// Recommend generates recommendations for a user
func (cr *ContentRecommender) Recommend(ctx context.Context, userID string, limit int) ([]Recommendation, error) {
	profile, exists := cr.userProfiles[userID]
	if !exists {
		return []Recommendation{}, fmt.Errorf("user not found")
	}

	// Calculate scores for all items
	scores := make(map[string]float64)
	reasons := make(map[string]string)

	for itemID, item := range cr.itemCatalog {
		score := cr.calculateScore(profile, item, userID)
		scores[itemID] = score
		reasons[itemID] = cr.generateReason(profile, item)
	}

	// Sort by score
	recommendations := make([]Recommendation, 0)
	for itemID, score := range scores {
		item := cr.itemCatalog[itemID]
		recommendations = append(recommendations, Recommendation{
			ItemID:      itemID,
			Score:       score,
			Reason:      reasons[itemID],
			ItemType:    item.Type,
			ItemContent: item.Content,
		})
	}

	// Sort recommendations by score
	for i := 0; i < len(recommendations)-1; i++ {
		for j := i + 1; j < len(recommendations); j++ {
			if recommendations[j].Score > recommendations[i].Score {
				recommendations[i], recommendations[j] = recommendations[j], recommendations[i]
			}
		}
	}

	// Limit results
	if limit > 0 && len(recommendations) > limit {
		recommendations = recommendations[:limit]
	}

	return recommendations, nil
}

// calculateScore calculates recommendation score
func (cr *ContentRecommender) calculateScore(profile *UserProfile, item *Item, userID string) float64 {
	// Popularity score
	popularityScore := item.Popularity

	// Recency score
	recencyScore := cr.calculateRecencyScore(item.Timestamp)

	// Similarity score
	similarityScore := cr.calculateSimilarityScore(profile, item)

	// Diversity score
	diversityScore := cr.calculateDiversityScore(userID, item)

	// Weighted sum
	score := cr.modelWeights.PopularityWeight*popularityScore +
		cr.modelWeights.RecencyWeight*recencyScore +
		cr.modelWeights.SimilarityWeight*similarityScore +
		cr.modelWeights.DiversityWeight*diversityScore

	return score
}

// calculateRecencyScore calculates recency score
func (cr *ContentRecommender) calculateRecencyScore(timestamp time.Time) float64 {
	hoursSince := time.Since(timestamp).Hours()

	// Exponential decay
	score := 1.0 / (1.0 + hoursSince/24.0)
	return score
}

// calculateSimilarityScore calculates similarity score
func (cr *ContentRecommender) calculateSimilarityScore(profile *UserProfile, item *Item) float64 {
	score := 0.0

	// Interest matching
	for _, interest := range profile.Interests {
		if strings.ToLower(interest) == strings.ToLower(item.Category) {
			score += 0.5
		}
		for _, tag := range item.Tags {
			if strings.ToLower(interest) == strings.ToLower(tag) {
				score += 0.3
			}
		}
	}

	// Preference matching
	if pref, exists := profile.Preferences[item.Category]; exists {
		score += pref
	}

	return score
}

// calculateDiversityScore calculates diversity score
func (cr *ContentRecommender) calculateDiversityScore(userID string, item *Item) float64 {
	// Check user's recent interactions
	recentCategories := make(map[string]int)

	for i := len(cr.interactions) - 1; i >= 0 && i > len(cr.interactions)-10; i-- {
		interaction := cr.interactions[i]
		if interaction.UserID == userID {
			if item, exists := cr.itemCatalog[interaction.ItemID]; exists {
				recentCategories[item.Category]++
			}
		}
	}

	// Bonus for diversity
	if count, exists := recentCategories[item.Category]; exists {
		return 1.0 / (1.0 + float64(count))
	}

	return 1.0
}

// generateReason generates recommendation reason
func (cr *ContentRecommender) generateReason(profile *UserProfile, item *Item) string {
	reasons := []string{}

	// Check interest match
	for _, interest := range profile.Interests {
		if strings.ToLower(interest) == strings.ToLower(item.Category) {
			reasons = append(reasons, fmt.Sprintf("Matches your interest in %s", interest))
		}
	}

	// Check popularity
	if item.Popularity > 0.8 {
		reasons = append(reasons, "Popular content")
	}

	// Check recency
	if time.Since(item.Timestamp).Hours() < 24 {
		reasons = append(reasons, "Recently posted")
	}

	if len(reasons) > 0 {
		return reasons[0]
	}

	return "Recommended for you"
}

// UpdateModelWeights updates model weights
func (cr *ContentRecommender) UpdateModelWeights(weights ModelWeights) {
	cr.modelWeights = weights
}

// GetUserProfile gets user profile
func (cr *ContentRecommender) GetUserProfile(userID string) *UserProfile {
	return cr.userProfiles[userID]
}

// GetRecommendationStats gets recommendation statistics
func (cr *ContentRecommender) GetRecommendationStats() map[string]interface{} {
	totalInteractions := len(cr.interactions)
	totalItems := len(cr.itemCatalog)
	totalUsers := len(cr.userProfiles)

	return map[string]interface{}{
		"total_users":        totalUsers,
		"total_items":        totalItems,
		"total_interactions": totalInteractions,
		"avg_interactions_per_user": float64(totalInteractions) / float64(totalUsers),
	}
}

// SimulateRecommendations simulates recommendation generation for testing
func SimulateRecommendations() {
	cr := NewContentRecommender()

	// Create user profiles
	cr.CreateUserProfile("user1", []string{"technology", "science", "music"})
	cr.CreateUserProfile("user2", []string{"sports", "news", "entertainment"})

	// Add items
	items := []*Item{
		{ID: "1", Type: "message", Category: "technology", Tags: []string{"AI", "ML"}, Content: "AI news", Popularity: 0.9, Timestamp: time.Now()},
		{ID: "2", Type: "message", Category: "sports", Tags: []string{"football"}, Content: "Sports update", Popularity: 0.7, Timestamp: time.Now().Add(-24 * time.Hour)},
		{ID: "3", Type: "group", Category: "music", Tags: []string{"rock"}, Content: "Music group", Popularity: 0.8, Timestamp: time.Now().Add(-12 * time.Hour)},
	}

	for _, item := range items {
		cr.AddItem(item)
	}

	// Record interactions
	cr.RecordInteraction("user1", "1", "like")
	cr.RecordInteraction("user1", "3", "view")
	cr.RecordInteraction("user2", "2", "share")

	// Generate recommendations
	ctx := context.Background()
	recs, _ := cr.Recommend(ctx, "user1", 5)

	fmt.Println("Recommendations for user1:")
	for _, rec := range recs {
		fmt.Printf("- %s (Score: %.2f) - %s\n", rec.ItemID, rec.Score, rec.Reason)
	}
}
