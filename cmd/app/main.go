package main

/*
main — точка входа приложения.

Здесь происходит:
- инициализация хранилища;
- создание сервисного слоя;
- создание HTTP обработчиков;
- регистрация маршрутов Gin;
- запуск HTTP сервера;
- graceful shutdown приложения.

Graceful shutdown нужен,
чтобы сервис корректно завершал работу:
- переставал принимать новые запросы;
- дожидался завершения текущих запросов;
- корректно освобождал ресурсы.
*/

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rabbitmq/amqp091-go"

	"rwb-contest/internal/config"
	"rwb-contest/internal/handlers"
	"rwb-contest/internal/metrics"
	"rwb-contest/internal/rabbitmq"
	"rwb-contest/internal/service"
	"rwb-contest/internal/storage"
	"rwb-contest/internal/worker"
)

func main() {

	cfg := config.Load()
	metrics.Init()

	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Инициализация in-memory хранилища.
	storage := storage.NewRingBufferStorage()

	// Инициализация сервисного слоя.
	trendsService := service.NewTrendsService(storage)

	// Инициализация HTTP обработчиков.
	trendsHandler := handlers.NewTrendsHandler(trendsService)
	stopWordsHandler := handlers.NewStopWordsHandler(trendsService)
	healthHandler := handlers.NewHealthHandler()

	// Внутренний канал задач для worker pool.
	jobs := make(chan amqp091.Delivery, 50)

	consumer := rabbitmq.NewConsumer(
		cfg.RabbitURL,
		cfg.RabbitQueue,
	)

	if err := consumer.Connect(); err != nil {
		log.Fatalf("RabbitMQ connection error: %v", err)
	}
	defer consumer.Close()

	workerPool := worker.NewPool(
		cfg.WorkersCount,
		trendsService,
	)

	go workerPool.Start(appCtx, jobs)

	go func() {
		if err := consumer.Consume(appCtx, jobs); err != nil {
			log.Printf("RabbitMQ consume error: %v", err)
			cancel()
		}
	}()

	go func() {
		<-appCtx.Done()
		close(jobs)
	}()

	// Создание Gin роутера.
	router := gin.Default()

	// Healthcheck endpoint.
	router.GET("/health", healthHandler.Health)

	// Получение Top-N запросов.
	router.GET("/top", trendsHandler.GetTop)

	// Работа со стоп-словами.
	router.GET("/stop-words", stopWordsHandler.GetStopWords)
	router.POST("/stop-words", stopWordsHandler.AddStopWord)
	router.DELETE("/stop-words/:word", stopWordsHandler.RemoveStopWord)

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Конфигурация HTTP сервера.
	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: router,
	}

	// Запускаем HTTP сервер в отдельной горутине.
	go func() {

		log.Printf("Server started on %s", cfg.HTTPAddr)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server starting error: %v", err)
		}
	}()

	// Канал для получения сигналов остановки приложения.
	quit := make(chan os.Signal, 1)

	// Подписываемся на SIGINT и SIGTERM.
	signal.Notify(
		quit,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	// Ожидаем сигнал завершения.
	<-quit

	log.Println("Signal received, server is down.")

	// Останавливаем все горутины приложения.
	cancel()

	// Контекст с timeout для graceful shutdown.
	ctx, shutdownCancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer shutdownCancel()

	// Корректное завершение HTTP сервера.
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Graceful shutdown error: %v", err)
	}

	log.Println("Server correctly stopped.")
}
