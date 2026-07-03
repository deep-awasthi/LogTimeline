package cmd

import (
	"context"
	"strings"

	"github.com/deepawasthi/logtimeline/internal/export"
	"github.com/deepawasthi/logtimeline/internal/parser"
	"github.com/deepawasthi/logtimeline/internal/timeline"
	"github.com/spf13/cobra"
)

var exportInput string

var exportCmd = &cobra.Command{
	Use:   "export <output-file>",
	Short: "Export a parsed timeline as JSON, Markdown, or CSV",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		events, _, err := parser.ParseFile(context.Background(), exportInput, cfg.Parser)
		if err != nil {
			return err
		}
		format, _ := cmd.Flags().GetString("format")
		if format == "" || format == "auto" {
			format = strings.TrimPrefix(strings.ToLower(export.Ext(args[0])), ".")
		}
		return export.Write(args[0], timeline.Build(events), export.Format(format))
	},
}

func init() {
	exportCmd.Flags().StringVarP(&exportInput, "input", "i", "application.log", "input log file")
	exportCmd.Flags().StringP("format", "f", "auto", "export format: auto, json, md, csv")
	rootCmd.AddCommand(exportCmd)
}
