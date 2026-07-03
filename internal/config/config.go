package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	LogLevel string       `mapstructure:"log_level"`
	Parser   ParserConfig `mapstructure:"parser"`
	Tail     TailConfig   `mapstructure:"tail"`
	UI       UIConfig     `mapstructure:"ui"`
}

type ParserConfig struct {
	Workers       int           `mapstructure:"workers"`
	ChannelBuffer int           `mapstructure:"channel_buffer"`
	MaxLineBytes  int           `mapstructure:"max_line_bytes"`
	StackWindow   time.Duration `mapstructure:"stack_window"`
}

type TailConfig struct {
	PollInterval time.Duration `mapstructure:"poll_interval"`
	FromEnd      bool          `mapstructure:"from_end"`
}

type UIConfig struct {
	PageSize int `mapstructure:"page_size"`
}

func Default() Config {
	return Config{
		LogLevel: "warn",
		Parser: ParserConfig{
			Workers:       4,
			ChannelBuffer: 4096,
			MaxLineBytes:  4 * 1024 * 1024,
			StackWindow:   2 * time.Second,
		},
		Tail: TailConfig{
			PollInterval: 500 * time.Millisecond,
			FromEnd:      false,
		},
		UI: UIConfig{PageSize: 30},
	}
}

func Load(path string) (Config, error) {
	cfg := Default()
	v := viper.New()
	v.SetConfigType("yaml")
	if path != "" {
		v.SetConfigFile(path)
	} else {
		v.SetConfigName(".logtimeline")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME")
	}
	v.SetDefault("log_level", cfg.LogLevel)
	v.SetDefault("parser.workers", cfg.Parser.Workers)
	v.SetDefault("parser.channel_buffer", cfg.Parser.ChannelBuffer)
	v.SetDefault("parser.max_line_bytes", cfg.Parser.MaxLineBytes)
	v.SetDefault("parser.stack_window", cfg.Parser.StackWindow)
	v.SetDefault("tail.poll_interval", cfg.Tail.PollInterval)
	v.SetDefault("tail.from_end", cfg.Tail.FromEnd)
	v.SetDefault("ui.page_size", cfg.UI.PageSize)
	v.SetEnvPrefix("LOGTIMELINE")
	v.AutomaticEnv()
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok && path != "" {
			return cfg, err
		}
	}
	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, err
	}
	if cfg.Parser.Workers < 1 {
		cfg.Parser.Workers = 1
	}
	if cfg.Parser.ChannelBuffer < 1 {
		cfg.Parser.ChannelBuffer = 1
	}
	if cfg.Parser.MaxLineBytes < 1024 {
		cfg.Parser.MaxLineBytes = 1024
	}
	if cfg.UI.PageSize < 5 {
		cfg.UI.PageSize = 5
	}
	return cfg, nil
}
