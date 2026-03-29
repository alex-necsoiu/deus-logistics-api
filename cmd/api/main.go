package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/alex-necsoiu/deus-logistics-api/internal/config"
	"github.com/alex-necsoiu/deus-logistics-api/internal/events"
	"github.com/alex-necsoiu/deus-logistics-api/internal/postgres"
	"github.com/alex-necsoiu/deus-logistics-api/internal/service"
	transporthttp "github.com/alex-necsoiu/deus-logistics-api/internal/transport/http"
)

func main() {
	// 1. Load .env file if it exists
	_ = godotenv.Load()

	// 2. Initialize logger
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logger := zerolog.New(os.Stderr).With().Timestamp().Caller().Logger()
	log.Logger = logger

	ctx := context.Background()
	log.Info().Msg("=== DEUS Logistics API ===")
	log.Info().Msg("starting up...")

	// 3. Load configuration from environment variables
	cfg := config.LoadFromEnv()
	log.Info().
		Str("db_host", cfg.DBHost).
		Int("db_port", cfg.DBPort).
		Str("db_name", cfg.DBName).
		Str("environment", cfg.ServerEnv).
		Int("server_port", cfg.ServerPort).
		Msg("configuration loaded")

	// 4. Connect to PostgreSQL
	log.Info().Msg("connecting to PostgreSQL...")
	pool, err := postgres.New(ctx, cfg.DSN())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer postgres.Close(pool)
	log.Info().Msg("✓ connected to PostgreSQL")

	// 5. Verify database health
	if err := postgres.HealthCheck(ctx, pool); err != nil {
		log.Fatal().Err(err).Msg("database health check failed")
	}
	log.Info().Msg("✓ database health check passed")

	// 6. Create repositories
	cargoRepo := postgres.NewCargoRepository(pool)
	vesselRepo := postgres.NewVesselRepository(pool)
	trackingRepo := postgres.NewTrackingRepository(pool)
	eventRepo := postgres.NewEventRepository(pool)
	log.Info().Msg("✓ repositories initialized")

	// 7. Create Kafka producer
	log.Info().Msg("initializing Kafka producer...")
	publisher := events.NewEventPublisher(cfg.KafkaBrokers, cfg.KafkaTopicEvents)
	defer publisher.Close()
	log.Info().Str("brokers", fmt.Sprintf("%v", cfg.KafkaBrokers)).Str("topic", cfg.KafkaTopicEvents).Msg("✓ Kafka producer ready")

	// 8. Create domain services
	cargoSvc := service.NewCargoService(cargoRepo, publisher, trackingRepo)
	vesselSvc := service.NewVesselService(vesselRepo)
	trackingSvc := service.NewTrackingService(trackingRepo)
	log.Info().Msg("✓ domain services initialized")

	// 9. Create and start Kafka consumer
	log.Info().Msg("initializing Kafka consumer...")
	consumer := events.NewEventConsumer(cfg.KafkaBrokers, cfg.KafkaTopicEvents, "deus-api-consumer", eventRepo)
	consumer.Start(context.Background())
	defer consumer.Stop()
	log.Info().Msg("✓ Kafka consumer started")

	// 10. Setup HTTP router
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	// Register all routes
	transporthttp.Router(engine, cargoSvc, vesselSvc, trackingSvc)
	log.Info().Msg("✓ HTTP routes registered")

	// 11. Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.ServerPort),
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 12. Start server in background goroutine
	serverDone := make(chan error, 1)
	go func() {
		log.Info().Int("port", cfg.ServerPort).Msg("starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverDone <- err
		}
	}()

	// 13. Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-sigChan:
		log.Info().Msg("shutdown signal received")
	case err := <-serverDone:
		log.Error().Err(err).Msg("server error - initiating shutdown")
	}

	// 14. Graceful shutdown with timeout
	log.Info().Msg("gracefully shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("server shutdown error")
	}

	// 15. Final cleanup
	log.Info().Msg("✓ DEUS Logistics API stopped")
}
