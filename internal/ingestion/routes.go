package ingestion

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/BrunoGuimaraesSilva/receitago/config"
	postgres "github.com/BrunoGuimaraesSilva/receitago/internal/ingestion/postgres"
	"github.com/BrunoGuimaraesSilva/receitago/internal/maestro"
	"github.com/BrunoGuimaraesSilva/receitago/pkg/httputil"
)

func RegisterRoutes(r chi.Router, pg *pgx.Conn, mongo *mongo.Client, cfg *config.Config) {
	r.Post("/import/maestro", func(w http.ResponseWriter, r *http.Request) {
		if err := maestro.NewOrchestrator(pg, mongo).Run(r.Context(), cfg.DataDir); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, err)
			return
		}
		httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "success", "message": "Maestro imports completed"})
	})

	r.Post("/import/dictionaries", func(w http.ResponseWriter, r *http.Request) {
		repo := postgres.NewDictionaryRepo(pg)
		if err := postgres.ImportAllDictionaries(r.Context(), repo, cfg.DataDir+"/zips"); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, err)
			return
		}
		httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "success", "message": "Dictionaries imported"})
	})

	r.Post("/import/tributario", func(w http.ResponseWriter, r *http.Request) {
		repo := postgres.NewTributarioRepo(pg)
		if err := postgres.ImportAllRegimes(r.Context(), repo, cfg.DataDir); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, err)
			return
		}
		httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "success", "message": "Tributário imported"})
	})
}
