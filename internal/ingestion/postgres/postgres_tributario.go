package ingestion

import (
	"archive/zip"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

type RegimeDTO struct {
	Ano                     int
	CNPJ                    string
	CNPJdaSCP               string
	FormaTributacao         string
	QuantidadeEscrituracoes int
	Dataset                 string
}

type TributarioRepo struct {
	conn *pgx.Conn
	psql sq.StatementBuilderType
}

func NewTributarioRepo(conn *pgx.Conn) *TributarioRepo {
	return &TributarioRepo{
		conn: conn,
		psql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *TributarioRepo) InsertBatch(ctx context.Context, records []RegimeDTO) error {
	batch := &pgx.Batch{}

	for _, rec := range records {
		sql, args, err := r.psql.Insert("tributario.regimes").
			Columns("ano", "cnpj", "cnpj_da_scp", "forma_de_tributacao", "quantidade_de_escrituracoes", "dataset").
			Values(rec.Ano, rec.CNPJ, rec.CNPJdaSCP, rec.FormaTributacao, rec.QuantidadeEscrituracoes, rec.Dataset).
			ToSql()
		if err != nil {
			return err
		}
		batch.Queue(sql, args...)
	}

	br := r.conn.SendBatch(ctx, batch)
	return br.Close()
}

func normalizeCNPJ(raw string) string {
	raw = strings.ReplaceAll(raw, ".", "")
	raw = strings.ReplaceAll(raw, "/", "")
	raw = strings.ReplaceAll(raw, "-", "")
	return fmt.Sprintf("%014s", raw)
}

func ImportRegimeZip(ctx context.Context, repo *TributarioRepo, zipPath, dataset string) error {
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
		reader.Comma = ','
		reader.LazyQuotes = true

		// Skip header
		if _, err := reader.Read(); err != nil {
			return fmt.Errorf("read header: %w", err)
		}

		var batch []RegimeDTO

		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("read row: %w", err)
			}
			if len(record) < 5 {
				continue
			}

			ano, _ := strconv.Atoi(record[0])
			cnpj := normalizeCNPJ(record[1])
			cnpjSCP := record[2]
			forma := strings.TrimSpace(record[3])
			qtd, _ := strconv.Atoi(record[4])

			batch = append(batch, RegimeDTO{
				Ano:                     ano,
				CNPJ:                    cnpj,
				CNPJdaSCP:               cnpjSCP,
				FormaTributacao:         forma,
				QuantidadeEscrituracoes: qtd,
				Dataset:                 dataset,
			})

			if len(batch) >= 10000 {
				if err := repo.InsertBatch(ctx, batch); err != nil {
					return err
				}
				batch = batch[:0]
			}
		}

		if len(batch) > 0 {
			if err := repo.InsertBatch(ctx, batch); err != nil {
				return err
			}
		}
	}

	return nil
}

func ImportAllRegimes(ctx context.Context, repo *TributarioRepo, baseDir string) error {
	files := []struct {
		Name    string
		Dataset string
	}{
		{"Lucro Presumido.zip", "Lucro Presumido"},
		{"Lucro Real.zip", "Lucro Real"},
		{"Lucro Arbitrado.zip", "Lucro Arbitrado"},
		{"Imunes e Isentas.zip", "Imunes e Isentas"},
	}

	for _, f := range files {
		zipPath := filepath.Join(baseDir, f.Name)
		if _, err := os.Stat(zipPath); os.IsNotExist(err) {
			fmt.Printf("⚠️ missing: %s\n", zipPath)
			continue
		}
		fmt.Printf("📥 Importing %s into tributario.regimes...\n", f.Dataset)
		if err := ImportRegimeZip(ctx, repo, zipPath, f.Dataset); err != nil {
			return fmt.Errorf("import %s: %w", f.Dataset, err)
		}
		fmt.Printf("✅ Done %s\n", f.Dataset)
	}
	return nil
}
