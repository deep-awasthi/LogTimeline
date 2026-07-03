package export

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/deepawasthi/logtimeline/internal/models"
)

type Format string

const (
	FormatJSON     Format = "json"
	FormatMarkdown Format = "md"
	FormatCSV      Format = "csv"
)

func Ext(path string) string {
	return filepath.Ext(path)
}

func Write(path string, timeline models.Timeline, format Format) error {
	switch normalized(format) {
	case FormatJSON:
		return writeJSON(path, timeline)
	case FormatMarkdown:
		return writeMarkdown(path, timeline)
	case FormatCSV:
		return writeCSV(path, timeline)
	default:
		return fmt.Errorf("unsupported export format %q", format)
	}
}

func normalized(format Format) Format {
	switch strings.ToLower(strings.TrimPrefix(string(format), ".")) {
	case "json":
		return FormatJSON
	case "md", "markdown":
		return FormatMarkdown
	case "csv":
		return FormatCSV
	default:
		return ""
	}
}

func writeJSON(path string, timeline models.Timeline) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	return enc.Encode(timeline)
}

func writeMarkdown(path string, timeline models.Timeline) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	for _, req := range timeline.Requests {
		fmt.Fprintf(file, "## Request %s\n\n", req.ID)
		fmt.Fprintf(file, "- Grouped by: %s\n- Status: %s\n- Duration: %s\n- Events: %d\n\n", req.GroupBy, req.Status, models.FormatDuration(req.Duration), len(req.Events))
		for i, event := range req.Events {
			if i > 0 {
				fmt.Fprintln(file, "\n↓")
			}
			fmt.Fprintf(file, "%s `%s` %s\n", event.Timestamp.Format("15:04:05"), event.Level, event.Message)
		}
		fmt.Fprintln(file)
	}
	return nil
}

func writeCSV(path string, timeline models.Timeline) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	if err := writer.Write([]string{"request_id", "status", "duration", "timestamp", "level", "kind", "logger", "message"}); err != nil {
		return err
	}
	for _, req := range timeline.Requests {
		for _, event := range req.Events {
			if err := writer.Write([]string{req.ID, req.Status, models.FormatDuration(req.Duration), event.Timestamp.Format(timeLayout), string(event.Level), string(event.Kind), event.Logger, event.Message}); err != nil {
				return err
			}
		}
	}
	if err := writer.Error(); err != nil {
		return err
	}
	return nil
}

const timeLayout = "2006-01-02T15:04:05.000Z07:00"

var ErrUnsupportedFormat = errors.New("unsupported export format")
