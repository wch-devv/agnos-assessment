package router

import (
	"net/http"

	"agnos-assessment/internal/handler"
	"agnos-assessment/internal/middleware"
	"agnos-assessment/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRouter(
	staffHandler *handler.StaffHandler,
	patientHandler *handler.PatientHandler,
	authService service.AuthService,
) *gin.Engine {
	r := gin.Default()

	// CORS and Recovery Middlewares
	r.Use(gin.Recovery())
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Health check route
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Hospital Middleware Service is running",
		})
	})

	// Helper to register routes on any router group or root
	registerRoutes := func(rg *gin.RouterGroup) {
		// Staff endpoints (Public)
		staff := rg.Group("/staff")
		{
			staff.POST("/create", staffHandler.CreateStaff)
			staff.POST("/login", staffHandler.Login)
		}

		// Patient endpoints (Protected by JWT)
		patient := rg.Group("/patient")
		patient.Use(middleware.AuthMiddleware(authService))
		{
			patient.GET("/search", patientHandler.Search)
		}
	}

	// Register directly on root (e.g. /staff/create, /patient/search as requested in PDF)
	registerRoutes(&r.RouterGroup)

	// Also register under /api/v1 for standard API versioning
	apiV1 := r.Group("/api/v1")
	registerRoutes(apiV1)

	return r
}
