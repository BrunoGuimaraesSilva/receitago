package maestro

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	ingestion "github.com/BrunoGuimaraesSilva/receitago/internal/ingestion/mongo"
	"github.com/jackc/pgx/v5"
	"go.mongodb.org/mongo-driver/mongo"
)

type Orchestrator struct {
	pg    *pgx.Conn
	mongo *mongo.Client
}

func NewOrchestrator(pg *pgx.Conn, mongo *mongo.Client) *Orchestrator {
	return &Orchestrator{pg: pg, mongo: mongo}
}

// helper: import all files for a prefix (Empresas, Estabelecimentos, Socios)
func (o *Orchestrator) importMultiple(
	ctx context.Context,
	coll *mongo.Collection,
	baseDir, prefix string,
	importFunc func(context.Context, *mongo.Collection, string) error,
) error {
	pattern := filepath.Join(baseDir, prefix+"*.zip")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("glob %s: %w", pattern, err)
	}
	if len(files) == 0 {
		fmt.Printf("⚠️ No %s files found in %s\n", prefix, baseDir)
		return nil
	}

	for _, file := range files {
		fmt.Printf("📥 Importing %s...\n", filepath.Base(file))
		if err := importFunc(ctx, coll, file); err != nil {
			return fmt.Errorf("import %s: %w", file, err)
		}
	}
	return nil
}

func (o *Orchestrator) Run(ctx context.Context, baseDir string) error {
	start := time.Now()
	db := o.mongo.Database("receita")

	fmt.Println("📥 Importing Empresas, Estabelecimentos, Sócios into Mongo...")

	if err := o.importMultiple(ctx, db.Collection("empresas"), baseDir, "Empresas", ingestion.ImportEmpresasZip); err != nil {
		return err
	}

	if err := o.importMultiple(ctx, db.Collection("estabelecimentos"), baseDir, "Estabelecimentos", ingestion.ImportEstabelecimentosZip); err != nil {
		return err
	}

	if err := o.importMultiple(ctx, db.Collection("socios"), baseDir, "Socios", ingestion.ImportSociosZip); err != nil {
		return err
	}

	fmt.Println("✅ Big datasets imported")
	fmt.Printf("🎉 All imports finished in %s\n", time.Since(start))
	return nil
}
