package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	"github.com/BrunoGuimaraesSilva/receitago/internal/downloader/domain/dataset"
	"github.com/BrunoGuimaraesSilva/receitago/internal/downloader/usecase/download"
	"github.com/BrunoGuimaraesSilva/receitago/internal/downloader/usecase/process"
)

type Server struct {
	Downloader dataset.DownloaderPort
	Filestorer dataset.FilestorerPort
	Timeout    time.Duration
	logger     zerolog.Logger
}

func New(d dataset.DownloaderPort, f dataset.FilestorerPort, logger zerolog.Logger) *Server {
	return &Server{
		Downloader: d,
		Filestorer: f,
		Timeout:    15 * time.Minute,
		logger:     logger,
	}
}

func (s *Server) HandleProvider(mux *http.ServeMux, route string, prov dataset.DatasetProvider, maxRetries int, retryDelay time.Duration) {
	mux.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info().Str("route", route).Str("method", r.Method).Msg("Handling provider request")

		ctx, cancel := context.WithTimeout(r.Context(), s.Timeout)
		defer cancel()

		uc, err := download.NewInteractor(prov, s.Downloader, s.Filestorer, maxRetries, retryDelay)
		if err != nil {
			http.Error(w, fmt.Sprintf("create interactor: %v", err), http.StatusInternalServerError)
			return
		}
		res, err := uc.Run(ctx)
		if err != nil {
			s.logger.Error().Err(err).Str("route", route).Msg("Provider request failed")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		s.logger.Info().Str("route", route).Msg("Provider request completed successfully")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	})
}

func (s *Server) HandleWorkflow(mux *http.ServeMux, route string, steps []process.Step) {
	mux.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info().Str("route", route).Str("method", r.Method).Int("steps", len(steps)).Msg("Handling workflow request")

		ctx, cancel := context.WithTimeout(r.Context(), s.Timeout)
		defer cancel()

		pm := process.NewManager(steps, s.Downloader, s.Filestorer)
		rep, err := pm.Run(ctx)
		status := http.StatusOK
		if err != nil || !rep.Success {
			s.logger.Error().Err(err).Str("route", route).Bool("success", rep.Success).Msg("Workflow failed")
			status = http.StatusBadGateway
		} else {
			s.logger.Info().Str("route", route).Msg("Workflow completed successfully")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(rep)
	})
}
