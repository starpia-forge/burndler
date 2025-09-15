package handlers

import (
	"net/http"

	"github.com/burndler/burndler/internal/services"
	"github.com/gin-gonic/gin"
)

// ComposeHandler handles compose-related endpoints
type ComposeHandler struct {
	merger *services.Merger
	linter *services.Linter
}

// NewComposeHandler creates a new compose handler
func NewComposeHandler(merger *services.Merger, linter *services.Linter) *ComposeHandler {
	return &ComposeHandler{
		merger: merger,
		linter: linter,
	}
}

// Merge handles compose merge requests
func (h *ComposeHandler) Merge(c *gin.Context) {
	var req services.MergeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Invalid merge request",
			"details": err.Error(),
		})
		return
	}

	// Validate input
	if len(req.Modules) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "NO_MODULES",
			"message": "At least one module is required",
		})
		return
	}

	// Perform merge
	result, err := h.merger.Merge(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "MERGE_FAILED",
			"message": "Failed to merge compose files",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Lint handles compose lint requests
func (h *ComposeHandler) Lint(c *gin.Context) {
	var req services.LintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "Invalid lint request",
			"details": err.Error(),
		})
		return
	}

	// Validate input
	if req.Compose == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "NO_COMPOSE",
			"message": "Compose content is required",
		})
		return
	}

	// Perform lint
	result, err := h.linter.Lint(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "LINT_FAILED",
			"message": "Failed to lint compose file",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}
