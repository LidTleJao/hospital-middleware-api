// Package http registers the HTTP routes the service exposes.
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/LidTleJao/hospital-middleware-api/internal/service"
	"github.com/LidTleJao/hospital-middleware-api/internal/transport/http/handler"
	"github.com/LidTleJao/hospital-middleware-api/internal/transport/http/middleware"
)

// Handlers collects the handlers the router needs. Grouping them keeps
// NewRouter's signature stable as endpoints are added.
type Handlers struct {
	Staff   *handler.Staff
	Patient *handler.Patient
}

// TokenParser verifies the bearer token guarding the patient routes.
type TokenParser interface {
	Parse(raw string) (service.Claims, error)
}

// NewRouter builds the engine with every route registered.
func NewRouter(h Handlers, tokens TokenParser) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Liveness probe used by Docker and by nginx to confirm the app is up.
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	staff := router.Group("/staff")
	{
		staff.POST("/create", h.Staff.Create)
		staff.POST("/login", h.Staff.Login)
	}

	// Everything below requires a valid token. The hospital a caller may read
	// comes from that token, never from the request body.
	patient := router.Group("/patient", middleware.RequireAuth(tokens))
	{
		patient.POST("/search", h.Patient.Search)
		patient.POST("/import", h.Patient.Import)
	}

	return router
}
