// Copyright 2026 Nestor G Pestelos Jr and contributors. Licensed under Apache-2.0. See LICENSE.

package store

import (
	"context"
	"database/sql"
	"fmt"
)

// migrateExtras runs after the generated store migrations and before the
// schema-version stamp. It is the canonical place for novel-feature auxiliary
// tables that need to live in the local store.
//
// Edit this file when adding tables for novel commands. Keep migrations
// idempotent with CREATE TABLE IF NOT EXISTS / CREATE INDEX IF NOT EXISTS so
// every store open can safely re-run them.
func (s *Store) migrateExtras(ctx context.Context, conn *sql.Conn) error {
	// Additive only — do not bump StoreSchemaVersion for aws_obs so older
	// binaries that share data.db keep opening (they ignore unknown tables).
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS aws_obs (
			station_id TEXT NOT NULL,
			observed_at TEXT NOT NULL,
			captured_at TEXT NOT NULL,
			station_name TEXT,
			temp_c REAL,
			humidity_pct REAL,
			wind_kmh REAL,
			wind_dir TEXT,
			precip_mm_hr REAL,
			pressure REAL,
			solar REAL,
			data TEXT,
			PRIMARY KEY (station_id, observed_at)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_aws_obs_station_observed
			ON aws_obs (station_id, observed_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_aws_obs_observed
			ON aws_obs (observed_at)`,
	}
	for _, m := range migrations {
		if _, err := conn.ExecContext(ctx, m); err != nil {
			return fmt.Errorf("extra migration failed: %w", err)
		}
	}
	return nil
}
