package cmd

import (
	"context"
	"fmt"

	"github.com/deepawasthi/logtimeline/internal/filters"
	"github.com/deepawasthi/logtimeline/internal/parser"
	"github.com/deepawasthi/logtimeline/internal/timeline"
	"github.com/spf13/cobra"
)

var filterInput string

var filterCmd = &cobra.Command{
	Use:   "filter <level-or-kind>",
	Short: "Filter requests by level or kind such as ERROR, SQL, Kafka, Redis, REST, Exceptions",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		events, _, err := parser.ParseFile(context.Background(), filterInput, cfg.Parser)
		if err != nil {
			return err
		}
		results := filters.Apply(timeline.Build(events).Requests, args[0])
		for _, req := range results {
			fmt.Printf("%s\t%s\t%s\t%d events\t%s\n", req.ID, req.Status, req.Duration, len(req.Events), req.Summary())
		}
		return nil
	},
}

func init() {
	filterCmd.Flags().StringVarP(&filterInput, "input", "i", "application.log", "input log file")
	rootCmd.AddCommand(filterCmd)
}
