package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/BrunoGuimaraesSilva/receitago/config"
	"github.com/BrunoGuimaraesSilva/receitago/internal/api"
)

func main() {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	ctx := context.Background()
	cfg := config.Load()

	// DB connections
	pgConn, err := pgx.Connect(ctx, cfg.PostgresDSN)
	if err != nil {
		logger.Fatal().Err(err).Msg("❌ postgres connect failed")
	}
	defer pgConn.Close(ctx)

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		logger.Fatal().Err(err).Msg("❌ mongo connect failed")
	}
	defer mongoClient.Disconnect(ctx)

	srv := api.NewServer(cfg, logger, pgConn, mongoClient)

	go srv.Run()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctxShutdown)
}
