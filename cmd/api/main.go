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
	"github.com/Wizzerin/immogucker-go/internal/repository"
	"github.com/Wizzerin/immogucker-go/internal/worker"
	"github.com/gin-gonic/gin"

	_ "github.com/Wizzerin/immogucker-go/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Immogucker API
// @version         1.0
// @description     Asynchronous REST API for scraping real estate listings from WG-Gesucht.
// @contact.name    Roman Mishyn
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	log.Println("Запуск сервиса Immogucker...")

	config.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("Критическая ошибка: переменная DATABASE_URL не задана")
	}

	db, err := repository.InitDB(dbURL)
	if err != nil {
		log.Fatalf("Критическая ошибка: %v", err)
	}
	defer db.Close()

	// Создаем буферизованный канал на 100 задач
	taskChan := make(chan string, 100)
	var wg sync.WaitGroup // Создаем WaitGroup

	// Запускаем Worker Pool (например, 3 параллельных воркера)
	worker.StartPool(db, taskChan, 3, &wg)

	apiDeps := &handlers.API{
		DB:       db,
		TaskChan: taskChan,
	}

	// 2. Настройка роутера Gin
	router := gin.Default()

	// --- ВОТ ЭТА СТРОКА ВКЛЮЧАЕТ ИНТЕРФЕЙС SWAGGER ---
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Группа маршрутов API
	api := router.Group("/api/v1")
	{
		api.POST("/tasks", apiDeps.CreateTask)
		api.GET("/tasks/:id", apiDeps.GetTaskStatus)

	}

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		log.Println("Сервер запущен на http://localhost:8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Ошибка сервера: %v", err)
		}
	}()

	// --- БЛОК GRACEFUL SHUTDOWN ---

	// Создаем канал для прослушивания сигналов ОС (Ctrl+C, Docker stop
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Блокируем main, пока не придет сигнал

	log.Println("Получен сигнал завершения. Начинаем Graceful Shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Принудительное завершение сервера: %v", err)
	}

	// 2. Закрываем канал задач. Это даст сигнал воркерам выйти из цикла 'for range'
	log.Println("Закрываем очередь задач...")
	close(taskChan)
	wg.Wait()

	log.Println("Все воркеры завершили работу. Соединение с БД закрыто. Сервис остановлен.")
}
