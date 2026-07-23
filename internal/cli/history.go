// Copyright 2026 Nestor G Pestelos Jr and contributors. Licensed under Apache-2.0. See LICENSE.
// pp:data-source local

package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

type historyEntry struct {
	CapturedAt string `json:"captured_at"`
	StormName  string `json:"storm_name,omitempty"`
	StormKind  string `json:"storm_kind,omitempty"`
	Synopsis   string `json:"synopsis,omitempty"`
}

func newNovelHistoryCmd(flags *rootFlags) *cobra.Command {
	var flagLimit int
	cmd := &cobra.Command{
		Use:         "history",
		Short:       "Past synopsis/cyclone snapshots this CLI has recorded locally",
		Long:        "PAGASA serves only the latest synopsis. This command reads the local SQLite mirror of snapshots that `now`, `storm`, and `digest` persist over time, newest first. It returns an empty list (not an error) before any snapshot has been recorded.",
		Example:     "  pagasa-pp-cli history --limit 10 --json",
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				fmt.Fprintln(cmd.OutOrStdout(), "would read local snapshot history")
				return nil
			}
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			snaps, err := loadSnapshots(ctx, defaultDBPath("pagasa-pp-cli"))
			if err != nil {
				return err
			}
			if flagLimit > 0 && len(snaps) > flagLimit {
				snaps = snaps[:flagLimit]
			}
			entries := make([]historyEntry, 0, len(snaps))
			for _, s := range snaps {
				entries = append(entries, historyEntry{
					CapturedAt: s.CapturedAt, StormName: s.StormName,
					StormKind: s.StormKind, Synopsis: s.Synopsis,
				})
			}
			if machineOut(cmd, flags) {
				return printJSONFiltered(cmd.OutOrStdout(), entries, flags)
			}
			if len(entries) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No snapshots recorded yet. Run 'pagasa-pp-cli digest' or 'now' to start a local history.")
				return nil
			}
			for _, e := range entries {
				storm := "no active storm"
				if e.StormName != "" {
					storm = fmt.Sprintf("%s %q", e.StormKind, e.StormName)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s  %s\n", e.CapturedAt, storm)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&flagLimit, "limit", 20, "maximum snapshots to return (newest first)")
	return cmd
}
