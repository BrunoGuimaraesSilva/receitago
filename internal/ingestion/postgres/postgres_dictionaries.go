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
	"github.com/rs/zerolog"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type DictionaryDTO struct {
	ID          int64
	Description string
}

type DictionaryRepo struct {
	conn *pgx.Conn
	psql sq.StatementBuilderType
}

func NewDictionaryRepo(conn *pgx.Conn) *DictionaryRepo {
	return &DictionaryRepo{
		conn: conn,
		psql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *DictionaryRepo) InsertBatch(ctx context.Context, table string, records []DictionaryDTO) error {
	batch := &pgx.Batch{}

	for _, rec := range records {
		sql, args, err := r.psql.Insert(table).
			Columns("id", "description").
			Values(rec.ID, rec.Description).
			Suffix("ON CONFLICT (id) DO UPDATE SET description = EXCLUDED.description").
			ToSql()
		if err != nil {
			return err
		}
		batch.Queue(sql, args...)
	}

	br := r.conn.SendBatch(ctx, batch)
	return br.Close()
}

func normalizeID(raw string) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, fmt.Errorf("empty id")
	}
	prefixed := "1" + raw
	return strconv.ParseInt(prefixed, 10, 64)
}

func ImportDictionaryZip(ctx context.Context, repo *DictionaryRepo, zipPath, table string, logger zerolog.Logger) error {
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

		utf8Reader := transform.NewReader(rc, charmap.ISO8859_1.NewDecoder())

		reader := csv.NewReader(utf8Reader)
		reader.Comma = ';'
		reader.LazyQuotes = true

		var batch []DictionaryDTO

		for {
			row, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("read row: %w", err)
			}
			if len(row) < 2 {
				continue
			}

			id, err := normalizeID(row[0])
			if err != nil {
				logger.Warn().Str("id", row[0]).Msg("Skipping invalid id")
				continue
			}

			desc := strings.TrimSpace(row[1])

			batch = append(batch, DictionaryDTO{ID: id, Description: desc})

			if len(batch) >= 5000 {
				if err := repo.InsertBatch(ctx, table, batch); err != nil {
					return err
				}
				batch = batch[:0]
			}
		}

		if len(batch) > 0 {
			if err := repo.InsertBatch(ctx, table, batch); err != nil {
				return err
			}
		}
	}

	return nil
}

func ImportAllDictionaries(ctx context.Context, repo *DictionaryRepo, baseDir string, logger zerolog.Logger) error {
	files := []struct {
		Name  string
		Table string
	}{
		{"Cnaes.zip", "dictionaries.cnaes"},
		{"Motivos.zip", "dictionaries.motivos"},
		{"Qualificacoes.zip", "dictionaries.qualificacoes"},
		{"Municipios.zip", "dictionaries.municipios"},
		{"Paises.zip", "dictionaries.paises"},
		{"Naturezas.zip", "dictionaries.naturezas"},
	}

	cwd, _ := os.Getwd()
	logger.Debug().Str("cwd", cwd).Msg("Current working directory")

	for _, f := range files {
		zipPath := filepath.Join(baseDir, f.Name)
		if _, err := os.Stat(zipPath); os.IsNotExist(err) {
			logger.Warn().Str("file", zipPath).Msg("Dictionary file missing")
			continue
		}

		logger.Info().Str("file", f.Name).Str("table", f.Table).Msg("Importing dictionary")
		if err := ImportDictionaryZip(ctx, repo, zipPath, f.Table, logger); err != nil {
			return fmt.Errorf("import %s: %w", f.Table, err)
		}
		logger.Info().Str("table", f.Table).Msg("Dictionary import completed")
	}
	return nil
}
