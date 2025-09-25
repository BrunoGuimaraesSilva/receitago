package ingestion

import (
	"archive/zip"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

func ImportSociosZip(ctx context.Context, coll *mongo.Collection, zipPath string) error {
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

		_, _ = reader.Read() // skip header

		const batchSize = 5000
		var batch []interface{}

		for {
			row, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("read row: %w", err)
			}

			doc := map[string]interface{}{
				"cnpj_basico":                getField(row, 0),
				"identificador_socio":        getField(row, 1),
				"nome_socio":                 getField(row, 2),
				"cnpj_cpf_socio":             getField(row, 3),
				"qualificacao_socio":         getField(row, 4),
				"data_entrada_sociedade":     getField(row, 5),
				"pais":                       getField(row, 6),
				"cpf_representante_legal":    getField(row, 7),
				"nome_representante_legal":   getField(row, 8),
				"qualificacao_representante": getField(row, 9),
				"faixa_etaria":               getField(row, 10),
				"_imported_at":               time.Now(),
			}

			batch = append(batch, doc)
			if len(batch) >= batchSize {
				if err := insertBatch(ctx, coll, batch); err != nil {
					return err
				}
				batch = batch[:0]
			}
		}
		if len(batch) > 0 {
			if err := insertBatch(ctx, coll, batch); err != nil {
				return err
			}
		}
	}
	return nil
}
