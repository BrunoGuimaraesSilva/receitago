package scheduler

import (
	"context"
	"fmt"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
)

type Scheduler struct {
	cron     *cron.Cron
	pipeline *Pipeline
	logger   zerolog.Logger
}

func NewScheduler(pipeline *Pipeline, logger zerolog.Logger) (*Scheduler, error) {
	if pipeline == nil {
		return nil, fmt.Errorf("pipeline is required")
	}
	return &Scheduler{
		cron:     cron.New(),
		pipeline: pipeline,
		logger:   logger,
	}, nil
}

func (s *Scheduler) Start(cronExpr string) error {
	s.logger.Info().Str("schedule", cronExpr).Msg("📅 Starting cron scheduler")

	_, err := s.cron.AddFunc(cronExpr, func() {
		ctx := context.Background()
		s.logger.Info().Msg("⏰ Cron triggered, starting pipeline")

		if err := s.pipeline.Run(ctx); err != nil {
			s.logger.Error().Err(err).Msg("Pipeline failed")
		}
	})

	if err != nil {
		return err
	}

	s.cron.Start()
	return nil
}

func (s *Scheduler) Stop() {
	s.logger.Info().Msg("Stopping cron scheduler")
	s.cron.Stop()
}
