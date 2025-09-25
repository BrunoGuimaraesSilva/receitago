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

func ImportEmpresasZip(ctx context.Context, coll *mongo.Collection, zipPath string) error {
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

		for {
			row, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("read row: %w", err)
			}

			if len(row) < 7 {
				fmt.Printf("⚠️ Skipping row (len=%d): %v\n", len(row), row)
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
				if err := insertBatch(ctx, coll, batch); err != nil {
					return err
				}
				fmt.Printf("✅ Inserted %d docs into %s (running total: %d)\n", len(batch), coll.Name(), total)
				batch = batch[:0]
			}
		}

		if len(batch) > 0 {
			if err := insertBatch(ctx, coll, batch); err != nil {
				return err
			}
			fmt.Printf("✅ Inserted %d docs into %s (final batch, total: %d)\n", len(batch), coll.Name(), total)
		}

		fmt.Printf("🎯 Finished %s → total inserted: %d\n", f.Name, total)
	}
	return nil
}
