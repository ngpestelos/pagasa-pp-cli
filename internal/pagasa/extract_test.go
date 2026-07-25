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

const stormDetailHTMLNoSignal = `
<div class=" panel">
    <div class="panel-heading">Location of Eye/center</div>
    <div class="panel-body">
        <p>The center of Typhoon KIYAPO was estimated based on all available data at  465 km West of Basco, Batanes (21.1 &deg;N, 117.4 &deg;E )</p>
    </div>
</div>
<div class="panel">
    <div class="panel-heading">Movement</div>
    <div class="panel-body">
        <p>Moving Northwestward at 25 km/h</p>
    </div>
</div>
<div class=" panel">
    <div class="panel-heading">Strength</div>
    <div class="panel-body">
        <p>Maximum sustained winds of 120 km/h near the center and gustiness of up to 150 km/h</p>
    </div>
</div>
<div class="panel">
    <div class="panel-heading">Forecast Position</div>
    <div class="panel-body">
        <ul>
            <li>Jul 25, 2026 08:00 PM - 610 km West Northwest of Itbayat, Batanes  (OUTSIDE PAR)</li>
            <li>Jul 26, 2026 08:00 AM - 785 km West Northwest of Itbayat, Batanes (OUTSIDE PAR)</li>
        </ul>
    </div>
</div>
<div class="panel">
    <div class="panel-heading">Wind Signal</div>
    <div class="panel-body">
        <span>No Tropical Cyclone Wind Signal</span>
    </div>
</div>
<div class="panel-heading">Tropical Cyclone Bulletin Archive</div>
`

func TestParseStormDetail(t *testing.T) {
	d := ParseStormDetail(stormDetailHTMLNoSignal)
	if d.LatDeg != 21.1 || d.LonDeg != 117.4 {
		t.Errorf("got (%v,%v), want (21.1,117.4)", d.LatDeg, d.LonDeg)
	}
	if d.MoveDir != "Northwestward" || d.MoveSpeedKmh != 25 {
		t.Errorf("got dir=%q speed=%d, want Northwestward/25", d.MoveDir, d.MoveSpeedKmh)
	}
	if d.MaxWindKmh != 120 || d.GustKmh != 150 {
		t.Errorf("got wind=%d gust=%d, want 120/150", d.MaxWindKmh, d.GustKmh)
	}
	if len(d.Forecast) != 2 {
		t.Fatalf("expected 2 forecast points, got %d", len(d.Forecast))
	}
	if d.Forecast[0].ValidAt != "Jul 25, 2026 08:00 PM" {
		t.Errorf("got valid_at %q", d.Forecast[0].ValidAt)
	}
	if d.WindSignals != nil {
		t.Errorf("expected no wind signals, got %+v", d.WindSignals)
	}
}

const windSignalActiveHTML = `
<div class="panel-heading">Wind Signal
    <a href="https://pubfiles.pagasa.dost.gov.ph/tamss/weather/signals_kristine.png">(Areas with TCWS)</a>
</div>
<table class="table text-center table-header">
    <thead>
        <tr><th colspan="2" class="signalno3">Tropical Cyclone Wind Signal no.  <img src="tcws3.png"></th></tr>
    </thead>
    <tbody>
        <tr>
            <td class="text-nowrap text-middle bg-danger"><strong>Affected Areas</strong></td>
            <td>
                <ul style="text-align: left;">
                    <li><strong>Luzon</strong>
                        <ul><li>Cagayan, Isabela, and Batanes</li></ul>
                    </li>
                </ul>
            </td>
        </tr>
        <tr>
            <td class="text-nowrap text-middle bg-info"><strong>Meteorological Condition</strong></td>
            <td><ul style="text-align: left"><li>A tropical cyclone will affect the locality.</li></ul></td>
        </tr>
    </tbody>
    <thead>
        <tr><th colspan="2" class="signalno1">Tropical Cyclone Wind Signal no. <img src="tcws1.png"></th></tr>
    </thead>
    <tbody>
        <tr>
            <td class="text-nowrap text-middle bg-danger"><strong>Affected Areas</strong></td>
            <td>
                <ul style="text-align: left;">
                    <li><strong>Visayas</strong>
                        <ul><li>Eastern Samar and Northern Samar</li></ul>
                    </li>
                </ul>
            </td>
        </tr>
    </tbody>
</table>
<div class="panel-heading">Tropical Cyclone Bulletin Archive</div>
`

func TestParseStormDetailWindSignalsActive(t *testing.T) {
	d := ParseStormDetail(windSignalActiveHTML)
	if len(d.WindSignals) != 2 {
		t.Fatalf("expected 2 wind signals, got %d: %+v", len(d.WindSignals), d.WindSignals)
	}
	if d.WindSignals[0].Signal != 3 {
		t.Errorf("got signal %d, want 3", d.WindSignals[0].Signal)
	}
	if want := "Luzon: Cagayan, Isabela, and Batanes"; d.WindSignals[0].AffectedAreas != want {
		t.Errorf("got areas %q, want %q", d.WindSignals[0].AffectedAreas, want)
	}
	if d.WindSignals[1].Signal != 1 {
		t.Errorf("got signal %d, want 1", d.WindSignals[1].Signal)
	}
	if want := "Visayas: Eastern Samar and Northern Samar"; d.WindSignals[1].AffectedAreas != want {
		t.Errorf("got areas %q, want %q", d.WindSignals[1].AffectedAreas, want)
	}
}
