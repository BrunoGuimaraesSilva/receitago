package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/BrunoGuimaraesSilva/receitago/services/downloader/infra/downloader"
	"github.com/BrunoGuimaraesSilva/receitago/services/downloader/infra/providers"
	"github.com/BrunoGuimaraesSilva/receitago/services/downloader/infra/storage"
	"github.com/BrunoGuimaraesSilva/receitago/services/downloader/usecase/download"
)

func main() {
	mux := http.NewServeMux()

	// Receita route
	mux.HandleFunc("/download/receita", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		provider := providers.NewReceitaProvider("./data/meta")
		dl := downloader.NewChunkDownloader(10*time.Minute, 50, 3, true)
		fs := storage.NewSmartFilestorer("./data")

		uc := download.NewInteractor(provider, dl, fs, 3, 5*time.Second)

		results, err := uc.Run(ctx)
		if err != nil {
			http.Error(w, fmt.Sprintf("download error: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(results)
	})

	// Tesouro route
	mux.HandleFunc("/download/tesouro", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		provider := providers.NewTesouroProvider("./data/meta")
		dl := downloader.NewHTTPDownloader(2*time.Minute, "")
		fs := storage.NewSmartFilestorer("./data")

		uc := download.NewInteractor(provider, dl, fs, 2, 2*time.Second)

		results, err := uc.Run(ctx)
		if err != nil {
			http.Error(w, fmt.Sprintf("download error: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(results)
	})

	addr := ":8080"
	log.Printf("🚀 Server running at %s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
