package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rishirajmaheshwari/4-in-a-row/internal/analytics"
	"github.com/rishirajmaheshwari/4-in-a-row/internal/config"
	"github.com/rishirajmaheshwari/4-in-a-row/internal/game"
	"github.com/rishirajmaheshwari/4-in-a-row/internal/server"
	"github.com/rishirajmaheshwari/4-in-a-row/internal/storage"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	repo, err := storage.NewRepository(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer repo.Close()

	producer := analytics.NewProducer(cfg.KafkaBrokers, "game-analytics")
	defer producer.Close()

	manager := game.NewManager()
	srv := server.New(cfg, manager, repo, producer)

	httpServer := &http.Server{Addr: ":" + cfg.Port, Handler: srv.Routes()}

	go func() {
		log.Printf("listening on :%s", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down")
	_ = httpServer.Shutdown(context.Background())
}
