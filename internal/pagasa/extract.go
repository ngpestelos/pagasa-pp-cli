// Copyright 2026 Nestor G Pestelos Jr and contributors. Licensed under Apache-2.0. See LICENSE.

// Package pagasa holds hand-written extractors for the public PAGASA weather
// surfaces (server-rendered HTML on www.pagasa.dost.gov.ph and downloadable
// artifacts on pubfiles.pagasa.dost.gov.ph). The site publishes no JSON API,
// so these parsers turn the panel/table HTML into structured Go values.
package pagasa

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/ngpestelos/pagasa-pp-cli/internal/cliutil"
)

// --- Synopsis (also used by the `now` command's sibling logic) ---

var (
	synopsisRe = regexp.MustCompile(`(?is)panel-heading">\s*Synopsis\s*</div>\s*<div class="panel-body">\s*<p>(.*?)</p>`)
	stormRe    = regexp.MustCompile(`(?i)(Super Typhoon|Typhoon|Severe Tropical Storm|Tropical Storm|Tropical Depression)\s+"([^"]+)"`)
)

// Synopsis holds the parsed weather synopsis and any active cyclone.
type Synopsis struct {
	Text      string `json:"synopsis"`
	StormName string `json:"storm_name,omitempty"`
	StormKind string `json:"storm_kind,omitempty"`
}

// ParseSynopsis extracts the synopsis paragraph from the /weather page HTML.
func ParseSynopsis(html string) (Synopsis, bool) {
	m := synopsisRe.FindStringSubmatch(html)
	if m == nil {
		return Synopsis{}, false
	}
	s := Synopsis{Text: cliutil.CleanText(m[1])}
	if sm := stormRe.FindStringSubmatch(s.Text); sm != nil {
		s.StormKind = strings.TrimSpace(sm[1])
		s.StormName = strings.TrimSpace(sm[2])
	}
	return s, true
}

// --- City forecast ---

var (
	// Each city is an accordion panel whose title anchor carries the city name:
	// <div class="panel-title"><a ... href="#acc-...">CITY &nbsp;<icon ...>
	cityPanelRe = regexp.MustCompile(`(?is)panel-title">\s*<a[^>]*href="#acc-[^"]*"[^>]*>\s*([A-Za-z().\s]+?)\s*(?:&nbsp;|<icon)`)
	dateThRe    = regexp.MustCompile(`(?is)<th[^>]*>\s*([A-Za-z]+)\s*</?br\s*/?>\s*([A-Za-z]+\s+\d{1,2},\s*\d{4})\s*</th>`)
	minMaxRe    = regexp.MustCompile(`(?is)class="min">\s*(\d+)°C</span>\s*<span class="max">\s*(\d+)°C`)
	rainRe      = regexp.MustCompile(`(?i)Chance of rain:\s*(\d+)%`)
	condRe      = regexp.MustCompile(`(?is)<img[^>]*title="([^"]+)"`)
)

// DayForecast is one day's forecast for a city.
type DayForecast struct {
	Day        string `json:"day"`
	Date       string `json:"date"`
	Condition  string `json:"condition,omitempty"`
	MinC       int    `json:"min_c"`
	MaxC       int    `json:"max_c"`
	RainChance int    `json:"rain_chance_pct"`
}

// CityForecast is the multi-day forecast for a single city.
type CityForecast struct {
	City string        `json:"city"`
	Days []DayForecast `json:"days"`
}

// ParseCityForecasts extracts all city forecast panels from the
// weather-outlook-selected-philippine-cities page HTML. The page renders each
// day twice (a desktop table and a mobile table); we keep the first N unique
// day/forecast tuples per city so the duplication does not double the output.
func ParseCityForecasts(html string) []CityForecast {
	idx := cityPanelRe.FindAllStringSubmatchIndex(html, -1)
	if idx == nil {
		return nil
	}
	var out []CityForecast
	for i, loc := range idx {
		city := cliutil.CleanText(html[loc[2]:loc[3]])
		start := loc[1]
		end := len(html)
		if i+1 < len(idx) {
			end = idx[i+1][0]
		}
		seg := html[start:end]
		out = append(out, CityForecast{City: city, Days: parseDays(seg)})
	}
	return out
}

func parseDays(seg string) []DayForecast {
	dates := dateThRe.FindAllStringSubmatch(seg, -1)
	temps := minMaxRe.FindAllStringSubmatch(seg, -1)
	rains := rainRe.FindAllStringSubmatch(seg, -1)
	conds := condRe.FindAllStringSubmatch(seg, -1)

	n := len(dates)
	for _, l := range [][]int{{len(temps)}, {len(rains)}} {
		if l[0] < n {
			n = l[0]
		}
	}
	// The desktop table already carries a full set of days; cap at a sane 7
	// so the mobile duplicate table does not extend the slice.
	if n > 7 {
		n = 7
	}
	var days []DayForecast
	seen := map[string]bool{}
	for i := 0; i < n; i++ {
		key := dates[i][2]
		if seen[key] {
			continue
		}
		seen[key] = true
		d := DayForecast{
			Day:        strings.TrimSpace(dates[i][1]),
			Date:       cliutil.CleanText(dates[i][2]),
			MinC:       atoi(temps[i][1]),
			MaxC:       atoi(temps[i][2]),
			RainChance: atoi(rains[i][1]),
		}
		if i < len(conds) {
			d.Condition = cliutil.CleanText(conds[i][1])
		}
		days = append(days, d)
	}
	return days
}

func atoi(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

// --- Tropical cyclone bulletin index ---

var (
	pdfLinkRe   = regexp.MustCompile(`(?i)href="(https://pubfiles\.pagasa\.dost\.gov\.ph/[^"]+\.pdf)"`)
	signalPngRe = regexp.MustCompile(`(?i)href="(https://pubfiles\.pagasa\.dost\.gov\.ph/[^"]+signals[^"]+\.png)"`)
)

// Bulletin holds links to the current tropical-cyclone bulletin artifacts.
type Bulletin struct {
	PDFs      []string `json:"bulletin_pdfs"`
	SignalMap string   `json:"signal_map,omitempty"`
}

// --- Storm position + distance ---

// coordRe matches a "(15.4°N, 130.6°E)" coordinate pair inside synopsis text.
var coordRe = regexp.MustCompile(`\(?\s*([0-9.]+)\s*°?\s*N\s*,?\s*([0-9.]+)\s*°?\s*E\s*\)?`)

// ParsePosition extracts the storm center latitude/longitude from synopsis text.
func ParsePosition(synopsis string) (lat, lon float64, ok bool) {
	m := coordRe.FindStringSubmatch(synopsis)
	if m == nil {
		return 0, 0, false
	}
	la, err1 := strconv.ParseFloat(m[1], 64)
	lo, err2 := strconv.ParseFloat(m[2], 64)
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return la, lo, true
}

// HaversineKm returns the great-circle distance in kilometres between two points.
func HaversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	const r = 6371.0
	rad := func(d float64) float64 { return d * math.Pi / 180 }
	dLat := rad(lat2 - lat1)
	dLon := rad(lon2 - lon1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(rad(lat1))*math.Cos(rad(lat2))*math.Sin(dLon/2)*math.Sin(dLon/2)
	return r * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// ParseBulletin extracts bulletin PDF and wind-signal map links from the
// severe-weather-bulletin page HTML.
func ParseBulletin(html string) Bulletin {
	var b Bulletin
	seen := map[string]bool{}
	for _, m := range pdfLinkRe.FindAllStringSubmatch(html, -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			b.PDFs = append(b.PDFs, m[1])
		}
	}
	if m := signalPngRe.FindStringSubmatch(html); m != nil {
		b.SignalMap = m[1]
	}
	return b
}
