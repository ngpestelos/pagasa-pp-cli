// Copyright 2026 Nestor G Pestelos Jr and contributors. Licensed under Apache-2.0. See LICENSE.
// pp:data-source live

package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ngpestelos/pagasa-pp-cli/internal/pagasa"
)

type approachView struct {
	Active     bool     `json:"active"`
	StormName  string   `json:"storm_name,omitempty"`
	StormKind  string   `json:"storm_kind,omitempty"`
	StormLat   float64  `json:"storm_lat,omitempty"`
	StormLon   float64  `json:"storm_lon,omitempty"`
	FromLat    float64  `json:"from_lat"`
	FromLon    float64  `json:"from_lon"`
	DistanceKm *float64 `json:"distance_km,omitempty"`
	Note       string   `json:"note,omitempty"`
	Source     string   `json:"source"`
}

func newNovelApproachCmd(flags *rootFlags) *cobra.Command {
	var flagLocation string
	cmd := &cobra.Command{
		Use:         "approach",
		Short:       "Distance from a fixed location to the active storm's center",
		Long:        "Parse the active tropical cyclone's center coordinates from the PAGASA synopsis and report the great-circle distance to a fixed --location \"lat,lon\". Reports active:false when no cyclone is being tracked.",
		Example:     "  pagasa-pp-cli approach --location 14.58,121.03 --json",
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				fmt.Fprintln(cmd.OutOrStdout(), "would fetch synopsis and compute distance")
				return nil
			}
			lat, lon, err := parseLatLon(flagLocation)
			if err != nil {
				_ = cmd.Usage()
				return usageErr(err)
			}
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			raw, err := c.Get(ctx, "/weather", nil)
			if err != nil {
				return classifyAPIError(err, flags)
			}
			view := approachView{FromLat: lat, FromLon: lon, Source: "live"}
			syn, ok := pagasa.ParseSynopsis(string(raw))
			if !ok || syn.StormName == "" {
				view.Note = "no tropical cyclone is currently being tracked"
				if machineOut(cmd, flags) {
					return printJSONFiltered(cmd.OutOrStdout(), view, flags)
				}
				fmt.Fprintln(cmd.OutOrStdout(), view.Note)
				return nil
			}
			view.Active = true
			view.StormName = syn.StormName
			view.StormKind = syn.StormKind
			if slat, slon, ok := pagasa.ParsePosition(syn.Text); ok {
				view.StormLat, view.StormLon = slat, slon
				d := pagasa.HaversineKm(lat, lon, slat, slon)
				view.DistanceKm = &d
			} else {
				view.Note = "storm active but center coordinates not found in synopsis"
			}
			if machineOut(cmd, flags) {
				return printJSONFiltered(cmd.OutOrStdout(), view, flags)
			}
			if view.DistanceKm != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "%s %q is %.0f km from your location.\n",
					view.StormKind, view.StormName, *view.DistanceKm)
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), view.Note)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&flagLocation, "location", "", "fixed point as \"lat,lon\" (e.g. 14.58,121.03 for Mandaluyong)")
	return cmd
}

func parseLatLon(s string) (float64, float64, error) {
	parts := strings.Split(strings.TrimSpace(s), ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("--location must be \"lat,lon\", e.g. 14.58,121.03")
	}
	lat, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	lon, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err1 != nil || err2 != nil {
		return 0, 0, fmt.Errorf("--location must be numeric \"lat,lon\", e.g. 14.58,121.03")
	}
	return lat, lon, nil
}
