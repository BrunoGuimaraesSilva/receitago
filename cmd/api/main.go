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

// @title ReceitaGo API
// @version 1.0
// @description API para download e ingestão de dados públicos brasileiros (Receita Federal e Tesouro Nacional)

// @contact.name API Support
// @contact.url https://github.com/BrunoGuimaraesSilva/receitago

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token

func main() {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	ctx := context.Background()
	cfg := config.Load()

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

	srv, err := api.NewServer(cfg, logger, pgConn, mongoClient)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create server")
	}

	go srv.Run()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctxShutdown)
}
