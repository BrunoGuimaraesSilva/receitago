package api

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/BrunoGuimaraesSilva/receitago/config"
	download "github.com/BrunoGuimaraesSilva/receitago/internal/downloader"
	"github.com/BrunoGuimaraesSilva/receitago/internal/ingestion"
)

type Server struct {
	http   *http.Server
	logger zerolog.Logger
}

func NewServer(cfg *config.Config, logger zerolog.Logger, pg *pgx.Conn, mongo *mongo.Client) *Server {
	r := chi.NewRouter()
	r.Use(
		middleware.RequestID,
		middleware.Logger,
		middleware.Recoverer,
		middleware.Timeout(cfg.Timeout),
	)

	// health
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	// modules
	ingestion.RegisterRoutes(r, pg, mongo, cfg)
	download.RegisterRoutes(r, cfg)

	// walk all routes and log them
	_ = chi.Walk(r, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		logger.Info().Msgf("📌 route: %-6s %s", method, route)
		return nil
	})

	return &Server{
		http: &http.Server{
			Addr:         ":" + cfg.ServerPort,
			Handler:      r,
			ReadTimeout:  cfg.Timeout,
			WriteTimeout: cfg.Timeout,
			IdleTimeout:  60 * time.Second,
		},
		logger: logger,
	}
}

func (s *Server) Run() {
	s.logger.Info().Msg("🚀 API listening on " + s.http.Addr)
	if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Error().Err(err).Msg("❌ server error")
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info().Msg("🛑 shutting down server...")
	return s.http.Shutdown(ctx)
}
