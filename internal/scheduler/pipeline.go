package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/BrunoGuimaraesSilva/receitago/config"
	"github.com/BrunoGuimaraesSilva/receitago/internal/downloader/infra/providers"
	"github.com/BrunoGuimaraesSilva/receitago/internal/downloader/usecase/download"
	postgres "github.com/BrunoGuimaraesSilva/receitago/internal/ingestion/postgres"
	"github.com/BrunoGuimaraesSilva/receitago/pkg/downloader"
	"github.com/BrunoGuimaraesSilva/receitago/pkg/storage"
	"github.com/BrunoGuimaraesSilva/receitago/pkg/storage/local"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
)

type Pipeline struct {
	cfg    *config.Config
	pg     *pgx.Conn
	logger zerolog.Logger
}

func NewPipeline(cfg *config.Config, pg *pgx.Conn, logger zerolog.Logger) (*Pipeline, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if pg == nil {
		return nil, fmt.Errorf("postgres connection is required")
	}
	return &Pipeline{
		cfg:    cfg,
		pg:     pg,
		logger: logger,
	}, nil
}

func (p *Pipeline) Run(ctx context.Context) error {
	start := time.Now()
	p.logger.Info().Msg("🚀 Starting automated pipeline")

	if err := p.downloadReceita(ctx); err != nil {
		return fmt.Errorf("download receita: %w", err)
	}
	if err := p.downloadTesouro(ctx); err != nil {
		return fmt.Errorf("download tesouro: %w", err)
	}
	if err := p.importDictionaries(ctx); err != nil {
		return fmt.Errorf("import dictionaries: %w", err)
	}
	if err := p.importTributario(ctx); err != nil {
		return fmt.Errorf("import tributario: %w", err)
	}

	p.logger.Info().Dur("total_duration", time.Since(start)).Msg("🎉 Pipeline completed successfully")
	return nil
}

func (p *Pipeline) downloadReceita(ctx context.Context) error {
	p.logger.Info().Msg("📥 Step 1/4: Downloading Receita datasets")

	provider := providers.NewReceitaProvider(p.cfg.DataDir+"/receita", p.logger)
	dl := downloader.NewChunkDownloader(downloader.DefaultChunkConfig())
	fs := storage.NewSmartFilestorer(p.cfg.DataDir, local.LocalFS{}, storage.StdZipReader{})

	uc, err := download.NewInteractor(provider, dl, fs, 3, 5*time.Second)
	if err != nil {
		return fmt.Errorf("create interactor: %w", err)
	}
	_, err = uc.Run(ctx)
	return err
}

func (p *Pipeline) downloadTesouro(ctx context.Context) error {
	p.logger.Info().Msg("📥 Step 2/4: Downloading Tesouro datasets")

	provider := providers.NewTesouroProvider(p.cfg.DataDir+"/tesouro", p.logger)
	dl := downloader.NewHTTPDownloader(downloader.DefaultHTTPConfig())
	fs := storage.NewSmartFilestorer(p.cfg.DataDir, local.LocalFS{}, storage.StdZipReader{})

	uc, err := download.NewInteractor(provider, dl, fs, 2, 2*time.Second)
	if err != nil {
		return fmt.Errorf("create interactor: %w", err)
	}
	_, err = uc.Run(ctx)
	return err
}

func (p *Pipeline) importDictionaries(ctx context.Context) error {
	p.logger.Info().Msg("📥 Step 3/4: Importing dictionaries")

	repo := postgres.NewDictionaryRepo(p.pg)
	return postgres.ImportAllDictionaries(ctx, repo, p.cfg.DataDir+"/zips", p.logger)
}

func (p *Pipeline) importTributario(ctx context.Context) error {
	p.logger.Info().Msg("📥 Step 4/4: Importing tributário")

	repo := postgres.NewTributarioRepo(p.pg)
	return postgres.ImportAllRegimes(ctx, repo, p.cfg.DataDir, p.logger)
}
