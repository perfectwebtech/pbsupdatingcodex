package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"streaming-gateway/internal/service"
)

// PowerfulFeaturesHandler handles HTTP requests for the powerful features
// (Dynamic Pricing, Live Chat, Multi-CDN, Fraud Detection, AVOD)
type PowerfulFeaturesHandler struct {
	pricingService *service.DynamicPricingService
	chatService    *service.LiveChatService
	cdnService     *service.MultiCDNService
	fraudService   *service.FraudDetectionService
	avodService    *service.AVODService
}

// NewPowerfulFeaturesHandler creates a new handler for powerful features
func NewPowerfulFeaturesHandler(
	pricing *service.DynamicPricingService,
	chat *service.LiveChatService,
	cdn *service.MultiCDNService,
	fraud *service.FraudDetectionService,
	avod *service.AVODService,
) *PowerfulFeaturesHandler {
	return &PowerfulFeaturesHandler{
		pricingService: pricing,
		chatService:    chat,
		cdnService:     cdn,
		fraudService:   fraud,
		avodService:    avod,
	}
}

// =====================================================
// HELPERS
// =====================================================

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func getQuery(r *http.Request, key, defaultVal string) string {
	v := r.URL.Query().Get(key)
	if v == "" {
		return defaultVal
	}
	return v
}

// =====================================================
// DYNAMIC PRICING ENDPOINTS
// =====================================================

// CalculatePrice calculates dynamic price for a user
func (h *PowerfulFeaturesHandler) CalculatePrice(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	packageID := r.URL.Query().Get("package_id")
	country := getQuery(r, "country", "US")

	if userID == "" || packageID == "" {
		writeError(w, http.StatusBadRequest, "user_id and package_id are required")
		return
	}

	calc, err := h.pricingService.CalculatePrice(r.Context(), userID, packageID, country)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, calc)
}

// GetChurnPreventionPrice returns special price for at-risk users
func (h *PowerfulFeaturesHandler) GetChurnPreventionPrice(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	packageID := r.URL.Query().Get("package_id")

	calc, err := h.pricingService.CalculateChurnPreventionPrice(r.Context(), userID, packageID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, calc)
}

// CreatePricingRule creates a new pricing rule
func (h *PowerfulFeaturesHandler) CreatePricingRule(w http.ResponseWriter, r *http.Request) {
	var rule service.PricingRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.pricingService.CreatePricingRule(r.Context(), &rule); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, rule)
}

// ListPricingRules lists active pricing rules
func (h *PowerfulFeaturesHandler) ListPricingRules(w http.ResponseWriter, r *http.Request) {
	packageID := r.URL.Query().Get("package_id")
	rules, err := h.pricingService.ListPricingRules(r.Context(), packageID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"rules": rules})
}

// =====================================================
// LIVE CHAT ENDPOINTS
// =====================================================

// CreateChatRoom creates a chat room for a stream
func (h *PowerfulFeaturesHandler) CreateChatRoom(w http.ResponseWriter, r *http.Request) {
	var req struct {
		StreamID string `json:"stream_id"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.StreamID == "" {
		writeError(w, http.StatusBadRequest, "stream_id required")
		return
	}

	room, err := h.chatService.CreateChatRoom(r.Context(), req.StreamID, req.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, room)
}

// JoinChatRoom joins a user to a chat room
func (h *PowerfulFeaturesHandler) JoinChatRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]

	var req struct {
		UserID    string `json:"user_id"`
		Username  string `json:"username"`
		AvatarURL string `json:"avatar_url"`
		Role      string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Role == "" {
		req.Role = "viewer"
	}

	if err := h.chatService.JoinChatRoom(r.Context(), roomID, req.UserID, req.Username, req.AvatarURL, req.Role); err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "joined"})
}

// SendChatMessage sends a chat message
func (h *PowerfulFeaturesHandler) SendChatMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]

	var req struct {
		UserID  string `json:"user_id"`
		Message string `json:"message"`
		Type    string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Type == "" {
		req.Type = "text"
	}

	msg, err := h.chatService.SendMessage(r.Context(), roomID, req.UserID, req.Message, req.Type)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, msg)
}

// SendSuperChat sends a paid super chat
func (h *PowerfulFeaturesHandler) SendSuperChat(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]

	var req struct {
		UserID  string  `json:"user_id"`
		Message string  `json:"message"`
		Amount  float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	msg, err := h.chatService.SendSuperChat(r.Context(), roomID, req.UserID, req.Message, req.Amount)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, msg)
}

// GetRecentMessages returns recent chat messages
func (h *PowerfulFeaturesHandler) GetRecentMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	limit, _ := strconv.Atoi(getQuery(r, "limit", "50"))

	messages, err := h.chatService.GetRecentMessages(r.Context(), roomID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"messages": messages})
}

// AddReaction adds a real-time reaction
func (h *PowerfulFeaturesHandler) AddReaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamID := vars["streamId"]

	var req struct {
		UserID    string `json:"user_id"`
		Emoji     string `json:"emoji"`
		Timestamp int64  `json:"timestamp"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	reaction, err := h.chatService.AddReaction(r.Context(), streamID, req.UserID, req.Emoji, req.Timestamp)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, reaction)
}

// GetStreamReactions returns aggregated reactions
func (h *PowerfulFeaturesHandler) GetStreamReactions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamID := vars["streamId"]
	seconds, _ := strconv.Atoi(getQuery(r, "seconds", "60"))

	reactions, err := h.chatService.GetStreamReactions(r.Context(), streamID, seconds)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"reactions": reactions})
}

// CreatePoll creates a live poll
func (h *PowerfulFeaturesHandler) CreatePoll(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamID := vars["streamId"]

	var req struct {
		Question string   `json:"question"`
		Options  []string `json:"options"`
		Duration int      `json:"duration_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Duration == 0 {
		req.Duration = 60
	}

	poll, err := h.chatService.CreatePoll(r.Context(), streamID, req.Question, req.Options, req.Duration)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, poll)
}

// VotePoll records a poll vote
func (h *PowerfulFeaturesHandler) VotePoll(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pollID := vars["pollId"]

	var req struct {
		OptionID string `json:"option_id"`
		UserID   string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.chatService.VotePoll(r.Context(), pollID, req.OptionID, req.UserID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "voted"})
}

// =====================================================
// MULTI-CDN ENDPOINTS
// =====================================================

// SelectCDN selects the best CDN for a request
func (h *PowerfulFeaturesHandler) SelectCDN(w http.ResponseWriter, r *http.Request) {
	contentPath := r.URL.Query().Get("content_path")
	country := getQuery(r, "country", "US")
	userIP := r.RemoteAddr

	selection, err := h.cdnService.SelectCDN(r.Context(), userIP, contentPath, country)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, selection)
}

// GetCDNStats returns CDN usage statistics
func (h *PowerfulFeaturesHandler) GetCDNStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	providerID := vars["providerId"]
	days, _ := strconv.Atoi(getQuery(r, "days", "30"))

	stats, err := h.cdnService.GetCDNStats(r.Context(), providerID, days)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"stats": stats})
}

// GetCostOptimization returns cost optimization recommendations
func (h *PowerfulFeaturesHandler) GetCostOptimization(w http.ResponseWriter, r *http.Request) {
	result, err := h.cdnService.GetCostOptimization(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// =====================================================
// FRAUD DETECTION ENDPOINTS
// =====================================================

// CheckFraud performs fraud detection check
func (h *PowerfulFeaturesHandler) CheckFraud(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID    string `json:"user_id"`
		IPAddress string `json:"ip_address"`
		DeviceID  string `json:"device_id"`
		UserAgent string `json:"user_agent"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.IPAddress == "" {
		req.IPAddress = r.RemoteAddr
	}
	if req.UserAgent == "" {
		req.UserAgent = r.Header.Get("User-Agent")
	}

	check, err := h.fraudService.CheckFraud(r.Context(), req.UserID, req.IPAddress, req.DeviceID, req.UserAgent)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, check)
}

// =====================================================
// AVOD ENDPOINTS
// =====================================================

// ServeAd serves an ad based on user/content context
func (h *PowerfulFeaturesHandler) ServeAd(w http.ResponseWriter, r *http.Request) {
	var req service.AdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	ad, err := h.avodService.ServeAd(r.Context(), &req)
	if err != nil {
		writeJSON(w, http.StatusNoContent, map[string]string{"message": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, ad)
}

// GetVASTAd returns a VAST 4.0 XML response for an ad
func (h *PowerfulFeaturesHandler) GetVASTAd(w http.ResponseWriter, r *http.Request) {
	req := &service.AdRequest{
		UserID:     r.URL.Query().Get("user_id"),
		StreamID:   r.URL.Query().Get("stream_id"),
		Position:   getQuery(r, "position", "pre-roll"),
		Country:    getQuery(r, "country", "US"),
		DeviceType: getQuery(r, "device_type", "web"),
	}

	ad, err := h.avodService.ServeAd(r.Context(), req)
	if err != nil {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><VAST version="4.0"></VAST>`))
		return
	}

	impressionID := r.URL.Query().Get("impression_id")
	vast := h.avodService.GenerateVASTResponse(r.Context(), ad, impressionID)

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(vast))
}

// GetAdBreakSchedule returns ad insertion points for content
func (h *PowerfulFeaturesHandler) GetAdBreakSchedule(w http.ResponseWriter, r *http.Request) {
	contentID := r.URL.Query().Get("content_id")
	duration, _ := strconv.Atoi(getQuery(r, "duration", "0"))

	schedule, err := h.avodService.GetAdBreakSchedule(r.Context(), contentID, duration)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, schedule)
}

// RecordImpression records an ad impression
func (h *PowerfulFeaturesHandler) RecordImpression(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	impressionID := vars["impressionId"]
	if err := h.avodService.RecordImpression(r.Context(), impressionID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "recorded"})
}

// TrackAdEvent tracks ad playback events
func (h *PowerfulFeaturesHandler) TrackAdEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	impressionID := vars["impressionId"]
	event := vars["event"]

	if err := h.avodService.TrackAdEvent(r.Context(), impressionID, event); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "tracked"})
}

// RecordAdClick records an ad click
func (h *PowerfulFeaturesHandler) RecordAdClick(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	impressionID := vars["impressionId"]
	if err := h.avodService.RecordClick(r.Context(), impressionID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "click_recorded"})
}

// CreateAdvertisement creates a new advertisement
func (h *PowerfulFeaturesHandler) CreateAdvertisement(w http.ResponseWriter, r *http.Request) {
	var ad service.Advertisement
	if err := json.NewDecoder(r.Body).Decode(&ad); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.avodService.CreateAdvertisement(r.Context(), &ad); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, ad)
}

// GetAdvertiserMetrics returns advertiser performance metrics
func (h *PowerfulFeaturesHandler) GetAdvertiserMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	advertiserID := vars["advertiserId"]
	days, _ := strconv.Atoi(getQuery(r, "days", "30"))

	metrics, err := h.avodService.GetAdvertiserMetrics(r.Context(), advertiserID, days)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, metrics)
}

// =====================================================
// ROUTE REGISTRATION
// =====================================================

// RegisterRoutes registers all powerful feature routes with mux router
func (h *PowerfulFeaturesHandler) RegisterRoutes(r *mux.Router) {
	// Dynamic Pricing
	r.HandleFunc("/pricing/calculate", h.CalculatePrice).Methods("GET")
	r.HandleFunc("/pricing/churn-prevention", h.GetChurnPreventionPrice).Methods("GET")
	r.HandleFunc("/pricing/rules", h.CreatePricingRule).Methods("POST")
	r.HandleFunc("/pricing/rules", h.ListPricingRules).Methods("GET")

	// Live Chat
	r.HandleFunc("/chat/rooms", h.CreateChatRoom).Methods("POST")
	r.HandleFunc("/chat/rooms/{roomId}/join", h.JoinChatRoom).Methods("POST")
	r.HandleFunc("/chat/rooms/{roomId}/messages", h.SendChatMessage).Methods("POST")
	r.HandleFunc("/chat/rooms/{roomId}/super-chat", h.SendSuperChat).Methods("POST")
	r.HandleFunc("/chat/rooms/{roomId}/messages", h.GetRecentMessages).Methods("GET")

	// Stream Reactions & Polls
	r.HandleFunc("/streams/{streamId}/reactions", h.AddReaction).Methods("POST")
	r.HandleFunc("/streams/{streamId}/reactions", h.GetStreamReactions).Methods("GET")
	r.HandleFunc("/streams/{streamId}/polls", h.CreatePoll).Methods("POST")
	r.HandleFunc("/polls/{pollId}/vote", h.VotePoll).Methods("POST")

	// Multi-CDN
	r.HandleFunc("/cdn/select", h.SelectCDN).Methods("GET")
	r.HandleFunc("/cdn/{providerId}/stats", h.GetCDNStats).Methods("GET")
	r.HandleFunc("/cdn/cost-optimization", h.GetCostOptimization).Methods("GET")

	// Fraud Detection
	r.HandleFunc("/fraud/check", h.CheckFraud).Methods("POST")

	// AVOD
	r.HandleFunc("/ads/serve", h.ServeAd).Methods("POST")
	r.HandleFunc("/ads/vast", h.GetVASTAd).Methods("GET")
	r.HandleFunc("/ads/schedule", h.GetAdBreakSchedule).Methods("GET")
	r.HandleFunc("/ads", h.CreateAdvertisement).Methods("POST")
	r.HandleFunc("/avod/impression/{impressionId}", h.RecordImpression).Methods("POST")
	r.HandleFunc("/avod/track/{impressionId}/{event}", h.TrackAdEvent).Methods("POST")
	r.HandleFunc("/avod/click/{impressionId}", h.RecordAdClick).Methods("POST")
	r.HandleFunc("/advertisers/{advertiserId}/metrics", h.GetAdvertiserMetrics).Methods("GET")
}
