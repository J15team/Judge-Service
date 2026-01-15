package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"judge-service/internal/config"
	"judge-service/internal/domain/entity"
	"judge-service/internal/infrastructure/sandbox"
)

// Handler handles HTTP requests
type Handler struct {
	config  *config.Config
	sandbox *sandbox.DockerPoolSandbox
}

// NewHandler creates a new Handler
func NewHandler(cfg *config.Config) *Handler {
	return &Handler{
		config:  cfg,
		sandbox: sandbox.NewDockerPoolSandbox(cfg.Sandbox.CompileTimeout, 10),
	}
}

// JudgeRequest represents the judge API request
type JudgeRequest struct {
	Code        string            `json:"code" binding:"required"`
	Language    string            `json:"language" binding:"required"`
	TestCases   []entity.TestCase `json:"testCases" binding:"required"`
	TimeLimit   int               `json:"timeLimit"`
	MemoryLimit int               `json:"memoryLimit"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status             string   `json:"status"`
	SupportedLanguages []string `json:"supportedLanguages"`
}

// Judge handles POST /api/judge
func (h *Handler) Judge(c *gin.Context) {
	var req JudgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Apply defaults and limits
	timeLimit := req.TimeLimit
	if timeLimit <= 0 {
		timeLimit = h.config.Sandbox.DefaultTimeLimit
	}
	if timeLimit > h.config.Sandbox.MaxTimeLimit {
		timeLimit = h.config.Sandbox.MaxTimeLimit
	}

	memoryLimit := req.MemoryLimit
	if memoryLimit <= 0 {
		memoryLimit = h.config.Sandbox.DefaultMemoryLimit
	}
	if memoryLimit > h.config.Sandbox.MaxMemoryLimit {
		memoryLimit = h.config.Sandbox.MaxMemoryLimit
	}

	// Execute in Docker sandbox
	results := h.sandbox.CompileAndRun(req.Code, req.Language, req.TestCases, timeLimit, memoryLimit)

	c.JSON(http.StatusOK, entity.JudgeResponse{Results: results})
}

// Health handles GET /api/health
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:             "healthy",
		SupportedLanguages: []string{"c"},
	})
}
