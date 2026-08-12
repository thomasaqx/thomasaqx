package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/thomasaqx/finance-app/internal/event/kafka"
	grpcserver "github.com/thomasaqx/finance-app/internal/handler/grpc"
	"github.com/thomasaqx/finance-app/internal/handler/rest"
	"github.com/thomasaqx/finance-app/internal/repository/postgres"
	redisrepo "github.com/thomasaqx/finance-app/internal/repository/redis"
	"github.com/thomasaqx/finance-app/internal/service"
	"github.com/thomasaqx/finance-app/pkg/config"
)

func main() {
	cfg := config.Load()

	// --- Database ---
	db, err := postgres.Connect(cfg.Postgres.DSN)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer db.Close()

	// --- Redis ---
	cache := redisrepo.NewCache(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	defer cache.Close()

	// --- Kafka Producer ---
	brokers := strings.Split(cfg.Kafka.Brokers, ",")
	producer, err := kafka.NewProducer(brokers)
	if err != nil {
		log.Printf("kafka producer unavailable: %v (continuing without events)", err)
		producer = nil
	}

	var publisher interface {
		Publish(ctx context.Context, topic string, event interface{}) error
	}
	if producer != nil {
		publisher = producer
		defer producer.Close()
	} else {
		publisher = &noopPublisher{}
	}

	// --- Repositories ---
	accountRepo := postgres.NewAccountRepo(db)
	txRepo := postgres.NewTransactionRepo(db)
	catRepo := postgres.NewCategoryRepo(db)
	budgetRepo := postgres.NewBudgetRepo(db)

	// --- Services ---
	accountSvc := service.NewAccountService(accountRepo, cache, publisher)
	txSvc := service.NewTransactionService(txRepo, accountRepo, cache, publisher, 5)
	defer txSvc.Stop()
	catSvc := service.NewCategoryService(catRepo, budgetRepo)

	// --- Kafka Consumer (logs events) ---
	if producer != nil {
		consumer, err := kafka.NewConsumer(brokers, "finance-app", []string{"transaction.created", "account.created"}, logEvent)
		if err != nil {
			log.Printf("kafka consumer unavailable: %v", err)
		} else {
			ctx, cancelConsumer := context.WithCancel(context.Background())
			consumer.Start(ctx)
			defer cancelConsumer()
			defer consumer.Close()
		}
	}

	// --- gRPC Server (non-blocking) ---
	grpcSrv := grpcserver.NewFinanceGRPCServer(accountSvc, txSvc)
	go func() {
		addr := fmt.Sprintf(":%s", cfg.GRPC.Port)
		log.Printf("gRPC server listening on %s", addr)
		if err := grpcSrv.Start(addr); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()
	defer grpcSrv.Stop()

	// --- HTTP Router ---
	router, err := rest.NewRouter(accountSvc, txSvc, catSvc)
	if err != nil {
		log.Fatalf("build router: %v", err)
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("HTTP server listening on :%s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server: %v", err)
		}
	}()

	// --- Graceful Shutdown ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}
	log.Println("server stopped")
}

// noopPublisher discards all events when Kafka is unavailable.
type noopPublisher struct{}

func (n *noopPublisher) Publish(_ context.Context, topic string, _ interface{}) error {
	log.Printf("[noop] event on topic %q dropped", topic)
	return nil
}

func logEvent(_ context.Context, topic string, payload json.RawMessage) error {
	log.Printf("[kafka] topic=%s payload=%s", topic, string(payload))
	return nil
}
