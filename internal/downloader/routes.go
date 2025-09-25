package download

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/BrunoGuimaraesSilva/receitago/config"
	"github.com/BrunoGuimaraesSilva/receitago/internal/downloader/infra/providers"
	"github.com/BrunoGuimaraesSilva/receitago/internal/downloader/usecase/download"
	"github.com/BrunoGuimaraesSilva/receitago/pkg/downloader"
	"github.com/BrunoGuimaraesSilva/receitago/pkg/storage"
	"github.com/BrunoGuimaraesSilva/receitago/pkg/storage/local"
)

func RegisterRoutes(r chi.Router, cfg *config.Config) {

	r.Get("/download/receita", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		provider := providers.NewReceitaProvider("./data/meta")
		dl := downloader.NewChunkDownloader(10*time.Minute, 50, 3, true)
		fs := storage.NewSmartFilestorer("./data", local.LocalFS{}, storage.StdZipReader{})

		uc := download.NewInteractor(provider, dl, fs, 3, 5*time.Second)
		results, err := uc.Run(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(results)
	})

	r.Get("/download/tesouro", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		provider := providers.NewTesouroProvider("./data/meta")
		dl := downloader.NewHTTPDownloader(2*time.Minute, "")
		fs := storage.NewSmartFilestorer("./data", local.LocalFS{}, storage.StdZipReader{})

		uc := download.NewInteractor(provider, dl, fs, 2, 2*time.Second)
		results, err := uc.Run(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(results)
	})
}
