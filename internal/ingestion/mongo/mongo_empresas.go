package ingestion

import (
	"archive/zip"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
)

func ImportEmpresasZip(ctx context.Context, coll *mongo.Collection, zipPath string, logger zerolog.Logger) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("open file inside zip: %w", err)
		}
		defer rc.Close()

		reader := csv.NewReader(rc)
		reader.Comma = ';'
		reader.LazyQuotes = true

		const batchSize = 5000
		var batch []interface{}
		total := 0

		logger.Info().Str("file", f.Name).Msg("Processing file inside zip")

		for {
			row, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("read row: %w", err)
			}

			if len(row) < 7 {
				logger.Warn().Int("len", len(row)).Str("row_sample", fmt.Sprintf("%v", row)).Msg("⚠️ Skipping malformed row")
				continue
			}

			doc := map[string]interface{}{
				"cnpj_basico":          getField(row, 0),
				"razao_social":         getField(row, 1),
				"natureza_juridica":    getField(row, 2),
				"qualificacao_resp":    getField(row, 3),
				"capital_social":       getField(row, 4),
				"porte_empresa":        getField(row, 5),
				"ente_federativo_resp": getField(row, 6),
				"_imported_at":         time.Now(),
			}

			batch = append(batch, doc)
			total++

			if len(batch) >= batchSize {
				if err := insertBatch(ctx, coll, batch, logger); err != nil {
					return err
				}
				batch = batch[:0]
			}
		}

		if len(batch) > 0 {
			if err := insertBatch(ctx, coll, batch, logger); err != nil {
				return err
			}
		}

		logger.Info().Str("file", f.Name).Int("total", total).Msg("🎯 Finished processing file")
	}
	return nil
}
