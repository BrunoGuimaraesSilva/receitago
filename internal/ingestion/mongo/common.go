package ingestion

import (
	"context"
	"strings"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// safe field getter
func getField(row []string, idx int) string {
	if idx < len(row) {
		return strings.TrimSpace(row[idx])
	}
	return ""
}

// batch insert with logging
func insertBatch(ctx context.Context, coll *mongo.Collection, batch []interface{}, logger zerolog.Logger) error {
	opts := options.InsertMany().SetOrdered(false)
	_, err := coll.InsertMany(ctx, batch, opts)
	if err == nil {
		logger.Debug().Int("count", len(batch)).Str("collection", coll.Name()).Msg("✅ Inserted batch")
	}
	return err
}
