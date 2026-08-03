// Copyright 2026 Nestor G Pestelos Jr and contributors. Licensed under Apache-2.0. See LICENSE.

package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// DefaultAWSObsRetention is how long station observation rows are kept.
// Prune runs only on capture paths, not on live read-only obs.
const DefaultAWSObsRetention = 14 * 24 * time.Hour

// AWSObsRow is one stored automated weather station observation.
type AWSObsRow struct {
	StationID   string          `json:"station_id"`
	StationName string          `json:"station_name,omitempty"`
	ObservedAt  string          `json:"observed_at"`  // RFC3339 UTC
	CapturedAt  string          `json:"captured_at"`  // RFC3339 UTC
	TempC       *float64        `json:"temp_c,omitempty"`
	HumidityPct *float64        `json:"humidity_pct,omitempty"`
	WindKmh     *float64        `json:"wind_kmh,omitempty"`
	WindDir     string          `json:"wind_dir,omitempty"`
	PrecipMmHr  *float64        `json:"precip_mm_hr,omitempty"`
	Pressure    *float64        `json:"pressure,omitempty"`
	Solar       *float64        `json:"solar,omitempty"`
	Data        json.RawMessage `json:"data,omitempty"`
}

// UpsertAWSObs inserts or updates one observation keyed by (station_id, observed_at).
// ON CONFLICT updates metrics and captured_at so corrected re-scrapes win.
func (s *Store) UpsertAWSObs(ctx context.Context, row AWSObsRow) error {
	if strings.TrimSpace(row.StationID) == "" || strings.TrimSpace(row.ObservedAt) == "" {
		return fmt.Errorf("aws_obs upsert: station_id and observed_at are required")
	}
	if row.CapturedAt == "" {
		row.CapturedAt = time.Now().UTC().Format(time.RFC3339)
	}
	data := row.Data
	if len(data) == 0 {
		b, err := json.Marshal(row)
		if err != nil {
			return err
		}
		data = b
	}

	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO aws_obs (
			station_id, observed_at, captured_at, station_name,
			temp_c, humidity_pct, wind_kmh, wind_dir, precip_mm_hr, pressure, solar, data
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(station_id, observed_at) DO UPDATE SET
			captured_at = excluded.captured_at,
			station_name = excluded.station_name,
			temp_c = excluded.temp_c,
			humidity_pct = excluded.humidity_pct,
			wind_kmh = excluded.wind_kmh,
			wind_dir = excluded.wind_dir,
			precip_mm_hr = excluded.precip_mm_hr,
			pressure = excluded.pressure,
			solar = excluded.solar,
			data = excluded.data
	`,
		row.StationID, row.ObservedAt, row.CapturedAt, row.StationName,
		nullFloat(row.TempC), nullFloat(row.HumidityPct), nullFloat(row.WindKmh),
		row.WindDir, nullFloat(row.PrecipMmHr), nullFloat(row.Pressure), nullFloat(row.Solar),
		string(data),
	)
	if err != nil {
		return fmt.Errorf("upsert aws_obs: %w", err)
	}
	return nil
}

// UpsertAWSObsBatch upserts many rows in one transaction. Rows missing
// station_id or observed_at are skipped (counted in skipped).
func (s *Store) UpsertAWSObsBatch(ctx context.Context, rows []AWSObsRow) (written, skipped int, err error) {
	if len(rows) == 0 {
		return 0, 0, nil
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO aws_obs (
			station_id, observed_at, captured_at, station_name,
			temp_c, humidity_pct, wind_kmh, wind_dir, precip_mm_hr, pressure, solar, data
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(station_id, observed_at) DO UPDATE SET
			captured_at = excluded.captured_at,
			station_name = excluded.station_name,
			temp_c = excluded.temp_c,
			humidity_pct = excluded.humidity_pct,
			wind_kmh = excluded.wind_kmh,
			wind_dir = excluded.wind_dir,
			precip_mm_hr = excluded.precip_mm_hr,
			pressure = excluded.pressure,
			solar = excluded.solar,
			data = excluded.data
	`)
	if err != nil {
		return 0, 0, err
	}
	defer stmt.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	for _, row := range rows {
		if strings.TrimSpace(row.StationID) == "" || strings.TrimSpace(row.ObservedAt) == "" {
			skipped++
			continue
		}
		if row.CapturedAt == "" {
			row.CapturedAt = now
		}
		data := row.Data
		if len(data) == 0 {
			b, mErr := json.Marshal(row)
			if mErr != nil {
				return written, skipped, mErr
			}
			data = b
		}
		if _, err := stmt.ExecContext(ctx,
			row.StationID, row.ObservedAt, row.CapturedAt, row.StationName,
			nullFloat(row.TempC), nullFloat(row.HumidityPct), nullFloat(row.WindKmh),
			row.WindDir, nullFloat(row.PrecipMmHr), nullFloat(row.Pressure), nullFloat(row.Solar),
			string(data),
		); err != nil {
			return written, skipped, fmt.Errorf("upsert aws_obs batch: %w", err)
		}
		written++
	}
	if err := tx.Commit(); err != nil {
		return written, skipped, err
	}
	return written, skipped, nil
}

// ListAWSObs returns observations newest-first. stationFilter matches station_id
// exactly or station_name case-insensitive substring when non-empty.
// Missing table (pre-migrate RO open) returns empty slice, not error.
func (s *Store) ListAWSObs(ctx context.Context, stationFilter string, limit int) ([]AWSObsRow, error) {
	if limit <= 0 {
		limit = 20
	}
	q := `SELECT station_id, observed_at, captured_at, COALESCE(station_name, ''),
		temp_c, humidity_pct, wind_kmh, COALESCE(wind_dir, ''), precip_mm_hr, pressure, solar, data
		FROM aws_obs`
	args := []any{}
	if f := strings.TrimSpace(stationFilter); f != "" {
		q += ` WHERE station_id = ? OR LOWER(station_name) LIKE ?`
		args = append(args, f, "%"+strings.ToLower(escapeLike(f))+"%")
	}
	q += ` ORDER BY observed_at DESC LIMIT ?`
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		if isMissingTable(err, "aws_obs") {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	var out []AWSObsRow
	for rows.Next() {
		var r AWSObsRow
		var temp, hum, wind, precip, pressure, solar sql.NullFloat64
		var data sql.NullString
		if err := rows.Scan(
			&r.StationID, &r.ObservedAt, &r.CapturedAt, &r.StationName,
			&temp, &hum, &wind, &r.WindDir, &precip, &pressure, &solar, &data,
		); err != nil {
			return nil, err
		}
		r.TempC = nullFloatPtr(temp)
		r.HumidityPct = nullFloatPtr(hum)
		r.WindKmh = nullFloatPtr(wind)
		r.PrecipMmHr = nullFloatPtr(precip)
		r.Pressure = nullFloatPtr(pressure)
		r.Solar = nullFloatPtr(solar)
		if data.Valid {
			r.Data = json.RawMessage(data.String)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// PruneAWSObs deletes rows older than maxAge by observed_at. Non-positive maxAge
// uses DefaultAWSObsRetention. Returns rows deleted.
func (s *Store) PruneAWSObs(ctx context.Context, maxAge time.Duration) (int64, error) {
	if maxAge <= 0 {
		maxAge = DefaultAWSObsRetention
	}
	cutoff := time.Now().UTC().Add(-maxAge).Format(time.RFC3339)

	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	res, err := s.db.ExecContext(ctx, `DELETE FROM aws_obs WHERE observed_at < ?`, cutoff)
	if err != nil {
		if isMissingTable(err, "aws_obs") {
			return 0, nil
		}
		return 0, fmt.Errorf("prune aws_obs: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func nullFloat(p *float64) any {
	if p == nil {
		return nil
	}
	return *p
}

func nullFloatPtr(n sql.NullFloat64) *float64 {
	if !n.Valid {
		return nil
	}
	v := n.Float64
	return &v
}

func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}

func isMissingTable(err error, table string) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "no such table") && strings.Contains(msg, strings.ToLower(table))
}
