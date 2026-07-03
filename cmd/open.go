package cmd

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/deepawasthi/logtimeline/internal/parser"
	"github.com/deepawasthi/logtimeline/internal/timeline"
	"github.com/deepawasthi/logtimeline/internal/ui"
	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open <log-file>",
	Short: "Open a log file in the interactive timeline UI",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		events, report, err := parser.ParseFile(context.Background(), args[0], cfg.Parser)
		if err != nil {
			return err
		}
		tl := timeline.Build(events)
		model := ui.NewModel(tl, report, cfg.UI)
		_, err = tea.NewProgram(model, tea.WithAltScreen()).Run()
		return err
	},
}

func init() {
	rootCmd.AddCommand(openCmd)
}
