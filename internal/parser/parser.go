package parser

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/deepawasthi/logtimeline/internal/config"
	"github.com/deepawasthi/logtimeline/internal/models"
)

type Format string

const (
	FormatJSON       Format = "json"
	FormatSpringBoot Format = "spring_boot"
	FormatLogback    Format = "logback"
	FormatCustom     Format = "custom"
	FormatText       Format = "text"
)

type Report struct {
	Format     Format
	Lines      int64
	Parsed     int64
	Unparsed   int64
	StartedAt  time.Time
	FinishedAt time.Time
}

type lineJob struct {
	number int64
	text   string
}

type parsedLine struct {
	event models.Event
	ok    bool
}

type Detector struct {
	springRe *regexp.Regexp
	logback  *regexp.Regexp
	custom   *regexp.Regexp
	kv       *regexp.Regexp
	httpReq  *regexp.Regexp
	httpResp *regexp.Regexp
	duration *regexp.Regexp
}

func NewDetector() *Detector {
	return &Detector{
		springRe: regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}[ T]\d{2}:\d{2}:\d{2}(?:[.,]\d{1,9})?)\s+([A-Z]+)\s+(?:\d+\s+---\s+\[([^\]]+)\]\s+)?([^\s:]+)?\s*:?\s*(.*)$`),
		logback:  regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}[ T]\d{2}:\d{2}:\d{2}(?:[.,]\d{1,9})?)\s+\[([^\]]+)\]\s+([A-Z]+)\s+([^\s-]+)\s+-\s+(.*)$`),
		custom:   regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}[ T]\d{2}:\d{2}:\d{2}(?:[.,]\d{1,9})?)\s+([A-Z]+)\s+(.*)$`),
		kv:       regexp.MustCompile(`(?i)\b(request[_-]?id|trace[_-]?id|span[_-]?id|user[_-]?id|session[_-]?id|duration|elapsed|logger|thread)=("?[^"\s]+"?|\S+)`),
		httpReq:  regexp.MustCompile(`(?i)\b(GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS)\s+(/[^\s"]*)`),
		httpResp: regexp.MustCompile(`(?i)\b(?:response|status)\s+([1-5][0-9]{2})\b`),
		duration: regexp.MustCompile(`(?i)\b(?:duration|elapsed|took)[=: ]+([0-9]+(?:\.[0-9]+)?)(ms|s|sec|secs|seconds|m)?\b`),
	}
}

func ParseFile(ctx context.Context, path string, cfg config.ParserConfig) ([]models.Event, Report, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, Report{}, err
	}
	defer file.Close()
	return ParseReader(ctx, file, cfg)
}

func ParseReader(ctx context.Context, r io.Reader, cfg config.ParserConfig) ([]models.Event, Report, error) {
	if cfg.Workers < 1 {
		cfg.Workers = runtime.NumCPU()
	}
	if cfg.ChannelBuffer < 1 {
		cfg.ChannelBuffer = 4096
	}
	if cfg.MaxLineBytes < 1024 {
		cfg.MaxLineBytes = 4 * 1024 * 1024
	}

	report := Report{StartedAt: time.Now()}
	detector := NewDetector()
	jobs := make(chan lineJob, cfg.ChannelBuffer)
	results := make(chan parsedLine, cfg.ChannelBuffer)
	var wg sync.WaitGroup

	for i := 0; i < cfg.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
				}
				event, ok := detector.Parse(job.text, job.number)
				results <- parsedLine{event: event, ok: ok}
			}
		}()
	}

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), cfg.MaxLineBytes)
	go func() {
		defer close(jobs)
		for scanner.Scan() {
			report.Lines++
			text := scanner.Text()
			select {
			case <-ctx.Done():
				return
			case jobs <- lineJob{number: report.Lines, text: text}:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	events := make([]models.Event, 0, 1024)
	for result := range results {
		if result.ok {
			report.Parsed++
			events = append(events, result.event)
			if report.Format == "" {
				report.Format = detector.Detect(result.event.Raw)
			}
		} else {
			report.Unparsed++
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, report, err
	}
	if ctx.Err() != nil && !errors.Is(ctx.Err(), context.Canceled) {
		return nil, report, ctx.Err()
	}
	sort.SliceStable(events, func(i, j int) bool {
		if events[i].Timestamp.Equal(events[j].Timestamp) {
			return events[i].Line < events[j].Line
		}
		return events[i].Timestamp.Before(events[j].Timestamp)
	})
	if report.Format == "" {
		report.Format = FormatText
	}
	report.FinishedAt = time.Now()
	return events, report, nil
}

func (d *Detector) Detect(line string) Format {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "{") && json.Valid([]byte(trimmed)) {
		return FormatJSON
	}
	if d.springRe.MatchString(line) {
		return FormatSpringBoot
	}
	if d.logback.MatchString(line) {
		return FormatLogback
	}
	if d.custom.MatchString(line) {
		return FormatCustom
	}
	return FormatText
}

func (d *Detector) Parse(line string, number int64) (models.Event, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return models.Event{}, false
	}
	if strings.HasPrefix(trimmed, "{") {
		if event, ok := d.parseJSON(trimmed, number); ok {
			return event, true
		}
	}
	if event, ok := d.parseLogback(line, number); ok {
		return event, true
	}
	if event, ok := d.parseSpring(line, number); ok {
		return event, true
	}
	if event, ok := d.parseCustom(line, number); ok {
		return event, true
	}
	return d.parseText(line, number), true
}

func (d *Detector) parseJSON(line string, number int64) (models.Event, bool) {
	var data map[string]any
	if err := json.Unmarshal([]byte(line), &data); err != nil {
		return models.Event{}, false
	}
	event := models.Event{Raw: line, Line: number, Level: models.LevelInfo, Kind: models.KindGeneric}
	event.Timestamp = parseAnyTime(getString(data, "timestamp", "time", "@timestamp", "ts", "date"))
	event.Level = models.NormalizeLevel(getString(data, "level", "severity", "log.level"))
	event.Thread = getString(data, "thread", "thread_name", "process.thread.name")
	event.RequestID = getString(data, "request_id", "requestId", "req_id", "correlation_id", "correlationId", "x-request-id")
	event.TraceID = getString(data, "trace_id", "traceId", "trace.id")
	event.SpanID = getString(data, "span_id", "spanId", "span.id")
	event.UserID = getString(data, "user_id", "userId", "uid")
	event.SessionID = getString(data, "session_id", "sessionId", "sid")
	event.Logger = getString(data, "logger", "logger_name", "log.logger")
	event.Message = getString(data, "message", "msg", "event", "log")
	if event.Message == "" {
		encoded, _ := json.Marshal(data)
		event.Message = string(encoded)
	}
	event.Duration = parseDuration(getString(data, "duration", "elapsed", "latency"))
	d.enrich(&event)
	return event, true
}

func (d *Detector) parseLogback(line string, number int64) (models.Event, bool) {
	m := d.logback.FindStringSubmatch(line)
	if len(m) == 0 {
		return models.Event{}, false
	}
	event := models.Event{
		Timestamp: parseAnyTime(m[1]),
		Thread:    m[2],
		Level:     models.NormalizeLevel(m[3]),
		Logger:    m[4],
		Message:   strings.TrimSpace(m[5]),
		Raw:       line,
		Line:      number,
		Kind:      models.KindGeneric,
	}
	d.enrich(&event)
	return event, true
}

func (d *Detector) parseSpring(line string, number int64) (models.Event, bool) {
	m := d.springRe.FindStringSubmatch(line)
	if len(m) == 0 {
		return models.Event{}, false
	}
	event := models.Event{
		Timestamp: parseAnyTime(m[1]),
		Level:     models.NormalizeLevel(m[2]),
		Thread:    m[3],
		Logger:    m[4],
		Message:   strings.TrimSpace(m[5]),
		Raw:       line,
		Line:      number,
		Kind:      models.KindGeneric,
	}
	d.enrich(&event)
	return event, true
}

func (d *Detector) parseCustom(line string, number int64) (models.Event, bool) {
	m := d.custom.FindStringSubmatch(line)
	if len(m) == 0 {
		return models.Event{}, false
	}
	event := models.Event{
		Timestamp: parseAnyTime(m[1]),
		Level:     models.NormalizeLevel(m[2]),
		Message:   strings.TrimSpace(m[3]),
		Raw:       line,
		Line:      number,
		Kind:      models.KindGeneric,
	}
	d.enrich(&event)
	return event, true
}

func (d *Detector) parseText(line string, number int64) models.Event {
	event := models.Event{
		Timestamp: time.Now(),
		Level:     models.NormalizeLevel(extractLevel(line)),
		Message:   strings.TrimSpace(line),
		Raw:       line,
		Line:      number,
		Kind:      models.KindGeneric,
	}
	d.enrich(&event)
	return event
}

func (d *Detector) enrich(event *models.Event) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	for _, match := range d.kv.FindAllStringSubmatch(event.Raw+" "+event.Message, -1) {
		key := normalizeKey(match[1])
		value := strings.Trim(match[2], `"`)
		switch key {
		case "requestid":
			event.RequestID = firstNonEmpty(event.RequestID, value)
		case "traceid":
			event.TraceID = firstNonEmpty(event.TraceID, value)
		case "spanid":
			event.SpanID = firstNonEmpty(event.SpanID, value)
		case "userid":
			event.UserID = firstNonEmpty(event.UserID, value)
		case "sessionid":
			event.SessionID = firstNonEmpty(event.SessionID, value)
		case "logger":
			event.Logger = firstNonEmpty(event.Logger, value)
		case "thread":
			event.Thread = firstNonEmpty(event.Thread, value)
		case "duration", "elapsed":
			if event.Duration == 0 {
				event.Duration = parseDuration(value)
			}
		}
	}
	if event.Duration == 0 {
		if m := d.duration.FindStringSubmatch(event.Raw); len(m) > 0 {
			event.Duration = parseDuration(m[1] + m[2])
		}
	}
	event.Kind = classify(event.Message + " " + event.Raw)
	if m := d.httpReq.FindStringSubmatch(event.Message); len(m) > 0 {
		event.Kind = models.KindHTTP
		event.HTTP = &models.HTTPInfo{Method: strings.ToUpper(m[1]), URL: m[2]}
	}
	if m := d.httpResp.FindStringSubmatch(event.Message); len(m) > 0 {
		status, _ := strconv.Atoi(m[1])
		if event.HTTP == nil {
			event.HTTP = &models.HTTPInfo{}
		}
		event.HTTP.Status = status
	}
	if event.Kind == models.KindException || event.Level == models.LevelError || event.Level == models.LevelFatal {
		event.Error = detectError(strings.Join([]string{event.Message, event.Logger, event.Raw}, " "))
	}
	if event.RequestID == "" && event.TraceID == "" && event.UserID == "" && event.SessionID == "" {
		event.RequestID = fmt.Sprintf("line-%d", event.Line)
	}
}

func parseAnyTime(value string) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}
	}
	layouts := []string{
		time.RFC3339Nano,
		"2006-01-02T15:04:05.999999999",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05,999999999",
		"2006-01-02T15:04:05.999",
		"2006-01-02 15:04:05.999",
		"2006-01-02 15:04:05,999",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if ts, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return ts
		}
	}
	if unix, err := strconv.ParseInt(value, 10, 64); err == nil {
		if unix > 1_000_000_000_000 {
			return time.UnixMilli(unix)
		}
		return time.Unix(unix, 0)
	}
	return time.Time{}
}

func parseDuration(value string) time.Duration {
	value = strings.TrimSpace(strings.Trim(value, `"`))
	if value == "" {
		return 0
	}
	if d, err := time.ParseDuration(value); err == nil {
		return d
	}
	re := regexp.MustCompile(`(?i)^([0-9]+(?:\.[0-9]+)?)(ms|s|sec|secs|seconds|m)?$`)
	m := re.FindStringSubmatch(value)
	if len(m) == 0 {
		return 0
	}
	n, _ := strconv.ParseFloat(m[1], 64)
	switch strings.ToLower(m[2]) {
	case "s", "sec", "secs", "seconds":
		return time.Duration(n * float64(time.Second))
	case "m":
		return time.Duration(n * float64(time.Minute))
	default:
		return time.Duration(n * float64(time.Millisecond))
	}
}

func getString(data map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := data[key]; ok {
			switch v := value.(type) {
			case string:
				return v
			case float64:
				return strconv.FormatFloat(v, 'f', -1, 64)
			case bool:
				return strconv.FormatBool(v)
			default:
				return fmt.Sprint(v)
			}
		}
	}
	return ""
}

func classify(text string) models.EventKind {
	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "exception") || strings.Contains(lower, "panic") || strings.Contains(lower, "stack trace") || strings.Contains(lower, "deadlock") || strings.Contains(lower, "timeout"):
		return models.KindException
	case strings.Contains(lower, "select ") || strings.Contains(lower, "insert ") || strings.Contains(lower, "update ") || strings.Contains(lower, "delete ") || strings.Contains(lower, " sql "):
		return models.KindSQL
	case strings.Contains(lower, "kafka"):
		return models.KindKafka
	case strings.Contains(lower, "redis"):
		return models.KindRedis
	case strings.Contains(lower, "http") || strings.Contains(lower, "request") || strings.Contains(lower, "response"):
		return models.KindHTTP
	default:
		return models.KindGeneric
	}
}

func detectError(message string) *models.ErrorInfo {
	info := &models.ErrorInfo{RootCause: strings.TrimSpace(message)}
	known := []string{"NullPointerException", "TimeoutException", "SQLIntegrityConstraintViolationException", "SQLException", "Deadlock", "KafkaException", "RedisException", "HTTP 500", "HTTP 502", "HTTP 503", "Timeout"}
	for _, item := range known {
		if strings.Contains(strings.ToLower(message), strings.ToLower(item)) {
			info.Type = item
			break
		}
	}
	if info.Type == "" {
		fields := strings.Fields(message)
		for _, field := range fields {
			if strings.HasSuffix(field, "Exception") || strings.HasSuffix(field, "Error") {
				info.Type = strings.Trim(field, ":")
				break
			}
		}
	}
	if idx := strings.Index(message, " at "); idx >= 0 {
		info.Location = strings.TrimSpace(message[idx+4:])
	}
	return info
}

func normalizeKey(v string) string {
	replacer := strings.NewReplacer("_", "", "-", "", ".", "")
	return strings.ToLower(replacer.Replace(v))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func extractLevel(line string) string {
	for _, token := range strings.Fields(line) {
		switch strings.ToUpper(strings.Trim(token, "[]:")) {
		case "TRACE", "DEBUG", "INFO", "WARN", "WARNING", "ERROR", "FATAL":
			return token
		}
	}
	return "INFO"
}
