package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/burndler/burndler/internal/models"
	"github.com/burndler/burndler/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PackageHandler handles package-related endpoints
type PackageHandler struct {
	packager *services.Packager
	db       *gorm.DB
}

// NewPackageHandler creates a new package handler
func NewPackageHandler(packager *services.Packager, db *gorm.DB) *PackageHandler {
	return &PackageHandler{
		packager: packager,
		db:       db,
	}
}

// Create handles package creation requests (Developer only)
func (h *PackageHandler) Create(c *gin.Context) {
	var req services.PackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Invalid package request",
			"details": err.Error(),
		})
		return
	}

	// Validate input
	if req.Name == "" || req.Compose == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "MISSING_FIELDS",
			"message": "Name and compose content are required",
		})
		return
	}

	// Get user ID from context
	userIDInterface, _ := c.Get("user_id")
	userIDStr, _ := userIDInterface.(string)
	userID, _ := strconv.ParseUint(userIDStr, 10, 32)

	// Create build record
	build := &models.Build{
		Name:        req.Name,
		Status:      "queued",
		UserID:      uint(userID),
		ComposeYAML: req.Compose,
	}

	if err := h.db.Create(build).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "DB_ERROR",
			"message": "Failed to create build record",
			"details": err.Error(),
		})
		return
	}

	// Start async package creation
	go h.processPackage(build, &req)

	c.JSON(http.StatusAccepted, gin.H{
		"build_id": build.ID.String(),
		"status":   build.Status,
	})
}

// Status returns build status
func (h *PackageHandler) Status(c *gin.Context) {
	buildIDStr := c.Param("id")
	buildID, err := uuid.Parse(buildIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_BUILD_ID",
			"message": "Invalid build ID format",
		})
		return
	}

	var build models.Build
	if err := h.db.First(&build, "id = ?", buildID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "BUILD_NOT_FOUND",
				"message": "Build not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "DB_ERROR",
			"message": "Failed to fetch build",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"build_id":     build.ID.String(),
		"status":       build.Status,
		"progress":     build.Progress,
		"download_url": build.DownloadURL,
		"error":        build.Error,
		"created_at":   build.CreatedAt,
		"completed_at": build.CompletedAt,
	})
}

// processPackage handles async package creation
func (h *PackageHandler) processPackage(build *models.Build, req *services.PackageRequest) {
	// Update status to building
	build.Status = "building"
	build.Progress = 10
	h.db.Save(build)

	// Create package
	ctx := context.Background()
	url, err := h.packager.CreatePackage(ctx, req)

	if err != nil {
		// Update build with error
		build.Status = "failed"
		build.Error = err.Error()
		h.db.Save(build)
		return
	}

	// Update build with success
	now := gorm.DeletedAt{}
	build.Status = "completed"
	build.Progress = 100
	build.DownloadURL = url
	build.CompletedAt = &now.Time
	h.db.Save(build)
}
