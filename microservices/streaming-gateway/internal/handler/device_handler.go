package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// DeviceHandler handles HTTP requests for device management
type DeviceHandler struct {
	deviceService *DeviceService
}

// NewDeviceHandler creates a new device handler
func NewDeviceHandler(deviceService *DeviceService) *DeviceHandler {
	return &DeviceHandler{
		deviceService: deviceService,
	}
}

// =====================================================
// REQUEST/RESPONSE DTOs
// =====================================================

// Device represents a registered device
type Device struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	Username        string    `json:"username,omitempty"`
	DeviceType      string    `json:"device_type"`
	DeviceID        string    `json:"device_id"`
	DeviceName      string    `json:"device_name"`
	MACAddress      string    `json:"mac_address,omitempty"`
	IPAddress       string    `json:"ip_address,omitempty"`
	UserAgent       string    `json:"user_agent,omitempty"`
	AppVersion      string    `json:"app_version,omitempty"`
	OSVersion       string    `json:"os_version,omitempty"`
	IsActive        bool      `json:"is_active"`
	IsBlocked       bool      `json:"is_blocked"`
	LastSeenAt      *time.Time `json:"last_seen_at,omitempty"`
	ActivatedAt     time.Time `json:"activated_at"`
	BlockedAt       *time.Time `json:"blocked_at,omitempty"`
	BlockReason     string    `json:"block_reason,omitempty"`
	Notes           string    `json:"notes,omitempty"`
	ActiveSessions  int       `json:"active_sessions"`
	TotalSessions   int       `json:"total_sessions"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// DeviceSession represents an active device session
type DeviceSession struct {
	ID            int64     `json:"id"`
	DeviceID      int64     `json:"device_id"`
	UserID        int64     `json:"user_id"`
	StreamID      *int64    `json:"stream_id,omitempty"`
	StreamName    string    `json:"stream_name,omitempty"`
	SessionToken  string    `json:"session_token,omitempty"`
	IPAddress     string    `json:"ip_address"`
	UserAgent     string    `json:"user_agent,omitempty"`
	StartedAt     time.Time `json:"started_at"`
	LastActivityAt time.Time `json:"last_activity_at"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	IsActive      bool      `json:"is_active"`
	BytesStreamed int64     `json:"bytes_streamed"`
	Duration      int       `json:"duration"` // in seconds
	CreatedAt     time.Time `json:"created_at"`
}

// RegisterDeviceRequest represents a device registration request
type RegisterDeviceRequest struct {
	UserID      int64  `json:"user_id" binding:"required"`
	DeviceType  string `json:"device_type" binding:"required,oneof=mag enigma2 android ios web stb smart_tv"`
	DeviceID    string `json:"device_id" binding:"required"`
	DeviceName  string `json:"device_name" binding:"required"`
	MACAddress  string `json:"mac_address"`
	IPAddress   string `json:"ip_address"`
	UserAgent   string `json:"user_agent"`
	AppVersion  string `json:"app_version"`
	OSVersion   string `json:"os_version"`
}

// UpdateDeviceRequest represents a device update request
type UpdateDeviceRequest struct {
	DeviceName string `json:"device_name"`
	Notes      string `json:"notes"`
	IsActive   *bool  `json:"is_active"`
}

// BlockDeviceRequest represents a device block request
type BlockDeviceRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// DeviceStats represents device statistics
type DeviceStats struct {
	TotalDevices      int64              `json:"total_devices"`
	ActiveDevices     int64              `json:"active_devices"`
	BlockedDevices    int64              `json:"blocked_devices"`
	DevicesByType     map[string]int64   `json:"devices_by_type"`
	ActiveSessions    int64              `json:"active_sessions"`
	TotalBandwidth    int64              `json:"total_bandwidth"` // bytes
	TopDevices        []DeviceUsageStats `json:"top_devices"`
}

// DeviceUsageStats represents usage statistics for a device
type DeviceUsageStats struct {
	DeviceID      int64  `json:"device_id"`
	DeviceName    string `json:"device_name"`
	DeviceType    string `json:"device_type"`
	TotalSessions int    `json:"total_sessions"`
	BytesStreamed int64  `json:"bytes_streamed"`
	LastSeen      string `json:"last_seen"`
}

// =====================================================
// DEVICE ENDPOINTS
// =====================================================

// ListDevices lists all devices with pagination and filters
// GET /api/v1/admin/devices
func (h *DeviceHandler) ListDevices(c *gin.Context) {
	ctx := context.Background()

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	userID := c.Query("user_id")
	deviceType := c.Query("device_type")
	status := c.Query("status")
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Build filters
	filters := make(map[string]interface{})
	if userID != "" {
		if uid, err := strconv.ParseInt(userID, 10, 64); err == nil {
			filters["user_id"] = uid
		}
	}
	if deviceType != "" {
		filters["device_type"] = deviceType
	}
	if status != "" {
		filters["status"] = status
	}
	if search != "" {
		filters["search"] = search
	}

	devices, total, err := h.deviceService.ListDevices(ctx, limit, offset, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch devices",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": devices,
		"pagination": gin.H{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetDevice gets a device by ID
// GET /api/v1/admin/devices/:id
func (h *DeviceHandler) GetDevice(c *gin.Context) {
	ctx := context.Background()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device ID"})
		return
	}

	device, err := h.deviceService.GetDeviceByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": device})
}

// RegisterDevice registers a new device
// POST /api/v1/admin/devices
func (h *DeviceHandler) RegisterDevice(c *gin.Context) {
	ctx := context.Background()

	var req RegisterDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get IP from request if not provided
	if req.IPAddress == "" {
		req.IPAddress = c.ClientIP()
	}

	device, err := h.deviceService.RegisterDevice(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to register device",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Device registered successfully",
		"data":    device,
	})
}

// UpdateDevice updates a device
// PUT /api/v1/admin/devices/:id
func (h *DeviceHandler) UpdateDevice(c *gin.Context) {
	ctx := context.Background()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device ID"})
		return
	}

	var req UpdateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	device, err := h.deviceService.UpdateDevice(ctx, id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update device",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Device updated successfully",
		"data":    device,
	})
}

// DeleteDevice deletes a device
// DELETE /api/v1/admin/devices/:id
func (h *DeviceHandler) DeleteDevice(c *gin.Context) {
	ctx := context.Background()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device ID"})
		return
	}

	if err := h.deviceService.DeleteDevice(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete device",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device deleted successfully"})
}

// BlockDevice blocks a device
// POST /api/v1/admin/devices/:id/block
func (h *DeviceHandler) BlockDevice(c *gin.Context) {
	ctx := context.Background()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device ID"})
		return
	}

	var req BlockDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	device, err := h.deviceService.BlockDevice(ctx, id, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to block device",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Device blocked successfully",
		"data":    device,
	})
}

// UnblockDevice unblocks a device
// POST /api/v1/admin/devices/:id/unblock
func (h *DeviceHandler) UnblockDevice(c *gin.Context) {
	ctx := context.Background()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device ID"})
		return
	}

	device, err := h.deviceService.UnblockDevice(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to unblock device",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Device unblocked successfully",
		"data":    device,
	})
}

// =====================================================
// DEVICE SESSION ENDPOINTS
// =====================================================

// GetDeviceSessions gets sessions for a device
// GET /api/v1/admin/devices/:id/sessions
func (h *DeviceHandler) GetDeviceSessions(c *gin.Context) {
	ctx := context.Background()

	deviceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device ID"})
		return
	}

	activeOnly := c.DefaultQuery("active_only", "false") == "true"

	sessions, err := h.deviceService.GetDeviceSessions(ctx, deviceID, activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch device sessions",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": sessions})
}

// TerminateDeviceSessions terminates all sessions for a device
// POST /api/v1/admin/devices/:id/sessions/terminate
func (h *DeviceHandler) TerminateDeviceSessions(c *gin.Context) {
	ctx := context.Background()

	deviceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device ID"})
		return
	}

	count, err := h.deviceService.TerminateDeviceSessions(ctx, deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to terminate sessions",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           "Sessions terminated successfully",
		"sessions_terminated": count,
	})
}

// ListActiveSessions lists all active sessions
// GET /api/v1/admin/devices/sessions/active
func (h *DeviceHandler) ListActiveSessions(c *gin.Context) {
	ctx := context.Background()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	offset := (page - 1) * limit

	sessions, total, err := h.deviceService.ListActiveSessions(ctx, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch active sessions",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": sessions,
		"pagination": gin.H{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// =====================================================
// STATISTICS ENDPOINTS
// =====================================================

// GetDeviceStats gets device statistics
// GET /api/v1/admin/devices/stats
func (h *DeviceHandler) GetDeviceStats(c *gin.Context) {
	ctx := context.Background()

	stats, err := h.deviceService.GetDeviceStats(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch device statistics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// RegisterRoutes registers all device routes
func (h *DeviceHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Admin routes
	admin := router.Group("/admin/devices")
	{
		// Device management
		admin.GET("", h.ListDevices)
		admin.GET("/:id", h.GetDevice)
		admin.POST("", h.RegisterDevice)
		admin.PUT("/:id", h.UpdateDevice)
		admin.DELETE("/:id", h.DeleteDevice)
		admin.POST("/:id/block", h.BlockDevice)
		admin.POST("/:id/unblock", h.UnblockDevice)

		// Device sessions
		admin.GET("/:id/sessions", h.GetDeviceSessions)
		admin.POST("/:id/sessions/terminate", h.TerminateDeviceSessions)

		// Active sessions
		admin.GET("/sessions/active", h.ListActiveSessions)

		// Statistics
		admin.GET("/stats", h.GetDeviceStats)
	}
}
