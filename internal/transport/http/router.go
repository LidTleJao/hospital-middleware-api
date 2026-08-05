// Package http registers the HTTP routes the service exposes.
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/LidTleJao/hospital-middleware-api/internal/transport/http/handler"
)

// Handlers collects the handlers the router needs. Grouping them keeps
// NewRouter's signature stable as endpoints are added.
type Handlers struct {
	Staff *handler.Staff
}

// NewRouter builds the engine with every route registered.
func NewRouter(h Handlers) *gin.Engine {
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

	return router
}
