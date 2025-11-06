package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// SeriesHandler handles HTTP requests for series and episodes management
type SeriesHandler struct {
	seriesService *SeriesService
}

// NewSeriesHandler creates a new series handler
func NewSeriesHandler(seriesService *SeriesService) *SeriesHandler {
	return &SeriesHandler{
		seriesService: seriesService,
	}
}

// =====================================================
// REQUEST/RESPONSE DTOs
// =====================================================

// Series represents a TV series
type Series struct {
	ID           int64                  `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	CategoryID   *int64                 `json:"category_id,omitempty"`
	CategoryName string                 `json:"category_name,omitempty"`
	CoverURL     string                 `json:"cover_url,omitempty"`
	BackdropURL  string                 `json:"backdrop_url,omitempty"`
	TrailerURL   string                 `json:"trailer_url,omitempty"`
	Rating       float64                `json:"rating,omitempty"`
	ReleaseYear  int                    `json:"release_year,omitempty"`
	Genre        string                 `json:"genre,omitempty"`
	Cast         []string               `json:"cast,omitempty"`
	Director     string                 `json:"director,omitempty"`
	Producer     string                 `json:"producer,omitempty"`
	IsActive     bool                   `json:"is_active"`
	IsFeatured   bool                   `json:"is_featured"`
	TotalSeasons int                    `json:"total_seasons"`
	TotalEpisodes int                   `json:"total_episodes"`
	ViewCount    int64                  `json:"view_count"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// Episode represents a TV episode
type Episode struct {
	ID           int64     `json:"id"`
	SeriesID     int64     `json:"series_id"`
	SeriesName   string    `json:"series_name,omitempty"`
	Season       int       `json:"season"`
	Episode      int       `json:"episode"`
	Title        string    `json:"title"`
	Description  string    `json:"description,omitempty"`
	StreamURL    string    `json:"stream_url"`
	ThumbnailURL string    `json:"thumbnail_url,omitempty"`
	Duration     int       `json:"duration,omitempty"` // in seconds
	AirDate      *string   `json:"air_date,omitempty"`
	Rating       float64   `json:"rating,omitempty"`
	IsActive     bool      `json:"is_active"`
	ViewCount    int64     `json:"view_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateSeriesRequest represents a request to create a series
type CreateSeriesRequest struct {
	Name        string   `json:"name" binding:"required,min=1,max=255"`
	Description string   `json:"description"`
	CategoryID  *int64   `json:"category_id"`
	CoverURL    string   `json:"cover_url"`
	BackdropURL string   `json:"backdrop_url"`
	TrailerURL  string   `json:"trailer_url"`
	Rating      float64  `json:"rating" binding:"min=0,max=10"`
	ReleaseYear int      `json:"release_year"`
	Genre       string   `json:"genre"`
	Cast        []string `json:"cast"`
	Director    string   `json:"director"`
	Producer    string   `json:"producer"`
	IsFeatured  bool     `json:"is_featured"`
}

// UpdateSeriesRequest represents a request to update a series
type UpdateSeriesRequest struct {
	Name        string   `json:"name" binding:"min=1,max=255"`
	Description string   `json:"description"`
	CategoryID  *int64   `json:"category_id"`
	CoverURL    string   `json:"cover_url"`
	BackdropURL string   `json:"backdrop_url"`
	TrailerURL  string   `json:"trailer_url"`
	Rating      float64  `json:"rating" binding:"min=0,max=10"`
	ReleaseYear int      `json:"release_year"`
	Genre       string   `json:"genre"`
	Cast        []string `json:"cast"`
	Director    string   `json:"director"`
	Producer    string   `json:"producer"`
	IsActive    *bool    `json:"is_active"`
	IsFeatured  *bool    `json:"is_featured"`
}

// CreateEpisodeRequest represents a request to create an episode
type CreateEpisodeRequest struct {
	SeriesID     int64   `json:"series_id" binding:"required"`
	Season       int     `json:"season" binding:"required,min=1"`
	Episode      int     `json:"episode" binding:"required,min=1"`
	Title        string  `json:"title" binding:"required,min=1,max=255"`
	Description  string  `json:"description"`
	StreamURL    string  `json:"stream_url" binding:"required"`
	ThumbnailURL string  `json:"thumbnail_url"`
	Duration     int     `json:"duration" binding:"min=0"`
	AirDate      *string `json:"air_date"`
	Rating       float64 `json:"rating" binding:"min=0,max=10"`
}

// UpdateEpisodeRequest represents a request to update an episode
type UpdateEpisodeRequest struct {
	Title        string  `json:"title" binding:"min=1,max=255"`
	Description  string  `json:"description"`
	StreamURL    string  `json:"stream_url"`
	ThumbnailURL string  `json:"thumbnail_url"`
	Duration     int     `json:"duration" binding:"min=0"`
	AirDate      *string `json:"air_date"`
	Rating       float64 `json:"rating" binding:"min=0,max=10"`
	IsActive     *bool   `json:"is_active"`
}

// SeasonInfo represents information about a season
type SeasonInfo struct {
	Season       int `json:"season"`
	EpisodeCount int `json:"episode_count"`
}

// =====================================================
// SERIES ENDPOINTS
// =====================================================

// ListSeries lists all series with pagination and filters
// GET /api/v1/admin/series
func (h *SeriesHandler) ListSeries(c *gin.Context) {
	ctx := context.Background()

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	categoryID := c.Query("category_id")
	search := c.Query("search")
	featured := c.Query("featured")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Build filters
	filters := make(map[string]interface{})
	if categoryID != "" {
		if cid, err := strconv.ParseInt(categoryID, 10, 64); err == nil {
			filters["category_id"] = cid
		}
	}
	if search != "" {
		filters["search"] = search
	}
	if featured == "true" {
		filters["featured"] = true
	}

	// Get series
	seriesList, total, err := h.seriesService.ListSeries(ctx, limit, offset, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch series",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": seriesList,
		"pagination": gin.H{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetSeries gets a series by ID
// GET /api/v1/admin/series/:id
func (h *SeriesHandler) GetSeries(c *gin.Context) {
	ctx := context.Background()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid series ID"})
		return
	}

	series, err := h.seriesService.GetSeriesByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": series})
}

// CreateSeries creates a new series
// POST /api/v1/admin/series
func (h *SeriesHandler) CreateSeries(c *gin.Context) {
	ctx := context.Background()

	var req CreateSeriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	series, err := h.seriesService.CreateSeries(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create series",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Series created successfully",
		"data": series,
	})
}

// UpdateSeries updates an existing series
// PUT /api/v1/admin/series/:id
func (h *SeriesHandler) UpdateSeries(c *gin.Context) {
	ctx := context.Background()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid series ID"})
		return
	}

	var req UpdateSeriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	series, err := h.seriesService.UpdateSeries(ctx, id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update series",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Series updated successfully",
		"data": series,
	})
}

// DeleteSeries deletes a series (soft delete)
// DELETE /api/v1/admin/series/:id
func (h *SeriesHandler) DeleteSeries(c *gin.Context) {
	ctx := context.Background()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid series ID"})
		return
	}

	if err := h.seriesService.DeleteSeries(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete series",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Series deleted successfully"})
}

// GetSeriesSeasons gets season information for a series
// GET /api/v1/admin/series/:id/seasons
func (h *SeriesHandler) GetSeriesSeasons(c *gin.Context) {
	ctx := context.Background()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid series ID"})
		return
	}

	seasons, err := h.seriesService.GetSeriesSeasons(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch seasons",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": seasons})
}

// =====================================================
// EPISODE ENDPOINTS
// =====================================================

// ListEpisodes lists episodes for a series
// GET /api/v1/admin/episodes
func (h *SeriesHandler) ListEpisodes(c *gin.Context) {
	ctx := context.Background()

	seriesIDStr := c.Query("series_id")
	if seriesIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "series_id is required"})
		return
	}

	seriesID, err := strconv.ParseInt(seriesIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid series_id"})
		return
	}

	season := c.Query("season")
	var seasonFilter *int
	if season != "" {
		if s, err := strconv.Atoi(season); err == nil {
			seasonFilter = &s
		}
	}

	episodes, err := h.seriesService.ListEpisodes(ctx, seriesID, seasonFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch episodes",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": episodes})
}

// GetEpisode gets an episode by ID
// GET /api/v1/admin/episodes/:id
func (h *SeriesHandler) GetEpisode(c *gin.Context) {
	ctx := context.Background()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid episode ID"})
		return
	}

	episode, err := h.seriesService.GetEpisodeByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": episode})
}

// CreateEpisode creates a new episode
// POST /api/v1/admin/episodes
func (h *SeriesHandler) CreateEpisode(c *gin.Context) {
	ctx := context.Background()

	var req CreateEpisodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	episode, err := h.seriesService.CreateEpisode(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create episode",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Episode created successfully",
		"data": episode,
	})
}

// UpdateEpisode updates an existing episode
// PUT /api/v1/admin/episodes/:id
func (h *SeriesHandler) UpdateEpisode(c *gin.Context) {
	ctx := context.Background()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid episode ID"})
		return
	}

	var req UpdateEpisodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	episode, err := h.seriesService.UpdateEpisode(ctx, id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update episode",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Episode updated successfully",
		"data": episode,
	})
}

// DeleteEpisode deletes an episode
// DELETE /api/v1/admin/episodes/:id
func (h *SeriesHandler) DeleteEpisode(c *gin.Context) {
	ctx := context.Background()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid episode ID"})
		return
	}

	if err := h.seriesService.DeleteEpisode(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete episode",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Episode deleted successfully"})
}

// IncrementViewCount increments view count for an episode
// POST /api/v1/episodes/:id/view
func (h *SeriesHandler) IncrementViewCount(c *gin.Context) {
	ctx := context.Background()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid episode ID"})
		return
	}

	if err := h.seriesService.IncrementEpisodeViewCount(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to increment view count",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "View count incremented"})
}

// RegisterRoutes registers all series routes
func (h *SeriesHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Admin routes
	admin := router.Group("/admin")
	{
		// Series management
		admin.GET("/series", h.ListSeries)
		admin.GET("/series/:id", h.GetSeries)
		admin.POST("/series", h.CreateSeries)
		admin.PUT("/series/:id", h.UpdateSeries)
		admin.DELETE("/series/:id", h.DeleteSeries)
		admin.GET("/series/:id/seasons", h.GetSeriesSeasons)

		// Episode management
		admin.GET("/episodes", h.ListEpisodes)
		admin.GET("/episodes/:id", h.GetEpisode)
		admin.POST("/episodes", h.CreateEpisode)
		admin.PUT("/episodes/:id", h.UpdateEpisode)
		admin.DELETE("/episodes/:id", h.DeleteEpisode)
	}

	// Public/User routes
	router.POST("/episodes/:id/view", h.IncrementViewCount)
}
