package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Wizzerin/immogucker-go/internal/config"
	"github.com/Wizzerin/immogucker-go/internal/handlers"
	"github.com/Wizzerin/immogucker-go/internal/middleware"
	"github.com/Wizzerin/immogucker-go/internal/repository"
	"github.com/Wizzerin/immogucker-go/internal/worker"
	"github.com/gin-gonic/gin"

	_ "github.com/Wizzerin/immogucker-go/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Immogucker API
// @version         1.1
// @description     Asynchronous REST API for scraping real estate listings from WG-Gesucht and Kleinanzeigen.
// @contact.name    Roman Mishyn
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	log.Println("Starting Immogucker service...")

	config.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("Critical error: DATABASE_URL environment variable is not set")
	}

	db, err := repository.InitDB(dbURL)
	if err != nil {
		log.Fatalf("Critical database error: %v", err)
	}
	defer db.Close()

	// Initialize a buffered channel for up to 100 tasks
	taskChan := make(chan string, 100)
	var wg sync.WaitGroup

	// Start the Worker Pool
	worker.StartPool(db, taskChan, 3, &wg)

	apiDeps := &handlers.API{
		DB:       db,
		TaskChan: taskChan,
	}

	// Configure Gin router
	router := gin.Default()

	// Load HTML templates for HTMX Dashboard
	router.LoadHTMLGlob("web/templates/*")

	// --- ENABLE SWAGGER UI ---
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// --- API ROUTES ---
	auth := router.Group("/api/v1/auth")
	{
		auth.POST("/login", apiDeps.Login)
	}

	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(db))
	{
		protected.POST("/tasks", apiDeps.CreateTask)
		protected.GET("/tasks/:id", apiDeps.GetTaskStatus)
		protected.GET("/health", apiDeps.HealthCheck)
	}

	// --- UI DASHBOARD ROUTES ---
	ui := router.Group("/")
	{
		ui.GET("/", apiDeps.RenderDashboard)
		ui.POST("/ui/tasks", apiDeps.HandleSubmitTask)
		ui.GET("/ui/tasks/:id/status", apiDeps.HandleTaskStatus)
	}

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Run the server in a separate goroutine to avoid blocking the main thread
	go func() {
		log.Println("Server is running on http://localhost:8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// --- GRACEFUL SHUTDOWN BLOCK ---

	// Create a channel to listen for OS signals (Ctrl+C, Docker stop)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Block main thread until a signal is received

	log.Println("Shutdown signal received. Initiating Graceful Shutdown...")

	// 1. Stop accepting new HTTP requests (allow 5 seconds to finish current ones)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Forced server shutdown: %v", err)
	}

	// 2. Close the task channel to signal workers to exit the 'for range' loop
	log.Println("Closing task queue...")
	close(taskChan)
	wg.Wait()

	log.Println("All workers have finished. Database connection closed. Service stopped.")
}
