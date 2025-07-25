package controller

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"featureflags/pkg/logger"
	"featureflags/service"
	"featureflags/validator"

	"github.com/labstack/echo/v4"
)

type FlagController struct {
	flagService service.FlagService
	logger      *logger.Logger
}

func NewFlagController(fs service.FlagService, log *logger.Logger) *FlagController {
	return &FlagController{
		flagService: fs,
		logger:      log,
	}
}

// CreateFlag handles POST /flags
func (fc *FlagController) CreateFlag(c echo.Context) error {
	var req validator.FlagCreateRequest
	if err := c.Bind(&req); err != nil {
		fc.logger.Warnw("Failed to bind create flag request", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Get actor from context (in a real app, this would come from auth middleware)
	actor := getActorFromContext(c)

	flag, err := fc.flagService.CreateFlag(context.Background(), req, actor)
	if err != nil {
		return fc.handleServiceError(c, err)
	}

	fc.logger.Infow("Flag created via API", "flagID", flag.ID, "name", flag.Name, "actor", actor)
	return c.JSON(http.StatusCreated, flag)
}

// ToggleFlag handles POST /flags/:id/toggle
func (fc *FlagController) ToggleFlag(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid flag ID",
		})
	}

	var req validator.FlagToggleRequest
	if err := c.Bind(&req); err != nil {
		fc.logger.Warnw("Failed to bind toggle flag request", "error", err, "flagID", id)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	actor := getActorFromContext(c)

	err = fc.flagService.ToggleFlag(context.Background(), id, req, actor)
	if err != nil {
		return fc.handleServiceError(c, err)
	}

	status := "disabled"
	if req.Enable {
		status = "enabled"
	}

	fc.logger.Infow("Flag toggled via API", "flagID", id, "status", status, "actor", actor)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Flag " + status + " successfully",
		"flag_id": id,
		"status":  status,
	})
}

// ListFlags handles GET /flags
func (fc *FlagController) ListFlags(c echo.Context) error {
	flags, err := fc.flagService.ListFlags(context.Background())
	if err != nil {
		fc.logger.Errorw("Failed to list flags via API", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve flags",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"flags": flags,
		"count": len(flags),
	})
}

// GetFlag handles GET /flags/:id
func (fc *FlagController) GetFlag(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid flag ID",
		})
	}

	flag, err := fc.flagService.GetFlag(context.Background(), id)
	if err != nil {
		return fc.handleServiceError(c, err)
	}

	return c.JSON(http.StatusOK, flag)
}

// GetFlagAudit handles GET /flags/:id/audit
func (fc *FlagController) GetFlagAudit(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid flag ID",
		})
	}

	logs, err := fc.flagService.GetFlagAuditLogs(context.Background(), id)
	if err != nil {
		return fc.handleServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"audit_logs": logs,
		"count":      len(logs),
	})
}

// handleServiceError converts service errors to appropriate HTTP responses
func (fc *FlagController) handleServiceError(c echo.Context, err error) error {
	// Handle validation errors
	if validationErr, ok := err.(validator.ValidationErrors); ok {
		fc.logger.Warnw("Validation error in API", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":            "Validation failed",
			"validation_errors": validationErr.Errors,
		})
	}

	// Handle dependency errors (matching task requirements)
	if depErr, ok := err.(service.DependencyError); ok {
		fc.logger.Warnw("Dependency error in API", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":                depErr.Message,
			"missing_dependencies": depErr.MissingDependencies,
		})
	}

	// Handle specific service errors
	switch {
	case errors.Is(err, service.ErrFlagNotFound):
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Flag not found",
		})
	case errors.Is(err, service.ErrFlagAlreadyExists):
		return c.JSON(http.StatusConflict, map[string]string{
			"error": "Flag with this name already exists",
		})
	case errors.Is(err, service.ErrCircularDependency):
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Circular dependency detected",
		})
	default:
		fc.logger.Errorw("Internal error in API", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Internal server error",
		})
	}
}

// getActorFromContext extracts the actor from the request context
// In a real application, this would be populated by authentication middleware
func getActorFromContext(c echo.Context) string {
	// Check for actor in headers first
	if actor := c.Request().Header.Get("X-Actor"); actor != "" {
		return actor
	}
	
	// Check for actor in query params
	if actor := c.QueryParam("actor"); actor != "" {
		return actor
	}
	
	// Default to anonymous user
	return "anonymous"
} 