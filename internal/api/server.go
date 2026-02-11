package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/BrunoGuimaraesSilva/receitago/config"
	"github.com/BrunoGuimaraesSilva/receitago/internal/api/models"
	download "github.com/BrunoGuimaraesSilva/receitago/internal/downloader"
	"github.com/BrunoGuimaraesSilva/receitago/internal/ingestion"
	"github.com/BrunoGuimaraesSilva/receitago/internal/scheduler"

	_ "github.com/BrunoGuimaraesSilva/receitago/docs" // Swagger docs
)

type Server struct {
	http      *http.Server
	scheduler *scheduler.Scheduler
	logger    zerolog.Logger
}

func NewServer(cfg *config.Config, logger zerolog.Logger, pg *pgx.Conn, mongo *mongo.Client) (*Server, error) {
	r := chi.NewRouter()
	r.Use(
		middleware.RequestID,
		middleware.Logger,
		middleware.Recoverer,
		middleware.Timeout(cfg.Timeout),
	)

	// Swagger UI
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// @Summary Health check
	// @Description Returns API health status and version information
	// @Tags health
	// @Accept json
	// @Produce json
	// @Success 200 {object} models.HealthResponse
	// @Router /health [get]
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response := models.HealthResponse{
			Status:    "healthy",
			Timestamp: time.Now().Format(time.RFC3339),
			Version:   "1.0.0",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// modules
	// v1 API routes
	r.Route("/v1", func(v1 chi.Router) {
		download.RegisterRoutes(v1, cfg, logger)
		ingestion.RegisterRoutes(v1, pg, mongo, cfg, logger)
	})

	_ = chi.Walk(r, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		logger.Info().Msgf("📌 route: %-6s %s", method, route)
		return nil
	})

	pipeline, err := scheduler.NewPipeline(cfg, pg, logger)
	if err != nil {
		return nil, fmt.Errorf("create pipeline: %w", err)
	}

	cronScheduler, err := scheduler.NewScheduler(pipeline, logger)
	if err != nil {
		return nil, fmt.Errorf("create scheduler: %w", err)
	}

	if err := cronScheduler.Start(cfg.CronSchedule); err != nil {
		return nil, fmt.Errorf("start cron scheduler: %w", err)
	}

	return &Server{
		http: &http.Server{
			Addr:         ":" + cfg.ServerPort,
			Handler:      r,
			ReadTimeout:  cfg.Timeout,
			WriteTimeout: cfg.Timeout,
			IdleTimeout:  60 * time.Second,
		},
		scheduler: cronScheduler,
		logger:    logger,
	}, nil
}

func (s *Server) Run() {
	s.logger.Info().Msg("🚀 API listening on " + s.http.Addr)
	if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Error().Err(err).Msg("❌ server error")
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info().Msg("🛑 shutting down server...")
	s.scheduler.Stop()
	return s.http.Shutdown(ctx)
}
