package cmd

import (
	"context"
	"fmt"

	"github.com/deepawasthi/logtimeline/internal/parser"
	searchpkg "github.com/deepawasthi/logtimeline/internal/search"
	"github.com/deepawasthi/logtimeline/internal/timeline"
	"github.com/spf13/cobra"
)

var searchInput string

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search requests by text, IDs, logger, HTTP, SQL, Kafka, Redis, or exceptions",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		events, _, err := parser.ParseFile(context.Background(), searchInput, cfg.Parser)
		if err != nil {
			return err
		}
		results := searchpkg.Search(timeline.Build(events).Requests, args[0])
		for _, req := range results {
			fmt.Printf("%s\t%s\t%s\t%d events\t%s\n", req.ID, req.Status, req.Duration, len(req.Events), req.Summary())
		}
		return nil
	},
}

func init() {
	searchCmd.Flags().StringVarP(&searchInput, "input", "i", "application.log", "input log file")
	rootCmd.AddCommand(searchCmd)
}
