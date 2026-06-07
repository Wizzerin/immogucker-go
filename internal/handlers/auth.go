package handlers

import (
	"log"
	"net/http"

	"github.com/Wizzerin/immogucker-go/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *API) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	user, err := repository.GetUserByEmail(h.DB, req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	log.Printf("Attempting login for: %s", req.Email)
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		log.Printf("Bcrypt error: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	sessionID := uuid.New().String()
	log.Printf("Creating session for UserID: %d", user.ID)
	if err := repository.CreateSession(h.DB, sessionID, user.ID); err != nil {
		log.Printf("Session creation error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	c.SetCookie("session_id", sessionID, 86400, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{"message": "logged in seccessfully"})
}
