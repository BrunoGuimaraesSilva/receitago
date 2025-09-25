package ingestion

import (
	"context"
	"strings"

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
func insertBatch(ctx context.Context, coll *mongo.Collection, batch []interface{}) error {
	opts := options.InsertMany().SetOrdered(false)
	_, err := coll.InsertMany(ctx, batch, opts)
	if err == nil {
		// log progress
		collName := coll.Name()
		println("✅ Inserted", len(batch), "docs into", collName)
	}
	return err
}
