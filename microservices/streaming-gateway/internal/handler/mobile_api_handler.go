package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// MobileAPIHandler handles all mobile app-specific HTTP requests
type MobileAPIHandler struct {
	service MobileAPIService
}

// MobileAPIService interface defines the business logic for mobile APIs
type MobileAPIService interface {
	// Authentication
	MobileLogin(credentials *MobileLoginRequest) (*MobileAuthResponse, error)
	RefreshToken(refreshToken string) (*MobileAuthResponse, error)
	MobileLogout(deviceID string, userID int64) error

	// Device Management
	RegisterDevice(device *DeviceRegistration) (*Device, error)
	UpdateDevice(deviceID string, updates *DeviceUpdate) error
	GetUserDevices(userID int64) ([]Device, error)
	RemoveDevice(deviceID string, userID int64) error

	// Streams
	GetMobileStreams(userID int64, filters MobileStreamFilters) (*MobileStreamsResponse, error)
	GetStreamURL(streamID int64, deviceID string, quality string) (*StreamURLResponse, error)

	// VOD & Series
	GetMobileVOD(userID int64, filters VODFilters) (*MobileVODResponse, error)
	GetMobileSeries(userID int64) ([]MobileSeries, error)
	GetEpisodes(seriesID int64) ([]Episode, error)

	// Favorites & Watchlist
	GetFavorites(userID int64) (*FavoritesResponse, error)
	AddFavorite(userID int64, contentType string, contentID int64) error
	RemoveFavorite(userID int64, contentType string, contentID int64) error
	GetWatchlist(userID int64) ([]WatchlistItem, error)

	// Continue Watching
	GetContinueWatching(userID int64) ([]ContinueWatchingItem, error)
	UpdateWatchProgress(progress *WatchProgress) error

	// Push Notifications
	RegisterPushToken(token *PushToken) error
	UpdateNotificationSettings(userID int64, settings *NotificationSettings) error
	GetNotifications(userID int64, limit int) ([]Notification, error)
	MarkNotificationRead(notificationID int64) error

	// Offline Downloads
	GetDownloads(deviceID string) ([]Download, error)
	RequestDownload(request *DownloadRequest) (*Download, error)
	DeleteDownload(downloadID int64, deviceID string) error

	// EPG
	GetMobileEPG(channelIDs []int64, startTime time.Time, endTime time.Time) ([]EPGEvent, error)

	// User Profile
	GetMobileProfile(userID int64) (*MobileProfile, error)
	UpdateMobileProfile(userID int64, updates *ProfileUpdate) error

	// App Configuration
	GetAppConfig(platform string, version string) (*AppConfig, error)
}

// Models

type MobileLoginRequest struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required"`
	DeviceID   string `json:"device_id" validate:"required"`
	DeviceName string `json:"device_name" validate:"required"`
	Platform   string `json:"platform" validate:"required,oneof=ios android"`
	AppVersion string `json:"app_version"`
}

type MobileAuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         MobileUser `json:"user"`
}

type MobileUser struct {
	ID            int64     `json:"id"`
	Email         string    `json:"email"`
	Username      string    `json:"username"`
	FullName      string    `json:"full_name"`
	Avatar        string    `json:"avatar"`
	SubscriptionType string `json:"subscription_type"`
	SubscriptionExpiry time.Time `json:"subscription_expiry"`
	MaxDevices    int       `json:"max_devices"`
	ActiveDevices int       `json:"active_devices"`
	MaxStreams    int       `json:"max_streams"`
}

type DeviceRegistration struct {
	DeviceID   string `json:"device_id" validate:"required"`
	DeviceName string `json:"device_name" validate:"required"`
	Platform   string `json:"platform" validate:"required,oneof=ios android"`
	OSVersion  string `json:"os_version"`
	AppVersion string `json:"app_version"`
	UserID     int64  `json:"user_id"`
}

type Device struct {
	ID         int64     `json:"id"`
	DeviceID   string    `json:"device_id"`
	DeviceName string    `json:"device_name"`
	Platform   string    `json:"platform"`
	OSVersion  string    `json:"os_version"`
	AppVersion string    `json:"app_version"`
	LastActive time.Time `json:"last_active"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}

type DeviceUpdate struct {
	DeviceName *string `json:"device_name,omitempty"`
	OSVersion  *string `json:"os_version,omitempty"`
	AppVersion *string `json:"app_version,omitempty"`
}

type MobileStreamFilters struct {
	CategoryID *int64  `json:"category_id,omitempty"`
	Search     string  `json:"search,omitempty"`
	Type       string  `json:"type,omitempty"` // live, vod, series
	Limit      int     `json:"limit"`
	Offset     int     `json:"offset"`
}

type MobileStreamsResponse struct {
	Streams    []MobileStream  `json:"streams"`
	Categories []Category      `json:"categories"`
	Total      int             `json:"total"`
}

type MobileStream struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Logo        string   `json:"logo"`
	CategoryID  int64    `json:"category_id"`
	Type        string   `json:"type"`
	IsLive      bool     `json:"is_live"`
	IsFavorite  bool     `json:"is_favorite"`
	Qualities   []string `json:"qualities"` // sd, hd, fhd, uhd
	EPGEnabled  bool     `json:"epg_enabled"`
	CurrentShow *EPGEvent `json:"current_show,omitempty"`
}

type Category struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	StreamCount int    `json:"stream_count"`
}

type StreamURLResponse struct {
	URL           string    `json:"url"`
	Type          string    `json:"type"` // hls, dash, rtmp
	Quality       string    `json:"quality"`
	ExpiresAt     time.Time `json:"expires_at"`
	Token         string    `json:"token"`
	CDNEnabled    bool      `json:"cdn_enabled"`
	AdaptiveBitrate bool    `json:"adaptive_bitrate"`
}

type VODFilters struct {
	CategoryID *int64 `json:"category_id,omitempty"`
	Search     string `json:"search,omitempty"`
	Year       *int   `json:"year,omitempty"`
	Genre      string `json:"genre,omitempty"`
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
}

type MobileVODResponse struct {
	Movies []MobileMovie `json:"movies"`
	Total  int           `json:"total"`
}

type MobileMovie struct {
	ID          int64    `json:"id"`
	Title       string   `json:"title"`
	Poster      string   `json:"poster"`
	Backdrop    string   `json:"backdrop"`
	Year        int      `json:"year"`
	Duration    int      `json:"duration"` // minutes
	Rating      float64  `json:"rating"`
	Genres      []string `json:"genres"`
	Description string   `json:"description"`
	IsFavorite  bool     `json:"is_favorite"`
	Progress    int      `json:"progress"` // percentage watched
}

type MobileSeries struct {
	ID          int64    `json:"id"`
	Title       string   `json:"title"`
	Poster      string   `json:"poster"`
	Backdrop    string   `json:"backdrop"`
	Year        int      `json:"year"`
	Rating      float64  `json:"rating"`
	Genres      []string `json:"genres"`
	Description string   `json:"description"`
	SeasonCount int      `json:"season_count"`
	EpisodeCount int     `json:"episode_count"`
	IsFavorite  bool     `json:"is_favorite"`
}

type Episode struct {
	ID          int64     `json:"id"`
	SeriesID    int64     `json:"series_id"`
	SeasonNum   int       `json:"season_num"`
	EpisodeNum  int       `json:"episode_num"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Thumbnail   string    `json:"thumbnail"`
	Duration    int       `json:"duration"`
	AirDate     time.Time `json:"air_date"`
	Progress    int       `json:"progress"`
}

type FavoritesResponse struct {
	Streams []MobileStream `json:"streams"`
	Movies  []MobileMovie  `json:"movies"`
	Series  []MobileSeries `json:"series"`
}

type WatchlistItem struct {
	ContentType string    `json:"content_type"` // stream, movie, series
	ContentID   int64     `json:"content_id"`
	AddedAt     time.Time `json:"added_at"`
	Content     interface{} `json:"content"` // Actual content object
}

type ContinueWatchingItem struct {
	ContentType string    `json:"content_type"` // stream, movie, episode
	ContentID   int64     `json:"content_id"`
	Progress    int       `json:"progress"` // percentage
	UpdatedAt   time.Time `json:"updated_at"`
	Content     interface{} `json:"content"`
}

type WatchProgress struct {
	UserID      int64  `json:"user_id"`
	ContentType string `json:"content_type"`
	ContentID   int64  `json:"content_id"`
	Progress    int    `json:"progress"` // percentage
	CurrentTime int    `json:"current_time"` // seconds
	Duration    int    `json:"duration"` // seconds
}

type PushToken struct {
	UserID   int64  `json:"user_id"`
	DeviceID string `json:"device_id"`
	Token    string `json:"token" validate:"required"`
	Platform string `json:"platform" validate:"required,oneof=ios android"`
}

type NotificationSettings struct {
	EnablePush         bool `json:"enable_push"`
	NewContent         bool `json:"new_content"`
	LiveEvents         bool `json:"live_events"`
	Recommendations    bool `json:"recommendations"`
	SubscriptionExpiry bool `json:"subscription_expiry"`
	SystemUpdates      bool `json:"system_updates"`
}

type Notification struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	ImageURL  string    `json:"image_url,omitempty"`
	ActionURL string    `json:"action_url,omitempty"`
	Data      string    `json:"data,omitempty"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

type Download struct {
	ID          int64     `json:"id"`
	DeviceID    string    `json:"device_id"`
	ContentType string    `json:"content_type"`
	ContentID   int64     `json:"content_id"`
	Quality     string    `json:"quality"`
	Status      string    `json:"status"` // pending, downloading, completed, failed
	Progress    int       `json:"progress"`
	FileSize    int64     `json:"file_size"`
	FilePath    string    `json:"file_path"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
	Content     interface{} `json:"content"`
}

type DownloadRequest struct {
	DeviceID    string `json:"device_id" validate:"required"`
	ContentType string `json:"content_type" validate:"required,oneof=movie episode"`
	ContentID   int64  `json:"content_id" validate:"required"`
	Quality     string `json:"quality" validate:"required,oneof=sd hd fhd"`
}

type EPGEvent struct {
	ID          int64     `json:"id"`
	ChannelID   int64     `json:"channel_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Duration    int       `json:"duration"`
	IsLive      bool      `json:"is_live"`
	Progress    int       `json:"progress"`
}

type MobileProfile struct {
	User                 MobileUser            `json:"user"`
	NotificationSettings NotificationSettings  `json:"notification_settings"`
	AppSettings          AppSettings           `json:"app_settings"`
}

type AppSettings struct {
	AutoPlay            bool   `json:"auto_play"`
	PreferredQuality    string `json:"preferred_quality"`
	DataSaver           bool   `json:"data_saver"`
	DownloadOnlyOnWiFi  bool   `json:"download_only_on_wifi"`
	SubtitlesEnabled    bool   `json:"subtitles_enabled"`
	SubtitleLanguage    string `json:"subtitle_language"`
	ParentalControlPIN  string `json:"parental_control_pin,omitempty"`
}

type ProfileUpdate struct {
	FullName    *string      `json:"full_name,omitempty"`
	Avatar      *string      `json:"avatar,omitempty"`
	AppSettings *AppSettings `json:"app_settings,omitempty"`
}

type AppConfig struct {
	MinVersion      string   `json:"min_version"`
	LatestVersion   string   `json:"latest_version"`
	UpdateRequired  bool     `json:"update_required"`
	UpdateURL       string   `json:"update_url,omitempty"`
	Features        []string `json:"features"`
	CDNEnabled      bool     `json:"cdn_enabled"`
	CDNURL          string   `json:"cdn_url,omitempty"`
	StreamProtocol  string   `json:"stream_protocol"` // hls, dash
	MaxDownloads    int      `json:"max_downloads"`
	DownloadExpiry  int      `json:"download_expiry"` // days
	SupportEmail    string   `json:"support_email"`
	SupportPhone    string   `json:"support_phone"`
	PrivacyPolicyURL string  `json:"privacy_policy_url"`
	TermsURL        string   `json:"terms_url"`
}

// NewMobileAPIHandler creates a new mobile API handler
func NewMobileAPIHandler(service MobileAPIService) *MobileAPIHandler {
	return &MobileAPIHandler{
		service: service,
	}
}

// RegisterRoutes registers all mobile API routes
func (h *MobileAPIHandler) RegisterRoutes(r *mux.Router) {
	// Authentication
	r.HandleFunc("/mobile/auth/login", h.Login).Methods("POST")
	r.HandleFunc("/mobile/auth/refresh", h.RefreshToken).Methods("POST")
	r.HandleFunc("/mobile/auth/logout", h.Logout).Methods("POST")

	// Device Management
	r.HandleFunc("/mobile/devices", h.RegisterDevice).Methods("POST")
	r.HandleFunc("/mobile/devices", h.GetDevices).Methods("GET")
	r.HandleFunc("/mobile/devices/{deviceId}", h.UpdateDevice).Methods("PUT")
	r.HandleFunc("/mobile/devices/{deviceId}", h.RemoveDevice).Methods("DELETE")

	// Streams
	r.HandleFunc("/mobile/streams", h.GetStreams).Methods("GET")
	r.HandleFunc("/mobile/streams/{id}/url", h.GetStreamURL).Methods("GET")

	// VOD & Series
	r.HandleFunc("/mobile/vod", h.GetVOD).Methods("GET")
	r.HandleFunc("/mobile/series", h.GetSeries).Methods("GET")
	r.HandleFunc("/mobile/series/{id}/episodes", h.GetEpisodes).Methods("GET")

	// Favorites & Watchlist
	r.HandleFunc("/mobile/favorites", h.GetFavorites).Methods("GET")
	r.HandleFunc("/mobile/favorites", h.AddFavorite).Methods("POST")
	r.HandleFunc("/mobile/favorites/{type}/{id}", h.RemoveFavorite).Methods("DELETE")
	r.HandleFunc("/mobile/watchlist", h.GetWatchlist).Methods("GET")

	// Continue Watching
	r.HandleFunc("/mobile/continue-watching", h.GetContinueWatching).Methods("GET")
	r.HandleFunc("/mobile/watch-progress", h.UpdateWatchProgress).Methods("POST")

	// Push Notifications
	r.HandleFunc("/mobile/push/register", h.RegisterPushToken).Methods("POST")
	r.HandleFunc("/mobile/notifications/settings", h.UpdateNotificationSettings).Methods("PUT")
	r.HandleFunc("/mobile/notifications", h.GetNotifications).Methods("GET")
	r.HandleFunc("/mobile/notifications/{id}/read", h.MarkNotificationRead).Methods("POST")

	// Offline Downloads
	r.HandleFunc("/mobile/downloads", h.GetDownloads).Methods("GET")
	r.HandleFunc("/mobile/downloads", h.RequestDownload).Methods("POST")
	r.HandleFunc("/mobile/downloads/{id}", h.DeleteDownload).Methods("DELETE")

	// EPG
	r.HandleFunc("/mobile/epg", h.GetEPG).Methods("GET")

	// User Profile
	r.HandleFunc("/mobile/profile", h.GetProfile).Methods("GET")
	r.HandleFunc("/mobile/profile", h.UpdateProfile).Methods("PUT")

	// App Configuration
	r.HandleFunc("/mobile/config", h.GetConfig).Methods("GET")
}

// Handler Methods

func (h *MobileAPIHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req MobileLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	authResponse, err := h.service.MobileLogin(&req)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    authResponse,
	})
}

func (h *MobileAPIHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	authResponse, err := h.service.RefreshToken(req.RefreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid refresh token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    authResponse,
	})
}

func (h *MobileAPIHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DeviceID string `json:"device_id"`
		UserID   int64  `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.service.MobileLogout(req.DeviceID, req.UserID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Logout failed", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Logged out successfully",
	})
}

func (h *MobileAPIHandler) RegisterDevice(w http.ResponseWriter, r *http.Request) {
	var req DeviceRegistration
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	device, err := h.service.RegisterDevice(&req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Device registration failed", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    device,
	})
}

func (h *MobileAPIHandler) GetDevices(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	devices, err := h.service.GetUserDevices(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch devices", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    devices,
	})
}

func (h *MobileAPIHandler) UpdateDevice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceID := vars["deviceId"]

	var req DeviceUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.service.UpdateDevice(deviceID, &req); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Device update failed", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Device updated successfully",
	})
}

func (h *MobileAPIHandler) RemoveDevice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceID := vars["deviceId"]

	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	if err := h.service.RemoveDevice(deviceID, userID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Device removal failed", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Device removed successfully",
	})
}

func (h *MobileAPIHandler) GetStreams(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	filters := MobileStreamFilters{
		Search: r.URL.Query().Get("search"),
		Type:   r.URL.Query().Get("type"),
		Limit:  50,
		Offset: 0,
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filters.Limit = limit
		}
	}

	response, err := h.service.GetMobileStreams(userID, filters)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch streams", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    response,
	})
}

func (h *MobileAPIHandler) GetStreamURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid stream ID", err)
		return
	}

	deviceID := r.URL.Query().Get("device_id")
	quality := r.URL.Query().Get("quality")
	if quality == "" {
		quality = "hd"
	}

	urlResponse, err := h.service.GetStreamURL(streamID, deviceID, quality)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get stream URL", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    urlResponse,
	})
}

func (h *MobileAPIHandler) GetVOD(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	filters := VODFilters{
		Search: r.URL.Query().Get("search"),
		Genre:  r.URL.Query().Get("genre"),
		Limit:  50,
		Offset: 0,
	}

	response, err := h.service.GetMobileVOD(userID, filters)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch VOD", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    response,
	})
}

func (h *MobileAPIHandler) GetSeries(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	series, err := h.service.GetMobileSeries(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch series", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    series,
	})
}

func (h *MobileAPIHandler) GetEpisodes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	seriesID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid series ID", err)
		return
	}

	episodes, err := h.service.GetEpisodes(seriesID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch episodes", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    episodes,
	})
}

func (h *MobileAPIHandler) GetFavorites(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	favorites, err := h.service.GetFavorites(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch favorites", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    favorites,
	})
}

func (h *MobileAPIHandler) AddFavorite(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID      int64  `json:"user_id"`
		ContentType string `json:"content_type"`
		ContentID   int64  `json:"content_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.service.AddFavorite(req.UserID, req.ContentType, req.ContentID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to add favorite", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Added to favorites",
	})
}

func (h *MobileAPIHandler) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contentType := vars["type"]
	contentID, _ := strconv.ParseInt(vars["id"], 10, 64)

	userIDStr := r.URL.Query().Get("user_id")
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	if err := h.service.RemoveFavorite(userID, contentType, contentID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to remove favorite", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Removed from favorites",
	})
}

func (h *MobileAPIHandler) GetWatchlist(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	watchlist, err := h.service.GetWatchlist(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch watchlist", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    watchlist,
	})
}

func (h *MobileAPIHandler) GetContinueWatching(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	items, err := h.service.GetContinueWatching(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch continue watching", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    items,
	})
}

func (h *MobileAPIHandler) UpdateWatchProgress(w http.ResponseWriter, r *http.Request) {
	var progress WatchProgress
	if err := json.NewDecoder(r.Body).Decode(&progress); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.service.UpdateWatchProgress(&progress); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update progress", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Progress updated",
	})
}

func (h *MobileAPIHandler) RegisterPushToken(w http.ResponseWriter, r *http.Request) {
	var token PushToken
	if err := json.NewDecoder(r.Body).Decode(&token); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.service.RegisterPushToken(&token); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to register push token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Push token registered",
	})
}

func (h *MobileAPIHandler) UpdateNotificationSettings(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	var settings NotificationSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.service.UpdateNotificationSettings(userID, &settings); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update settings", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Settings updated",
	})
}

func (h *MobileAPIHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	notifications, err := h.service.GetNotifications(userID, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch notifications", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    notifications,
	})
}

func (h *MobileAPIHandler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	notificationID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid notification ID", err)
		return
	}

	if err := h.service.MarkNotificationRead(notificationID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to mark notification as read", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Notification marked as read",
	})
}

func (h *MobileAPIHandler) GetDownloads(w http.ResponseWriter, r *http.Request) {
	deviceID := r.URL.Query().Get("device_id")

	downloads, err := h.service.GetDownloads(deviceID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch downloads", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    downloads,
	})
}

func (h *MobileAPIHandler) RequestDownload(w http.ResponseWriter, r *http.Request) {
	var req DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	download, err := h.service.RequestDownload(&req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to request download", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    download,
	})
}

func (h *MobileAPIHandler) DeleteDownload(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	downloadID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid download ID", err)
		return
	}

	deviceID := r.URL.Query().Get("device_id")

	if err := h.service.DeleteDownload(downloadID, deviceID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete download", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Download deleted",
	})
}

func (h *MobileAPIHandler) GetEPG(w http.ResponseWriter, r *http.Request) {
	// Parse channel IDs
	channelIDsStr := r.URL.Query().Get("channel_ids")
	var channelIDs []int64
	// In production, parse comma-separated channel IDs

	// Parse time range
	startTime := time.Now()
	endTime := startTime.Add(24 * time.Hour)

	events, err := h.service.GetMobileEPG(channelIDs, startTime, endTime)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch EPG", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    events,
	})
}

func (h *MobileAPIHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	profile, err := h.service.GetMobileProfile(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch profile", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    profile,
	})
}

func (h *MobileAPIHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	var updates ProfileUpdate
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.service.UpdateMobileProfile(userID, &updates); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update profile", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Profile updated",
	})
}

func (h *MobileAPIHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	platform := r.URL.Query().Get("platform")
	version := r.URL.Query().Get("version")

	config, err := h.service.GetAppConfig(platform, version)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch config", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    config,
	})
}
