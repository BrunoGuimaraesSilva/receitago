package ingestion

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/BrunoGuimaraesSilva/receitago/config"
	postgres "github.com/BrunoGuimaraesSilva/receitago/internal/ingestion/postgres"
	"github.com/BrunoGuimaraesSilva/receitago/pkg/httputil"
)

func RegisterRoutes(r chi.Router, pg *pgx.Conn, mongo *mongo.Client, cfg *config.Config, logger zerolog.Logger) {
	// @Summary Import tax dictionaries
	// @Description Imports tax-related dictionaries into PostgreSQL
	// @Tags import
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Success 200 {object} models.SuccessResponse "Successfully imported"
	// @Failure 401 {object} models.UnauthorizedResponse "Missing or invalid token"
	// @Failure 500 {object} models.ErrorResponse "Internal server error"
	// @Router /v1/import/dictionaries [post]
	r.Post("/import/dictionaries", func(w http.ResponseWriter, r *http.Request) {
		repo := postgres.NewDictionaryRepo(pg)
		if err := postgres.ImportAllDictionaries(r.Context(), repo, cfg.DataDir+"/zips", logger); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, err)
			return
		}
		httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "success", "message": "Dictionaries imported"})
	})

	// @Summary Import tax regime data
	// @Description Imports tax regime (tributário) data into PostgreSQL
	// @Tags import
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Success 200 {object} models.SuccessResponse "Successfully imported"
	// @Failure 401 {object} models.UnauthorizedResponse "Missing or invalid token"
	// @Failure 500 {object} models.ErrorResponse "Internal server error"
	// @Router /v1/import/tributario [post]
	r.Post("/import/tributario", func(w http.ResponseWriter, r *http.Request) {
		repo := postgres.NewTributarioRepo(pg)
		if err := postgres.ImportAllRegimes(r.Context(), repo, cfg.DataDir, logger); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, err)
			return
		}
		httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "success", "message": "Tributário imported"})
	})
}
