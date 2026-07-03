package cmd

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/deepawasthi/logtimeline/internal/parser"
	"github.com/deepawasthi/logtimeline/internal/tail"
	"github.com/deepawasthi/logtimeline/internal/timeline"
	"github.com/deepawasthi/logtimeline/internal/ui"
	"github.com/spf13/cobra"
)

var tailCmd = &cobra.Command{
	Use:   "tail <log-file>",
	Short: "Follow a log file and update the timeline live",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		source := tail.NewSource(args[0], cfg.Tail)
		model := ui.NewLiveModel(timeline.NewStore(), parser.NewDetector(), source, cfg.UI)
		_, err := tea.NewProgram(model, tea.WithAltScreen()).Run()
		if err != nil {
			source.Stop(context.Background())
		}
		return err
	},
}

func init() {
	rootCmd.AddCommand(tailCmd)
}
