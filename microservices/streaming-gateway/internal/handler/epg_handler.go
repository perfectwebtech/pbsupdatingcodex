package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// EPGHandler handles HTTP requests for EPG management
type EPGHandler struct {
	epgService *EPGService
}

// NewEPGHandler creates a new EPG handler
func NewEPGHandler(epgService *EPGService) *EPGHandler {
	return &EPGHandler{
		epgService: epgService,
	}
}

// =====================================================
// REQUEST/RESPONSE DTOs
// =====================================================

// EPGProgram represents an EPG program
type EPGProgram struct {
	ID          int64                  `json:"id"`
	StreamID    int64                  `json:"stream_id"`
	StreamName  string                 `json:"stream_name,omitempty"`
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	Category    string                 `json:"category,omitempty"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     time.Time              `json:"end_time"`
	Duration    int                    `json:"duration"`
	ImageURL    string                 `json:"image_url,omitempty"`
	IconURL     string                 `json:"icon_url,omitempty"`
	Episode     int                    `json:"episode_number,omitempty"`
	Season      int                    `json:"season_number,omitempty"`
	Year        int                    `json:"year,omitempty"`
	Rating      string                 `json:"rating,omitempty"`
	Directors   []string               `json:"directors,omitempty"`
	Actors      []string               `json:"actors,omitempty"`
	Country     string                 `json:"country,omitempty"`
	Language    string                 `json:"language,omitempty"`
	IsLive      bool                   `json:"is_live"`
	IsRepeat    bool                   `json:"is_repeat"`
	IsPremiere  bool                   `json:"is_premiere"`
	Status      string                 `json:"status,omitempty"` // now_playing, upcoming, past
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// EPGSource represents an EPG data source
type EPGSource struct {
	ID              int64                  `json:"id"`
	Name            string                 `json:"name"`
	SourceType      string                 `json:"source_type"`
	URL             string                 `json:"url,omitempty"`
	UpdateInterval  int                    `json:"update_interval"`
	FormatConfig    map[string]interface{} `json:"format_config,omitempty"`
	IsActive        bool                   `json:"is_active"`
	LastSync        *time.Time             `json:"last_sync,omitempty"`
	LastSyncStatus  string                 `json:"last_sync_status,omitempty"`
	LastError       string                 `json:"last_error,omitempty"`
	ProgramsImported int                   `json:"programs_imported"`
	ProgramsUpdated  int                   `json:"programs_updated"`
	ProgramsFailed   int                   `json:"programs_failed"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// EPGImportLog represents an EPG import record
type EPGImportLog struct {
	ID               int64                  `json:"id"`
	SourceID         *int64                 `json:"source_id,omitempty"`
	ImportType       string                 `json:"import_type"`
	Status           string                 `json:"status"`
	ProgramsProcessed int                   `json:"programs_processed"`
	ProgramsCreated   int                   `json:"programs_created"`
	ProgramsUpdated   int                   `json:"programs_updated"`
	ProgramsDeleted   int                   `json:"programs_deleted"`
	ProgramsFailed    int                   `json:"programs_failed"`
	StartedAt        time.Time              `json:"started_at"`
	CompletedAt      *time.Time             `json:"completed_at,omitempty"`
	DurationSeconds  int                    `json:"duration_seconds,omitempty"`
	ErrorMessage     string                 `json:"error_message,omitempty"`
	ErrorDetails     map[string]interface{} `json:"error_details,omitempty"`
	ImportFile       string                 `json:"import_file,omitempty"`
	ImportedBy       *int64                 `json:"imported_by,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
}

// CreateEPGProgramRequest represents a request to create an EPG program
type CreateEPGProgramRequest struct {
	StreamID    int64    `json:"stream_id" binding:"required"`
	Title       string   `json:"title" binding:"required,min=1,max=500"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	StartTime   string   `json:"start_time" binding:"required"` // ISO 8601 format
	EndTime     string   `json:"end_time" binding:"required"`
	ImageURL    string   `json:"image_url"`
	IconURL     string   `json:"icon_url"`
	Episode     int      `json:"episode_number"`
	Season      int      `json:"season_number"`
	Year        int      `json:"year"`
	Rating      string   `json:"rating"`
	Directors   []string `json:"directors"`
	Actors      []string `json:"actors"`
	Country     string   `json:"country"`
	Language    string   `json:"language"`
	IsLive      bool     `json:"is_live"`
	IsRepeat    bool     `json:"is_repeat"`
	IsPremiere  bool     `json:"is_premiere"`
}

// UpdateEPGProgramRequest represents a request to update an EPG program
type UpdateEPGProgramRequest struct {
	Title       string   `json:"title" binding:"min=1,max=500"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	StartTime   string   `json:"start_time"`
	EndTime     string   `json:"end_time"`
	ImageURL    string   `json:"image_url"`
	Episode     int      `json:"episode_number"`
	Season      int      `json:"season_number"`
	Year        int      `json:"year"`
	Rating      string   `json:"rating"`
	Directors   []string `json:"directors"`
	Actors      []string `json:"actors"`
	IsLive      bool     `json:"is_live"`
	IsRepeat    bool     `json:"is_repeat"`
	IsPremiere  bool     `json:"is_premiere"`
}

// CreateEPGSourceRequest represents a request to create an EPG source
type CreateEPGSourceRequest struct {
	Name           string                 `json:"name" binding:"required,min=1,max=255"`
	SourceType     string                 `json:"source_type" binding:"required,oneof=xmltv json api manual"`
	URL            string                 `json:"url"`
	UpdateInterval int                    `json:"update_interval" binding:"min=60"`
	FormatConfig   map[string]interface{} `json:"format_config"`
	Credentials    map[string]interface{} `json:"credentials"`
}

// =====================================================
// EPG PROGRAM ENDPOINTS
// =====================================================

// ListEPGPrograms lists EPG programs with filters
// GET /api/v1/admin/epg/programs
func (h *EPGHandler) ListEPGPrograms(c *gin.Context) {
	ctx := context.Background()

	// Parse query parameters
	streamID := c.Query("stream_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	category := c.Query("category")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}

	offset := (page - 1) * limit

	// Build filters
	filters := make(map[string]interface{})
	if streamID != "" {
		if sid, err := strconv.ParseInt(streamID, 10, 64); err == nil {
			filters["stream_id"] = sid
		}
	}
	if startDate != "" {
		filters["start_date"] = startDate
	}
	if endDate != "" {
		filters["end_date"] = endDate
	}
	if category != "" {
		filters["category"] = category
	}

	programs, total, err := h.epgService.ListPrograms(ctx, limit, offset, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch EPG programs",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": programs,
		"pagination": gin.H{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetEPGProgram gets an EPG program by ID
// GET /api/v1/admin/epg/programs/:id
func (h *EPGHandler) GetEPGProgram(c *gin.Context) {
	ctx := context.Background()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid program ID"})
		return
	}

	program, err := h.epgService.GetProgramByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": program})
}

// CreateEPGProgram creates a new EPG program
// POST /api/v1/admin/epg/programs
func (h *EPGHandler) CreateEPGProgram(c *gin.Context) {
	ctx := context.Background()

	var req CreateEPGProgramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	program, err := h.epgService.CreateProgram(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create EPG program",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "EPG program created successfully",
		"data":    program,
	})
}

// UpdateEPGProgram updates an EPG program
// PUT /api/v1/admin/epg/programs/:id
func (h *EPGHandler) UpdateEPGProgram(c *gin.Context) {
	ctx := context.Background()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid program ID"})
		return
	}

	var req UpdateEPGProgramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	program, err := h.epgService.UpdateProgram(ctx, id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update EPG program",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "EPG program updated successfully",
		"data":    program,
	})
}

// DeleteEPGProgram deletes an EPG program
// DELETE /api/v1/admin/epg/programs/:id
func (h *EPGHandler) DeleteEPGProgram(c *gin.Context) {
	ctx := context.Background()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid program ID"})
		return
	}

	if err := h.epgService.DeleteProgram(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete EPG program",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "EPG program deleted successfully"})
}

// GetCurrentPrograms gets currently playing programs
// GET /api/v1/epg/current
func (h *EPGHandler) GetCurrentPrograms(c *gin.Context) {
	ctx := context.Background()

	streamID := c.Query("stream_id")

	filters := make(map[string]interface{})
	if streamID != "" {
		if sid, err := strconv.ParseInt(streamID, 10, 64); err == nil {
			filters["stream_id"] = sid
		}
	}

	programs, err := h.epgService.GetCurrentPrograms(ctx, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch current programs",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": programs})
}

// GetSchedule gets EPG schedule for a date range
// GET /api/v1/epg/schedule
func (h *EPGHandler) GetSchedule(c *gin.Context) {
	ctx := context.Background()

	streamID := c.Query("stream_id")
	date := c.DefaultQuery("date", time.Now().Format("2006-01-02"))

	filters := map[string]interface{}{
		"date": date,
	}

	if streamID != "" {
		if sid, err := strconv.ParseInt(streamID, 10, 64); err == nil {
			filters["stream_id"] = sid
		}
	}

	programs, err := h.epgService.GetSchedule(ctx, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch schedule",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": programs})
}

// =====================================================
// EPG SOURCE ENDPOINTS
// =====================================================

// ListEPGSources lists all EPG sources
// GET /api/v1/admin/epg/sources
func (h *EPGHandler) ListEPGSources(c *gin.Context) {
	ctx := context.Background()

	sources, err := h.epgService.ListSources(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch EPG sources",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": sources})
}

// CreateEPGSource creates a new EPG source
// POST /api/v1/admin/epg/sources
func (h *EPGHandler) CreateEPGSource(c *gin.Context) {
	ctx := context.Background()

	var req CreateEPGSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	source, err := h.epgService.CreateSource(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create EPG source",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "EPG source created successfully",
		"data":    source,
	})
}

// SyncEPGSource triggers EPG import from source
// POST /api/v1/admin/epg/sources/:id/sync
func (h *EPGHandler) SyncEPGSource(c *gin.Context) {
	ctx := context.Background()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid source ID"})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, _ := c.Get("user_id")
	var importedBy *int64
	if uid, ok := userID.(int64); ok {
		importedBy = &uid
	}

	importLog, err := h.epgService.SyncFromSource(ctx, id, importedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to sync EPG source",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "EPG sync started successfully",
		"data":    importLog,
	})
}

// GetImportHistory gets EPG import history
// GET /api/v1/admin/epg/import/history
func (h *EPGHandler) GetImportHistory(c *gin.Context) {
	ctx := context.Background()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	logs, total, err := h.epgService.GetImportHistory(ctx, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch import history",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": logs,
		"pagination": gin.H{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// CleanupOldPrograms removes old EPG data
// DELETE /api/v1/admin/epg/cleanup
func (h *EPGHandler) CleanupOldPrograms(c *gin.Context) {
	ctx := context.Background()

	daysToKeep, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if daysToKeep < 1 {
		daysToKeep = 30
	}

	deletedCount, err := h.epgService.CleanupOldPrograms(ctx, daysToKeep)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to cleanup old programs",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Old programs cleaned up successfully",
		"deleted": deletedCount,
	})
}

// RegisterRoutes registers all EPG routes
func (h *EPGHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Admin routes
	admin := router.Group("/admin/epg")
	{
		// Programs
		admin.GET("/programs", h.ListEPGPrograms)
		admin.GET("/programs/:id", h.GetEPGProgram)
		admin.POST("/programs", h.CreateEPGProgram)
		admin.PUT("/programs/:id", h.UpdateEPGProgram)
		admin.DELETE("/programs/:id", h.DeleteEPGProgram)

		// Sources
		admin.GET("/sources", h.ListEPGSources)
		admin.POST("/sources", h.CreateEPGSource)
		admin.POST("/sources/:id/sync", h.SyncEPGSource)

		// Import
		admin.GET("/import/history", h.GetImportHistory)

		// Maintenance
		admin.DELETE("/cleanup", h.CleanupOldPrograms)
	}

	// Public/User routes
	epg := router.Group("/epg")
	{
		epg.GET("/current", h.GetCurrentPrograms)
		epg.GET("/schedule", h.GetSchedule)
	}
}
