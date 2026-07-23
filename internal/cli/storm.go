// Copyright 2026 Nestor G Pestelos Jr and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ngpestelos/pagasa-pp-cli/internal/pagasa"
)

type stormView struct {
	Active    bool     `json:"active"`
	StormName string   `json:"storm_name,omitempty"`
	StormKind string   `json:"storm_kind,omitempty"`
	Synopsis  string   `json:"synopsis,omitempty"`
	Bulletins []string `json:"bulletin_pdfs,omitempty"`
	SignalMap string   `json:"signal_map,omitempty"`
	Source    string   `json:"source"`
}

func newStormCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "storm",
		Short:       "Active tropical cyclone: synopsis, bulletin PDFs, and wind-signal map",
		Long:        "Combine the PAGASA synopsis with the severe-weather-bulletin index so agents get the active cyclone name, downloadable bulletin PDFs, and the wind-signal map in one call. Reports active:false when no cyclone is being tracked.",
		Example:     "  pagasa-pp-cli storm --json",
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				fmt.Fprintln(cmd.OutOrStdout(), "would fetch /weather and /tropical-cyclone/severe-weather-bulletin")
				return nil
			}
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			view := stormView{Source: "live"}

			if raw, err := c.Get(ctx, "/weather", nil); err == nil {
				if syn, ok := pagasa.ParseSynopsis(string(raw)); ok {
					view.Synopsis = syn.Text
					view.StormName = syn.StormName
					view.StormKind = syn.StormKind
				}
			}

			raw, err := c.Get(ctx, "/tropical-cyclone/severe-weather-bulletin", nil)
			if err != nil {
				return classifyAPIError(err, flags)
			}
			b := pagasa.ParseBulletin(string(raw))
			view.Bulletins = b.PDFs
			view.SignalMap = b.SignalMap
			view.Active = view.StormName != "" || len(b.PDFs) > 0

			if machineOut(cmd, flags) {
				return printJSONFiltered(cmd.OutOrStdout(), view, flags)
			}
			if !view.Active {
				fmt.Fprintln(cmd.OutOrStdout(), "No tropical cyclone is currently being tracked.")
				return nil
			}
			if view.StormName != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "%s %q\n\n%s\n\n", view.StormKind, view.StormName, view.Synopsis)
			}
			for _, p := range view.Bulletins {
				fmt.Fprintf(cmd.OutOrStdout(), "  bulletin: %s\n", p)
			}
			if view.SignalMap != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "  signals:  %s\n", view.SignalMap)
			}
			return nil
		},
	}
	return cmd
}
