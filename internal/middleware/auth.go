package middleware

import (
	"database/sql"
	"net/http"

	"github.com/Wizzerin/immogucker-go/internal/repository"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("session_id")
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		userID, err := repository.GetUserIDBySession(db, sessionID)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		c.Set("userID", userID)
		c.Next()
	}
}
