package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/BrunoGuimaraesSilva/receitago/services/downloader/domain/dataset"
	"github.com/BrunoGuimaraesSilva/receitago/services/downloader/usecase/download"
	"github.com/BrunoGuimaraesSilva/receitago/services/downloader/usecase/process"
)

type Server struct {
	Downloader dataset.DownloaderPort
	Filestorer dataset.FilestorerPort
	Timeout    time.Duration
}

func New(d dataset.DownloaderPort, f dataset.FilestorerPort) *Server {
	return &Server{
		Downloader: d,
		Filestorer: f,
		Timeout:    15 * time.Minute,
	}
}

func (s *Server) HandleProvider(mux *http.ServeMux, route string, prov dataset.DatasetProvider, maxRetries int, retryDelay time.Duration) {
	mux.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), s.Timeout)
		defer cancel()
		uc := download.NewInteractor(prov, s.Downloader, s.Filestorer, maxRetries, retryDelay)
		res, err := uc.Run(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	})
}

func (s *Server) HandleWorkflow(mux *http.ServeMux, route string, steps []process.Step) {
	mux.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), s.Timeout)
		defer cancel()
		pm := process.NewManager(steps, s.Downloader, s.Filestorer)
		rep, err := pm.Run(ctx)
		status := http.StatusOK
		if err != nil || !rep.Success {
			status = http.StatusBadGateway
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(rep)
	})
}
