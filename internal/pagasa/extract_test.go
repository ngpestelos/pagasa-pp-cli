// Copyright 2026 Nestor G Pestelos Jr and contributors. Licensed under Apache-2.0. See LICENSE.

package pagasa

import "testing"

const synopsisHTML = `<div class="panel"><div class="panel-heading">Synopsis</div>
<div class="panel-body"><p>At 3:00 AM today, the center of Tropical Depression &quot;KIYAPO&quot; was estimated at 960 km East of Central Luzon (15.4&deg;N, 130.6&deg;E) with maximum sustained winds of 45 km/h.</p></div></div>`

func TestParseSynopsis(t *testing.T) {
	s, ok := ParseSynopsis(synopsisHTML)
	if !ok {
		t.Fatal("expected synopsis to parse")
	}
	if s.StormName != "KIYAPO" {
		t.Errorf("storm name = %q, want KIYAPO", s.StormName)
	}
	if s.StormKind != "Tropical Depression" {
		t.Errorf("storm kind = %q, want Tropical Depression", s.StormKind)
	}
	if s.Text == "" {
		t.Error("synopsis text is empty")
	}
}

func TestParseSynopsisNoStorm(t *testing.T) {
	html := `<div class="panel-heading">Synopsis</div><div class="panel-body"><p>Ridge of high pressure affecting the country.</p></div>`
	s, ok := ParseSynopsis(html)
	if !ok {
		t.Fatal("expected parse")
	}
	if s.StormName != "" {
		t.Errorf("expected no storm, got %q", s.StormName)
	}
}

func TestParseSynopsisMissing(t *testing.T) {
	if _, ok := ParseSynopsis("<html><body>no panel</body></html>"); ok {
		t.Error("expected missing synopsis to return false")
	}
}

const cityHTML = `<div class="panel-heading panel-pagasa"><div class="panel-title"><a data-toggle="collapse" href="#acc-64501000">Metro Manila &nbsp;<icon class="glyphicon"></icon></a></div></div>
<div class="panel-body"><table class="table"><thead class="desktop-view-thead"><tr>
<th class="text-center">Thursday </br> July 23, 2026</th>
<th class="text-center">Friday </br> July 24, 2026</th></tr></thead>
<tbody><tr>
<td><img src="x.png" title="Cloudy skies with rainshowers"/><div class="ol-temperature"><span class="min">24°C</span><span class="max">31°C</span></div><span>Chance of rain: 90%</span></td>
<td><img src="y.png" title="Partly cloudy"/><div class="ol-temperature"><span class="min">24°C</span><span class="max">30°C</span></div><span>Chance of rain: 100%</span></td>
</tr></tbody></table></div>`

func TestParseCityForecasts(t *testing.T) {
	cities := ParseCityForecasts(cityHTML)
	if len(cities) != 1 {
		t.Fatalf("expected 1 city, got %d", len(cities))
	}
	c := cities[0]
	if c.City != "Metro Manila" {
		t.Errorf("city = %q, want Metro Manila", c.City)
	}
	if len(c.Days) != 2 {
		t.Fatalf("expected 2 days, got %d", len(c.Days))
	}
	d0 := c.Days[0]
	if d0.Day != "Thursday" || d0.MinC != 24 || d0.MaxC != 31 || d0.RainChance != 90 {
		t.Errorf("day0 = %+v, unexpected", d0)
	}
	if d0.Condition == "" {
		t.Error("day0 condition empty")
	}
}

func TestParsePosition(t *testing.T) {
	lat, lon, ok := ParsePosition(`estimated at 960 km East of Central Luzon (15.4°N, 130.6°E) with winds`)
	if !ok {
		t.Fatal("expected position to parse")
	}
	if lat != 15.4 || lon != 130.6 {
		t.Errorf("got (%v,%v), want (15.4,130.6)", lat, lon)
	}
	if _, _, ok := ParsePosition("no coordinates here"); ok {
		t.Error("expected no position")
	}
}

func TestHaversineKm(t *testing.T) {
	// Mandaluyong (14.58,121.03) to the storm center (15.4,130.6) ~ 1000+ km.
	d := HaversineKm(14.58, 121.03, 15.4, 130.6)
	if d < 900 || d > 1100 {
		t.Errorf("distance = %.0f km, expected ~1000 km", d)
	}
	if d0 := HaversineKm(14.58, 121.03, 14.58, 121.03); d0 > 0.001 {
		t.Errorf("zero distance expected, got %v", d0)
	}
}

func TestParseBulletin(t *testing.T) {
	html := `<a href="https://pubfiles.pagasa.dost.gov.ph/tamss/weather/bulletin/TCB%231_kiyapo.pdf">TCB1</a>
<a href="https://pubfiles.pagasa.dost.gov.ph/tamss/weather/signals_kiyapo.png">signals</a>
<a href="https://pubfiles.pagasa.dost.gov.ph/tamss/weather/bulletin/TCB%231_kiyapo.pdf">dup</a>`
	b := ParseBulletin(html)
	if len(b.PDFs) != 1 {
		t.Errorf("expected 1 deduped pdf, got %d", len(b.PDFs))
	}
	if b.SignalMap == "" {
		t.Error("expected signal map link")
	}
}
