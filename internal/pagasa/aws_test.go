// Copyright 2026 Nestor G Pestelos Jr and contributors. Licensed under Apache-2.0. See LICENSE.

package pagasa

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseAWSStations_Fixture(t *testing.T) {
	html := readAWSFixture(t)
	stations := ParseAWSStations(html)
	if len(stations) < 50 {
		t.Fatalf("station count = %d, want >= 50 (Phase 0 kill threshold)", len(stations))
	}
	var science *AWSStation
	for i := range stations {
		if stations[i].StationID == "98" || stations[i].StationName == "Science Garden, Quezon City" {
			science = &stations[i]
			break
		}
	}
	if science == nil {
		t.Fatal("expected Science Garden (site 98) in fixture")
	}
	if science.StationID != "98" {
		t.Errorf("science station_id = %q, want 98", science.StationID)
	}
	if science.TempC == nil {
		t.Error("science temp_c nil")
	}
	if science.ObservedAt == "" {
		t.Error("science observed_at empty — Last Updated parse failed")
	}
	// ObservedAt must be RFC3339 UTC
	if _, err := time.Parse(time.RFC3339, science.ObservedAt); err != nil {
		t.Errorf("observed_at %q not RFC3339: %v", science.ObservedAt, err)
	}
}

func TestParseAWSLastUpdated_PHT(t *testing.T) {
	got, err := ParseAWSLastUpdated("August 3, 2026, 7:40 am")
	if err != nil {
		t.Fatal(err)
	}
	// 7:40 am PHT = 23:40 previous day UTC on Aug 2... wait: Aug 3 07:40 +8 = Aug 2 23:40 UTC? 
	// Aug 3 07:40 PHT = Aug 2 23:40 UTC. Yes.
	utc := got.UTC()
	if utc.Year() != 2026 || utc.Month() != time.August || utc.Day() != 2 {
		t.Errorf("UTC date = %s, want 2026-08-02", utc.Format(time.RFC3339))
	}
	if utc.Hour() != 23 || utc.Minute() != 40 {
		t.Errorf("UTC time = %02d:%02d, want 23:40", utc.Hour(), utc.Minute())
	}
	// Same wall clock must not depend on process local TZ
	if got.Location().String() != "PHT" {
		t.Errorf("location = %s, want PHT", got.Location())
	}
}

func TestParseAWSLastUpdated_PM(t *testing.T) {
	got, err := ParseAWSLastUpdated("July 28, 2026, 6:50 pm")
	if err != nil {
		t.Fatal(err)
	}
	utc := got.UTC()
	// 18:50 PHT = 10:50 UTC same day
	if utc.Hour() != 10 || utc.Minute() != 50 {
		t.Errorf("UTC = %s, want 10:50", utc.Format(time.RFC3339))
	}
}

func TestParseAWSStations_GarbledRowTolerated(t *testing.T) {
	html := `<table class="table"><tbody>
<tr>
<td>1</td><td>Good Station</td>
<td>25 °C</td><td>90 %</td><td>3.6 km/hr</td><td>N</td>
<td>0 mm/hr</td><td>1000</td><td>10</td>
<td>August 3, 2026, 7:40 am</td>
</tr>
<tr>
<td>2</td><td>Bad Temps</td>
<td>--</td><td>n/a</td><td></td><td></td>
<td>-</td><td></td><td></td>
<td>not a date</td>
</tr>
</tbody></table>`
	stations := ParseAWSStations(html)
	if len(stations) != 2 {
		t.Fatalf("got %d stations, want 2", len(stations))
	}
	if stations[0].TempC == nil || *stations[0].TempC != 25 {
		t.Errorf("good temp = %v", stations[0].TempC)
	}
	if stations[1].TempC != nil {
		t.Errorf("bad temp should be nil, got %v", stations[1].TempC)
	}
	if stations[1].ObservedAt != "" {
		t.Errorf("bad date should leave observed_at empty, got %q", stations[1].ObservedAt)
	}
	if len(stations[1].ParseWarnings) == 0 {
		t.Error("expected parse warning for unparseable date")
	}
}

func TestParseAWSFloat_NoTzdataDependency(t *testing.T) {
	// Ensure PHT is FixedZone, not LoadLocation
	if PHT.String() != "PHT" {
		t.Fatalf("PHT = %s", PHT)
	}
	// Offset must be +8h
	_, off := time.Now().In(PHT).Zone()
	if off != 8*3600 {
		t.Fatalf("PHT offset = %d, want %d", off, 8*3600)
	}
}

func readAWSFixture(t *testing.T) string {
	t.Helper()
	p := filepath.Join("testdata", "aws_stations.html")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return string(b)
}
