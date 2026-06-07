package middleware

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"github.com/Wizzerin/immogucker-go/internal/repository"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("session_id")

		if err != nil || sessionID == "" {
			log.Printf("❌ AuthMiddleware: Cookie 'session_id' not found or empty. Error: %v", err)
			handleUnauthorized(c)
			return
		}

		userID, err := repository.GetUserIDBySession(db, sessionID)
		if err != nil {
			log.Printf("❌ AuthMiddleware: Session %s invalid/not found in DB. Error: %v", sessionID, err)
			handleUnauthorized(c)
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}

func handleUnauthorized(c *gin.Context) {
	// Return JSON if the request targets the API (e.g., mobile apps, external services)
	if strings.HasPrefix(c.Request.URL.Path, "/api/") {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Handle HTMX requests from the UI by forcing a client-side redirect
	if c.GetHeader("HX-Request") == "true" {
		c.Header("HX-Redirect", "/login")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// Standard browser request fallback (e.g., user directly types the URL)
	c.Redirect(http.StatusFound, "/login")
	c.Abort()
}
