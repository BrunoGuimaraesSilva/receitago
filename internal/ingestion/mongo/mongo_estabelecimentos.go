package ingestion

import (
	"archive/zip"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

func ImportEstabelecimentosZip(ctx context.Context, coll *mongo.Collection, zipPath string) error {
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

			if len(row) < 30 {
				fmt.Printf("⚠️ Skipping row (len=%d): %v\n", len(row), row)
				continue
			}

			doc := map[string]interface{}{
				"cnpj_basico":            getField(row, 0),
				"cnpj_ordem":             getField(row, 1),
				"cnpj_dv":                getField(row, 2),
				"matriz_filial":          getField(row, 3),
				"nome_fantasia":          getField(row, 4),
				"situacao_cadastral":     getField(row, 5),
				"data_situacao":          getField(row, 6),
				"motivo_situacao":        getField(row, 7),
				"nome_cidade_exterior":   getField(row, 8),
				"pais":                   getField(row, 9),
				"data_inicio_atividade":  getField(row, 10),
				"cnae_principal":         getField(row, 11),
				"cnaes_secundarios":      strings.Split(getField(row, 12), ","),
				"tipo_logradouro":        getField(row, 13),
				"logradouro":             getField(row, 14),
				"numero":                 getField(row, 15),
				"complemento":            getField(row, 16),
				"bairro":                 getField(row, 17),
				"cep":                    getField(row, 18),
				"uf":                     getField(row, 19),
				"municipio":              getField(row, 20),
				"ddd1":                   getField(row, 21),
				"ddd2":                   getField(row, 22),
				"telefone1":              getField(row, 23),
				"telefone2":              getField(row, 24),
				"fax1":                   getField(row, 25),
				"fax2":                   getField(row, 26),
				"email":                  getField(row, 27),
				"situacao_especial":      getField(row, 28),
				"data_situacao_especial": getField(row, 29),
				"_imported_at":           time.Now(),
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
