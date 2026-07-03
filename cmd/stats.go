package cmd

import (
	"context"
	"fmt"

	"github.com/deepawasthi/logtimeline/internal/parser"
	statspkg "github.com/deepawasthi/logtimeline/internal/stats"
	"github.com/deepawasthi/logtimeline/internal/timeline"
	"github.com/spf13/cobra"
)

var statsInput string

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Print request, duration, error, exception, and endpoint statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		events, _, err := parser.ParseFile(context.Background(), statsInput, cfg.Parser)
		if err != nil {
			return err
		}
		report := statspkg.Compute(timeline.Build(events).Requests)
		fmt.Print(report.String())
		return nil
	},
}

func init() {
	statsCmd.Flags().StringVarP(&statsInput, "input", "i", "application.log", "input log file")
	rootCmd.AddCommand(statsCmd)
}
