// Package middleware holds the gin handlers that run before a route.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/LidTleJao/hospital-middleware-api/internal/service"
)

// claimsKey is the gin context key holding the verified claims.
const claimsKey = "claims"

// tokenParser is the slice of the token service this middleware uses.
type tokenParser interface {
	Parse(raw string) (service.Claims, error)
}

// RequireAuth rejects any request without a valid bearer token and stores the
// verified claims for the handler behind it.
func RequireAuth(tokens tokenParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		raw, found := strings.CutPrefix(header, "Bearer ")
		if !found || raw == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		claims, err := tokens.Parse(raw)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set(claimsKey, claims)
		c.Next()
	}
}

// ClaimsFrom returns the claims stored by RequireAuth.
func ClaimsFrom(c *gin.Context) (service.Claims, bool) {
	value, exists := c.Get(claimsKey)
	if !exists {
		return service.Claims{}, false
	}
	claims, ok := value.(service.Claims)
	return claims, ok
}
