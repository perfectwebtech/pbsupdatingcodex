package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"streaming-gateway/internal/service"
)

// AIRecommendationsHandler handles AI recommendation requests
type AIRecommendationsHandler struct {
	service *service.AIRecommendationsService
}

// NewAIRecommendationsHandler creates a new AI recommendations handler
func NewAIRecommendationsHandler(service *service.AIRecommendationsService) *AIRecommendationsHandler {
	return &AIRecommendationsHandler{
		service: service,
	}
}

// RegisterRoutes registers all recommendation routes
func (h *AIRecommendationsHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/recommendations/personalized", h.GetPersonalizedRecommendations).Methods("GET")
	r.HandleFunc("/recommendations/trending", h.GetTrendingContent).Methods("GET")
	r.HandleFunc("/recommendations/similar/{type}/{id}", h.GetSimilarContent).Methods("GET")
	r.HandleFunc("/recommendations/preferences", h.GetUserPreferences).Methods("GET")
	r.HandleFunc("/recommendations/for-you", h.GetForYou).Methods("GET")
	r.HandleFunc("/recommendations/discover", h.GetDiscoverContent).Methods("GET")
}

// GetPersonalizedRecommendations returns personalized content recommendations
func (h *AIRecommendationsHandler) GetPersonalizedRecommendations(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID := getUserIDFromContext(r.Context())

	// Get limit from query params
	limitStr := r.URL.Query().Get("limit")
	limit, _ := strconv.Atoi(limitStr)
	if limit == 0 {
		limit = 20
	}

	// Get recommendations
	recommendations, err := h.service.GetPersonalizedRecommendations(userID, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get recommendations", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success":         true,
		"recommendations": recommendations,
		"count":           len(recommendations),
	})
}

// GetTrendingContent returns currently trending content
func (h *AIRecommendationsHandler) GetTrendingContent(w http.ResponseWriter, r *http.Request) {
	// Get query params
	limitStr := r.URL.Query().Get("limit")
	limit, _ := strconv.Atoi(limitStr)
	if limit == 0 {
		limit = 20
	}

	category := r.URL.Query().Get("category")

	// Get trending content
	trending, err := h.service.GetTrendingContent(limit, category)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get trending content", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"trending": trending,
		"count":    len(trending),
	})
}

// GetSimilarContent returns content similar to a given item
func (h *AIRecommendationsHandler) GetSimilarContent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contentType := vars["type"]
	contentIDStr := vars["id"]

	contentID, err := strconv.ParseInt(contentIDStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid content ID", err)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit, _ := strconv.Atoi(limitStr)
	if limit == 0 {
		limit = 10
	}

	// Get similar content
	similar, err := h.service.GetSimilarContent(contentType, contentID, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get similar content", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"similar": similar,
		"count":   len(similar),
	})
}

// GetUserPreferences returns analyzed user preferences
func (h *AIRecommendationsHandler) GetUserPreferences(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())

	preferences, err := h.service.GetUserPreferences(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get preferences", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"preferences": preferences,
	})
}

// GetForYou returns a curated "For You" feed
func (h *AIRecommendationsHandler) GetForYou(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())

	// Get personalized recommendations
	personalized, _ := h.service.GetPersonalizedRecommendations(userID, 10)

	// Get trending content
	trending, _ := h.service.GetTrendingContent(5, "")

	// Combine into a feed
	feed := map[string]interface{}{
		"personalized": personalized,
		"trending":     trending,
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"feed":    feed,
	})
}

// GetDiscoverContent returns content for discovery
func (h *AIRecommendationsHandler) GetDiscoverContent(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())

	// Get user preferences
	prefs, err := h.service.GetUserPreferences(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get preferences", err)
		return
	}

	// Get recommendations based on preferences
	recommendations, err := h.service.GetPersonalizedRecommendations(userID, 30)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get recommendations", err)
		return
	}

	// Group by category for discovery
	discover := map[string]interface{}{
		"recommendations": recommendations,
		"preferences":     prefs,
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"discover": discover,
	})
}

// Helper function to get user ID from context
func getUserIDFromContext(ctx context.Context) int64 {
	// This would be set by auth middleware
	userID, ok := ctx.Value("user_id").(int64)
	if !ok {
		return 0
	}
	return userID
}

// Helper function to respond with JSON
func respondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteStatus(status)
	json.NewEncoder(w).Encode(data)
}

// Helper function to respond with error
func respondWithError(w http.ResponseWriter, status int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
		"details": err.Error(),
	})
}
