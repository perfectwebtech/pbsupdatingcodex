package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/iptv-platform/streaming-gateway/internal/service"
	"go.uber.org/zap"
)

type StreamHandler struct {
	streamService *service.StreamService
	hlsService    *service.HLSService
	logger        *zap.Logger
}

func NewStreamHandler(
	streamService *service.StreamService,
	hlsService *service.HLSService,
	logger *zap.Logger,
) *StreamHandler {
	return &StreamHandler{
		streamService: streamService,
		hlsService:    hlsService,
		logger:        logger,
	}
}

func (h *StreamHandler) ListStreams(c *gin.Context) {
	userID := c.GetInt64("user_id")

	filters := make(map[string]interface{})
	if streamType := c.Query("type"); streamType != "" {
		filters["type"] = streamType
	}
	if categoryID := c.Query("category_id"); categoryID != "" {
		if id, err := strconv.ParseInt(categoryID, 10, 64); err == nil {
			filters["category_id"] = id
		}
	}

	streams, err := h.streamService.ListStreams(c.Request.Context(), userID, filters)
	if err != nil {
		h.logger.Error("failed to list streams", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to retrieve streams",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    streams,
	})
}

func (h *StreamHandler) GetStream(c *gin.Context) {
	userID := c.GetInt64("user_id")

	streamID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid stream ID",
		})
		return
	}

	stream, err := h.streamService.GetStream(c.Request.Context(), userID, streamID)
	if err != nil {
		if err == service.ErrAccessDenied {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "access denied",
			})
			return
		}

		h.logger.Error("failed to get stream", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to retrieve stream",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stream,
	})
}

func (h *StreamHandler) GetStreamURL(c *gin.Context) {
	userID := c.GetInt64("user_id")

	streamID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid stream ID",
		})
		return
	}

	container := c.DefaultQuery("container", "m3u8")

	req := &service.StreamRequest{
		UserID:    userID,
		StreamID:  streamID,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Container: container,
	}

	resp, err := h.streamService.GetStreamURL(c.Request.Context(), req)
	if err != nil {
		switch err {
		case service.ErrAccessDenied:
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "access denied",
			})
		case service.ErrMaxConnectionsReached:
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "maximum connections reached",
			})
		case service.ErrGeoBlocked:
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "geographic location blocked",
			})
		case service.ErrServerUnavailable:
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"success": false,
				"error":   "no server available",
			})
		default:
			h.logger.Error("failed to get stream URL", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "internal server error",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}

func (h *StreamHandler) GetHLSPlaylist(c *gin.Context) {
	userID := c.GetInt64("user_id")

	streamID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid stream ID",
		})
		return
	}

	stream, err := h.streamService.GetStream(c.Request.Context(), userID, streamID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	sessionID := c.Query("session")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session required"})
		return
	}

	// Generate HLS playlist
	baseURL := "/api/v1/stream/" + c.Param("id")
	segmentBaseURL := baseURL + "/segment"

	playlist, err := h.hlsService.GenerateSimplePlaylist(segmentBaseURL, sessionID, stream.Type == "live")
	if err != nil {
		h.logger.Error("failed to generate playlist", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate playlist"})
		return
	}

	// Update session activity
	h.streamService.UpdateSessionActivity(c.Request.Context(), sessionID)

	c.Header("Content-Type", playlist.ContentType)
	c.Header("Cache-Control", "no-cache")
	c.String(http.StatusOK, playlist.Content)
}

func (h *StreamHandler) GetHLSSegment(c *gin.Context) {
	userID := c.GetInt64("user_id")

	streamID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	sessionID := c.Query("session")
	if sessionID == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	// Verify access
	_, err = h.streamService.GetStream(c.Request.Context(), userID, streamID)
	if err != nil {
		c.Status(http.StatusForbidden)
		return
	}

	// Update session activity
	h.streamService.UpdateSessionActivity(c.Request.Context(), sessionID)

	// In production, this would proxy to the actual streaming server
	// For now, return 404
	c.Status(http.StatusNotFound)
}

func (h *StreamHandler) ListCategories(c *gin.Context) {
	// TODO: Implement category listing
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    []interface{}{},
	})
}

func (h *StreamHandler) GetUserInfo(c *gin.Context) {
	userID := c.GetInt64("user_id")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"user_id": userID,
		},
	})
}

func (h *StreamHandler) GetActiveConnections(c *gin.Context) {
	// TODO: Implement active connections retrieval
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"active_connections": 0,
		},
	})
}
