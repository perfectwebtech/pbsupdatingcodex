package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iptv-platform/streaming-gateway/internal/service"
	"go.uber.org/zap"
)

type ResellerHandler struct {
	resellerService *service.ResellerService
	logger          *zap.Logger
}

func NewResellerHandler(resellerService *service.ResellerService, logger *zap.Logger) *ResellerHandler {
	return &ResellerHandler{
		resellerService: resellerService,
		logger:          logger,
	}
}

// CreateResellerRequest represents the request to create a reseller
type CreateResellerRequest struct {
	UserID              int64   `json:"user_id" binding:"required"`
	ParentID            *int64  `json:"parent_id"`
	Credits             float64 `json:"credits" binding:"min=0"`
	CommissionRate      float64 `json:"commission_rate" binding:"min=0,max=100"`
	CanCreateResellers  bool    `json:"can_create_resellers"`
	MaxUsers            int     `json:"max_users" binding:"min=0"`
	MaxResellers        int     `json:"max_resellers" binding:"min=0"`
	Notes               string  `json:"notes"`
}

// AddCreditsRequest represents the request to add credits to a reseller
type AddCreditsRequest struct {
	Amount      float64 `json:"amount" binding:"required,min=0"`
	Description string  `json:"description"`
}

// ListResellers returns all resellers with pagination
func (h *ResellerHandler) ListResellers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")
	parentID := c.Query("parent_id")

	filters := make(map[string]interface{})
	if search != "" {
		filters["search"] = search
	}
	if parentID != "" {
		if id, err := strconv.ParseInt(parentID, 10, 64); err == nil {
			filters["parent_id"] = id
		}
	}

	resellers, total, err := h.resellerService.ListResellers(c.Request.Context(), page, limit, filters)
	if err != nil {
		h.logger.Error("failed to list resellers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to retrieve resellers",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"resellers": resellers,
			"pagination": gin.H{
				"page":  page,
				"limit": limit,
				"total": total,
				"pages": (total + limit - 1) / limit,
			},
		},
	})
}

// GetReseller returns a single reseller by ID
func (h *ResellerHandler) GetReseller(c *gin.Context) {
	resellerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid reseller ID",
		})
		return
	}

	reseller, err := h.resellerService.GetResellerByID(c.Request.Context(), resellerID)
	if err != nil {
		if err == service.ErrResellerNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "reseller not found",
			})
			return
		}

		h.logger.Error("failed to get reseller", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to retrieve reseller",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    reseller,
	})
}

// CreateReseller creates a new reseller
func (h *ResellerHandler) CreateReseller(c *gin.Context) {
	var req CreateResellerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Get admin user ID from context
	adminUserID := c.GetInt64("user_id")

	reseller, err := h.resellerService.CreateReseller(c.Request.Context(), &service.CreateResellerRequest{
		UserID:             req.UserID,
		ParentID:           req.ParentID,
		Credits:            req.Credits,
		CommissionRate:     req.CommissionRate,
		CanCreateResellers: req.CanCreateResellers,
		MaxUsers:           req.MaxUsers,
		MaxResellers:       req.MaxResellers,
		Notes:              req.Notes,
		CreatedBy:          adminUserID,
	})

	if err != nil {
		switch err {
		case service.ErrUserAlreadyReseller:
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "user is already a reseller",
			})
		case service.ErrParentResellerNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "parent reseller not found",
			})
		default:
			h.logger.Error("failed to create reseller", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "failed to create reseller",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    reseller,
		"message": "reseller created successfully",
	})
}

// UpdateReseller updates an existing reseller
func (h *ResellerHandler) UpdateReseller(c *gin.Context) {
	resellerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid reseller ID",
		})
		return
	}

	var req CreateResellerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	err = h.resellerService.UpdateReseller(c.Request.Context(), resellerID, &service.UpdateResellerRequest{
		CommissionRate:     req.CommissionRate,
		CanCreateResellers: req.CanCreateResellers,
		MaxUsers:           req.MaxUsers,
		MaxResellers:       req.MaxResellers,
		Notes:              req.Notes,
	})

	if err != nil {
		if err == service.ErrResellerNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "reseller not found",
			})
			return
		}

		h.logger.Error("failed to update reseller", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to update reseller",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "reseller updated successfully",
	})
}

// DeleteReseller soft deletes a reseller
func (h *ResellerHandler) DeleteReseller(c *gin.Context) {
	resellerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid reseller ID",
		})
		return
	}

	err = h.resellerService.DeleteReseller(c.Request.Context(), resellerID)
	if err != nil {
		if err == service.ErrResellerNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "reseller not found",
			})
			return
		}

		h.logger.Error("failed to delete reseller", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to delete reseller",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "reseller deleted successfully",
	})
}

// AddCredits adds credits to a reseller
func (h *ResellerHandler) AddCredits(c *gin.Context) {
	resellerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid reseller ID",
		})
		return
	}

	var req AddCreditsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	adminUserID := c.GetInt64("user_id")

	newBalance, err := h.resellerService.AddCredits(c.Request.Context(), resellerID, req.Amount, req.Description, adminUserID)
	if err != nil {
		if err == service.ErrResellerNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "reseller not found",
			})
			return
		}

		h.logger.Error("failed to add credits", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to add credits",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"new_balance": newBalance,
		},
		"message": "credits added successfully",
	})
}

// DeductCredits deducts credits from a reseller
func (h *ResellerHandler) DeductCredits(c *gin.Context) {
	resellerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid reseller ID",
		})
		return
	}

	var req AddCreditsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	adminUserID := c.GetInt64("user_id")

	newBalance, err := h.resellerService.DeductCredits(c.Request.Context(), resellerID, req.Amount, req.Description, adminUserID)
	if err != nil {
		switch err {
		case service.ErrResellerNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "reseller not found",
			})
		case service.ErrInsufficientCredits:
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "insufficient credits",
			})
		default:
			h.logger.Error("failed to deduct credits", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "failed to deduct credits",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"new_balance": newBalance,
		},
		"message": "credits deducted successfully",
	})
}

// GetResellerCustomers returns all customers assigned to a reseller
func (h *ResellerHandler) GetResellerCustomers(c *gin.Context) {
	resellerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid reseller ID",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	customers, total, err := h.resellerService.GetResellerCustomers(c.Request.Context(), resellerID, page, limit)
	if err != nil {
		h.logger.Error("failed to get reseller customers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to retrieve customers",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"customers": customers,
			"pagination": gin.H{
				"page":  page,
				"limit": limit,
				"total": total,
				"pages": (total + limit - 1) / limit,
			},
		},
	})
}

// GetResellerStats returns statistics for a reseller
func (h *ResellerHandler) GetResellerStats(c *gin.Context) {
	resellerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid reseller ID",
		})
		return
	}

	stats, err := h.resellerService.GetResellerStats(c.Request.Context(), resellerID)
	if err != nil {
		h.logger.Error("failed to get reseller stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to retrieve statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetCreditHistory returns credit transaction history
func (h *ResellerHandler) GetCreditHistory(c *gin.Context) {
	resellerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid reseller ID",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	history, total, err := h.resellerService.GetCreditHistory(c.Request.Context(), resellerID, page, limit)
	if err != nil {
		h.logger.Error("failed to get credit history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to retrieve credit history",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"history": history,
			"pagination": gin.H{
				"page":  page,
				"limit": limit,
				"total": total,
				"pages": (total + limit - 1) / limit,
			},
		},
	})
}

// AssignCustomer assigns a user to a reseller
func (h *ResellerHandler) AssignCustomer(c *gin.Context) {
	resellerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid reseller ID",
		})
		return
	}

	var req struct {
		UserID    int64      `json:"user_id" binding:"required"`
		ExpiresAt *time.Time `json:"expires_at"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	err = h.resellerService.AssignCustomer(c.Request.Context(), resellerID, req.UserID, req.ExpiresAt)
	if err != nil {
		switch err {
		case service.ErrResellerNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "reseller not found",
			})
		case service.ErrResellerMaxUsersReached:
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "maximum users limit reached",
			})
		case service.ErrUserAlreadyAssigned:
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "user already assigned to a reseller",
			})
		default:
			h.logger.Error("failed to assign customer", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "failed to assign customer",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "customer assigned successfully",
	})
}

// UnassignCustomer removes a user from a reseller
func (h *ResellerHandler) UnassignCustomer(c *gin.Context) {
	resellerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid reseller ID",
		})
		return
	}

	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid user ID",
		})
		return
	}

	err = h.resellerService.UnassignCustomer(c.Request.Context(), resellerID, userID)
	if err != nil {
		h.logger.Error("failed to unassign customer", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to unassign customer",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "customer unassigned successfully",
	})
}
