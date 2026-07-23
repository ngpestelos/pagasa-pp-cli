// Copyright 2026 Nestor G Pestelos Jr and contributors. Licensed under Apache-2.0. See LICENSE.
// pp:data-source live

package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ngpestelos/pagasa-pp-cli/internal/pagasa"
)

type watchView struct {
	Area        string `json:"area"`
	StormActive bool   `json:"storm_active"`
	StormName   string `json:"storm_name,omitempty"`
	StormKind   string `json:"storm_kind,omitempty"`
	SignalMap   string `json:"signal_map,omitempty"`
	Note        string `json:"note"`
	Source      string `json:"source"`
}

func newNovelWatchCmd(flags *rootFlags) *cobra.Command {
	var flagArea string
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Wind-signal state relevant to a locality (honest when clear)",
		Long: strings.Trim(`
Report whether a tropical cyclone is active and point to the official wind-signal
map for a locality. PAGASA publishes per-area signal numbers only inside the
bulletin PDF (not machine-readable HTML), so when a storm is active this command
returns the signal-map image URL and the bulletin; when no storm is active it
returns an honest "clear" state rather than fabricating a signal level.`, "\n"),
		Example:     "  pagasa-pp-cli watch --area Mandaluyong --json",
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				fmt.Fprintln(cmd.OutOrStdout(), "would fetch synopsis and bulletin index")
				return nil
			}
			if strings.TrimSpace(flagArea) == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("--area is required"))
			}
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			view := watchView{Area: flagArea, Source: "live"}
			if raw, err := c.Get(ctx, "/weather", nil); err == nil {
				if syn, ok := pagasa.ParseSynopsis(string(raw)); ok && syn.StormName != "" {
					view.StormActive = true
					view.StormName = syn.StormName
					view.StormKind = syn.StormKind
				}
			}
			if braw, err := c.Get(ctx, "/tropical-cyclone/severe-weather-bulletin", nil); err == nil {
				view.SignalMap = pagasa.ParseBulletin(string(braw)).SignalMap
			}
			if view.StormActive {
				view.Note = fmt.Sprintf("%s %q is active; check the signal map for %s's wind signal number",
					view.StormKind, view.StormName, flagArea)
			} else {
				view.Note = fmt.Sprintf("no wind signal is in effect for %s (no tropical cyclone active)", flagArea)
			}
			if machineOut(cmd, flags) {
				return printJSONFiltered(cmd.OutOrStdout(), view, flags)
			}
			fmt.Fprintln(cmd.OutOrStdout(), view.Note)
			if view.SignalMap != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "signal map: %s\n", view.SignalMap)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&flagArea, "area", "", "locality name to watch (e.g. Mandaluyong)")
	return cmd
}
