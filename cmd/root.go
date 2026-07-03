package cmd

import (
	"fmt"
	"os"

	"github.com/deepawasthi/logtimeline/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	cfg     config.Config
	logger  = logrus.New()
)

var rootCmd = &cobra.Command{
	Use:           "logtimeline",
	Short:         "Convert application logs into interactive request timelines",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
	rootCmd.PersistentFlags().String("log-level", "warn", "application log level")
	_ = rootCmd.PersistentFlags().MarkHidden("log-level")
	_ = rootCmd.PersistentFlags().Lookup("log-level")
}

func initConfig() {
	var err error
	cfg, err = config.Load(cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
		os.Exit(1)
	}
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = logrus.WarnLevel
	}
	logger.SetLevel(level)
	logger.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
}
