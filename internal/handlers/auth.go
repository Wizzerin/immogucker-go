package handlers

import (
	"log"
	"net/http"

	"github.com/Wizzerin/immogucker-go/internal/notifier"
	"github.com/Wizzerin/immogucker-go/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Render HTML pages
func (h *API) RenderLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

func (h *API) RenderRegister(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", nil)
}

func (h *API) Login(c *gin.Context) {
	// Use PostForm instead of ShouldBindJSON for HTML form data
	email := c.PostForm("email")
	password := c.PostForm("password")

	if email == "" || password == "" {
		// Return an HTML snippet to be swapped by HTMX into the error container
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte("Email and password are required"))
		return
	}

	user, err := repository.GetUserByEmail(h.DB, email)
	if err != nil {
		c.Data(http.StatusUnauthorized, "text/html; charset=utf-8", []byte("Invalid credentials"))
		return
	}

	log.Printf("Attempting login for: %s", email)
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		c.Data(http.StatusUnauthorized, "text/html; charset=utf-8", []byte("Invalid credentials"))
		return
	}

	sessionID := uuid.New().String()
	if err := repository.CreateSession(h.DB, sessionID, user.ID); err != nil {
		log.Printf("Session creation error: %v", err)
		c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte("Internal server error"))
		return
	}

	// Set the session cookie
	c.SetCookie("session_id", sessionID, 86400, "/", "", false, true)

	// Instruct HTMX to perform a client-side redirect to the dashboard
	c.Header("HX-Redirect", "/")
	c.Status(http.StatusOK)
}

func (h *API) Register(c *gin.Context) {
	username := c.PostForm("username")
	email := c.PostForm("email")
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirm_password")

	if username == "" || email == "" || password == "" {
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte("Email and password are required"))
		return
	}

	if password != confirmPassword {
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte("Passwords do not match"))
		return
	}

	// Hash the provided password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte("Failed to process password"))
		return
	}

	verificationToken := uuid.New().String()

	// Note: Make sure CreateUser is implemented in your repository
	userID, err := repository.CreateUser(h.DB, username, email, string(hashedPassword), verificationToken)
	if err != nil {
		if err.Error() == "username_taken" {
			c.Data(http.StatusConflict, "text/html; charset=utf-8", []byte("Username is already taken"))
			return
		}
		if err.Error() == "email_taken" {
			c.Data(http.StatusConflict, "text/html; charset=utf-8", []byte("Email is already registered"))
			return
		}
		c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte("Server error during registration"))
		return
	}

	go func() {
		if err := notifier.SendVerificationEmail(email, username, verificationToken); err != nil {
			log.Printf("Failed to send verification email to %s: %v", email, err)
		}
	}()

	// Automatically authenticate the user after successful registration
	sessionID := uuid.New().String()
	if err := repository.CreateSession(h.DB, sessionID, userID); err != nil {
		c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte("User created, but login failed"))
		return
	}

	c.SetCookie("session_id", sessionID, 86400, "/", "", false, true)
	c.Header("HX-Redirect", "/")
	c.Status(http.StatusOK)
}

func (h *API) VerifyEmailHandler(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.String(http.StatusBadRequest, "Invalid or already used verification link.")
		return
	}

	err := repository.VerifyEmail(h.DB, token)
	if err != nil {
		log.Printf("[Auth] Verification failed for token %s: %v", token, err)
		c.String(http.StatusBadRequest, "Invalid or already used verification link.")
		return
	}

	c.HTML(http.StatusOK, "login.html", gin.H{
		"Message": "Email successfully verified! You can now use the scraper. Please log in if you haven't already.",
	})
}

func (h *API) Logout(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err == nil && sessionID != "" {
		repository.DeleteSession(h.DB, sessionID)
	}

	c.SetCookie("session_id", "", -1, "/", "", false, true)

	if c.GetHeader("HX-Request") == "true" {
		c.Header("HX-Redirect", "/login")
		c.Status(http.StatusOK)
		return
	}

	c.Redirect(http.StatusFound, "/login")
}
