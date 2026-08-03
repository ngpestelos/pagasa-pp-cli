// Copyright 2026 Nestor G Pestelos Jr and contributors. Licensed under Apache-2.0. See LICENSE.

package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestAWSObs_UpsertIdempotentAndUpdate(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	ctx := context.Background()
	temp := 25.0
	row := AWSObsRow{
		StationID: "98", StationName: "Science Garden", ObservedAt: "2026-08-03T00:00:00Z",
		CapturedAt: "2026-08-03T00:01:00Z", TempC: &temp,
	}
	if err := s.UpsertAWSObs(ctx, row); err != nil {
		t.Fatal(err)
	}
	temp2 := 26.5
	row.TempC = &temp2
	row.CapturedAt = "2026-08-03T00:05:00Z"
	if err := s.UpsertAWSObs(ctx, row); err != nil {
		t.Fatal(err)
	}
	list, err := s.ListAWSObs(ctx, "98", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("want 1 row after conflict update, got %d", len(list))
	}
	if list[0].TempC == nil || *list[0].TempC != 26.5 {
		t.Errorf("temp = %v, want 26.5", list[0].TempC)
	}
	if list[0].CapturedAt != "2026-08-03T00:05:00Z" {
		t.Errorf("captured_at = %s", list[0].CapturedAt)
	}
}

func TestAWSObs_BatchSkipMissingObservedAt(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	temp := 1.0
	w, skip, err := s.UpsertAWSObsBatch(ctx, []AWSObsRow{
		{StationID: "1", ObservedAt: "2026-08-03T01:00:00Z", TempC: &temp},
		{StationID: "2", ObservedAt: ""}, // skip
		{StationID: "", ObservedAt: "2026-08-03T01:00:00Z"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if w != 1 || skip != 2 {
		t.Fatalf("written=%d skipped=%d, want 1/2", w, skip)
	}
}

func TestAWSObs_Prune(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	old := time.Now().UTC().Add(-20 * 24 * time.Hour).Format(time.RFC3339)
	fresh := time.Now().UTC().Format(time.RFC3339)
	_ = s.UpsertAWSObs(ctx, AWSObsRow{StationID: "1", ObservedAt: old, CapturedAt: old})
	_ = s.UpsertAWSObs(ctx, AWSObsRow{StationID: "1", ObservedAt: fresh, CapturedAt: fresh})
	n, err := s.PruneAWSObs(ctx, 14*24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("pruned %d, want 1", n)
	}
	list, err := s.ListAWSObs(ctx, "1", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ObservedAt != fresh {
		t.Fatalf("remaining = %+v", list)
	}
}

func TestAWSObs_ListMissingTableEmpty(t *testing.T) {
	// RO open of empty file without migrate: use a fresh RW open then... actually
	// migrateExtras always runs on Open. Simulate by querying after DROP.
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.DB().Exec(`DROP TABLE aws_obs`); err != nil {
		t.Fatal(err)
	}
	list, err := s.ListAWSObs(context.Background(), "", 5)
	if err != nil {
		t.Fatalf("missing table should return empty, not error: %v", err)
	}
	if list != nil && len(list) != 0 {
		t.Fatalf("want empty, got %d", len(list))
	}
	s.Close()
}

func TestAWSObs_SchemaVersionUnchanged(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	v, err := s.SchemaVersion()
	if err != nil {
		t.Fatal(err)
	}
	if v != StoreSchemaVersion {
		t.Fatalf("schema version = %d, want %d (must not bump for aws_obs)", v, StoreSchemaVersion)
	}
	if StoreSchemaVersion != 9 {
		t.Fatalf("StoreSchemaVersion = %d, plan requires stay at 9", StoreSchemaVersion)
	}
	// Table exists
	var name string
	err = s.DB().QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='aws_obs'`).Scan(&name)
	if err != nil || name != "aws_obs" {
		t.Fatalf("aws_obs missing: %v", err)
	}
}
