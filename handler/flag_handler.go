package handler

import (
	"featureflags/config"
	"featureflags/controller"
	_ "featureflags/docs" // Import for swagger docs
	"featureflags/pkg/logger"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func RegisterRoutes(e *echo.Echo, fc *controller.FlagController, cfg *config.Config, log *logger.Logger) {
	// Add middleware
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogMethod: true,
		LogError:  true,
		LogValuesFunc: func(c echo.Context, values middleware.RequestLoggerValues) error {
			if values.Error != nil {
				log.Errorw("Request failed",
					"method", values.Method,
					"uri", values.URI,
					"status", values.Status,
					"error", values.Error,
				)
			} else {
				log.Infow("Request completed",
					"method", values.Method,
					"uri", values.URI,
					"status", values.Status,
				)
			}
			return nil
		},
	}))
	
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "healthy",
			"service": "featureflags",
		})
	})

	// Swagger documentation (if enabled)
	if cfg.Swagger.Enabled {
		log.Infow("Swagger documentation enabled", "path", "/swagger/*")
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}

	// API routes
	api := e.Group("/api/v1")
	
	// Flag routes
	api.POST("/flags", fc.CreateFlag)
	api.POST("/flags/:id/toggle", fc.ToggleFlag)
	api.GET("/flags", fc.ListFlags)
	api.GET("/flags/:id", fc.GetFlag)
	api.GET("/flags/:id/audit", fc.GetFlagAudit)
} 