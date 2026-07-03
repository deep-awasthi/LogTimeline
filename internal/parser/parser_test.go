package parser

import (
	"context"
	"strings"
	"testing"

	"github.com/deepawasthi/logtimeline/internal/config"
	"github.com/deepawasthi/logtimeline/internal/models"
)

func TestParseMixedFormats(t *testing.T) {
	input := strings.Join([]string{
		`2026-07-03 10:15:12 INFO Request received POST /users request_id=7ab45de`,
		`{"timestamp":"2026-07-03T10:15:13Z","level":"INFO","request_id":"7ab45de","message":"SQL INSERT users","logger":"repo.User"}`,
		`2026-07-03 10:15:14 ERROR KafkaException publish failed request_id=7ab45de`,
	}, "\n")
	events, report, err := ParseReader(context.Background(), strings.NewReader(input), config.Default().Parser)
	if err != nil {
		t.Fatalf("ParseReader() error = %v", err)
	}
	if report.Parsed != 3 {
		t.Fatalf("parsed = %d, want 3", report.Parsed)
	}
	if events[0].RequestID != "7ab45de" {
		t.Fatalf("request id = %q", events[0].RequestID)
	}
	var sawSQL bool
	var sawKafkaException bool
	for _, event := range events {
		if event.Kind == models.KindSQL {
			sawSQL = true
		}
		if event.Error != nil && event.Error.Type == "KafkaException" {
			sawKafkaException = true
		}
	}
	if !sawSQL {
		t.Fatalf("expected a SQL event, got %#v", events)
	}
	if !sawKafkaException {
		t.Fatalf("expected KafkaException, got %#v", events)
	}
}

func BenchmarkParseReader(b *testing.B) {
	var builder strings.Builder
	for i := 0; i < 10000; i++ {
		builder.WriteString(`2026-07-03 10:15:12 INFO Request received POST /users request_id=7ab45de duration=12ms` + "\n")
	}
	input := builder.String()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, err := ParseReader(context.Background(), strings.NewReader(input), config.Default().Parser); err != nil {
			b.Fatal(err)
		}
	}
}
