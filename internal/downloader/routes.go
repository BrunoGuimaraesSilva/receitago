package download

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/BrunoGuimaraesSilva/receitago/config"
	"github.com/BrunoGuimaraesSilva/receitago/internal/downloader/infra/providers"
	"github.com/BrunoGuimaraesSilva/receitago/internal/downloader/usecase/download"
	"github.com/BrunoGuimaraesSilva/receitago/pkg/downloader"
	"github.com/BrunoGuimaraesSilva/receitago/pkg/storage"
	"github.com/BrunoGuimaraesSilva/receitago/pkg/storage/local"
)

func RegisterRoutes(r chi.Router, cfg *config.Config, logger zerolog.Logger) {

	// @Summary Download Receita Federal datasets
	// @Description Downloads CNPJ datasets from Receita Federal with optional filters
	// @Tags download
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Param year query int false "Filter by year" example(2024)
	// @Param type query string false "Dataset type" Enums(empresas, estabelecimentos, socios)
	// @Success 200 {array} download.Result "Successfully downloaded datasets"
	// @Failure 400 {object} models.BadRequestResponse "Invalid parameters"
	// @Failure 401 {object} models.UnauthorizedResponse "Missing or invalid token"
	// @Failure 404 {object} models.NotFoundResponse "Dataset not found"
	// @Failure 500 {object} models.ErrorResponse "Internal server error"
	// @Router /v1/download/receita [get]
	r.Get("/download/receita", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		provider := providers.NewReceitaProvider(cfg.DataDir+"/receita", logger)
		dl := downloader.NewChunkDownloader(downloader.DefaultChunkConfig())
		fs := storage.NewSmartFilestorer(cfg.DataDir, local.LocalFS{}, storage.StdZipReader{})

		uc, err := download.NewInteractor(provider, dl, fs, 3, 5*time.Second)
		if err != nil {
			http.Error(w, fmt.Sprintf("create interactor: %v", err), http.StatusInternalServerError)
			return
		}
		results, err := uc.Run(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(results)
	})

	// @Summary Download Tesouro Nacional datasets
	// @Description Downloads tax regime datasets from Tesouro Nacional with optional filters
	// @Tags download
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Param year query int false "Filter by year" example(2024)
	// @Success 200 {array} download.Result "Successfully downloaded datasets"
	// @Failure 400 {object} models.BadRequestResponse "Invalid parameters"
	// @Failure 401 {object} models.UnauthorizedResponse "Missing or invalid token"
	// @Failure 404 {object} models.NotFoundResponse "Dataset not found"
	// @Failure 500 {object} models.ErrorResponse "Internal server error"
	// @Router /v1/download/tesouro [get]
	r.Get("/download/tesouro", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		provider := providers.NewTesouroProvider(cfg.DataDir+"/tesouro", logger)
		dl := downloader.NewHTTPDownloader(downloader.DefaultHTTPConfig())
		fs := storage.NewSmartFilestorer(cfg.DataDir, local.LocalFS{}, storage.StdZipReader{})

		uc, err := download.NewInteractor(provider, dl, fs, 2, 2*time.Second)
		if err != nil {
			http.Error(w, fmt.Sprintf("create interactor: %v", err), http.StatusInternalServerError)
			return
		}
		results, err := uc.Run(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(results)
	})
}
