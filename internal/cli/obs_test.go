// Copyright 2026 Nestor G Pestelos Jr and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"strings"
	"testing"

	"github.com/ngpestelos/pagasa-pp-cli/internal/pagasa"
)

func TestFilterAWSStations(t *testing.T) {
	stations := []pagasa.AWSStation{
		{StationID: "98", StationName: "Science Garden, Quezon City"},
		{StationID: "5001", StationName: "San Jose Synoptic Station"},
	}
	got := filterAWSStations(stations, "98")
	if len(got) != 1 || got[0].StationID != "98" {
		t.Fatalf("id filter: %+v", got)
	}
	got = filterAWSStations(stations, "science")
	if len(got) != 1 || got[0].StationID != "98" {
		t.Fatalf("name filter: %+v", got)
	}
	got = filterAWSStations(stations, "")
	if len(got) != 2 {
		t.Fatalf("empty filter: %d", len(got))
	}
}

func TestWhichIndex_IncludesObs(t *testing.T) {
	root := newRootCmd(&rootFlags{})
	found := false
	for _, e := range whichIndex {
		if e.Command == "obs" || e.Command == "obs history" {
			found = true
			parts := strings.Fields(e.Command)
			cmd, remaining, err := root.Find(parts)
			if err != nil || len(remaining) > 0 || cmd == nil {
				t.Errorf("whichIndex command %q does not resolve (err=%v remaining=%v)", e.Command, err, remaining)
			}
		}
	}
	if !found {
		t.Fatal("whichIndex missing obs entries")
	}
}

func TestObsDryRun(t *testing.T) {
	// Ensure command is registered on root
	root := newRootCmd(&rootFlags{dryRun: true, agent: true})
	cmd, _, err := root.Find([]string{"obs"})
	if err != nil || cmd == nil {
		t.Fatalf("obs not registered: %v", err)
	}
	hist, _, err := root.Find([]string{"obs", "history"})
	if err != nil || hist == nil {
		t.Fatalf("obs history not registered: %v", err)
	}
}
