package models

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type Level string

const (
	LevelTrace Level = "TRACE"
	LevelDebug Level = "DEBUG"
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
	LevelFatal Level = "FATAL"
)

type EventKind string

const (
	KindGeneric   EventKind = "generic"
	KindHTTP      EventKind = "http"
	KindSQL       EventKind = "sql"
	KindKafka     EventKind = "kafka"
	KindRedis     EventKind = "redis"
	KindException EventKind = "exception"
)

type Event struct {
	Timestamp time.Time     `json:"timestamp"`
	Level     Level         `json:"level"`
	Thread    string        `json:"thread,omitempty"`
	RequestID string        `json:"request_id,omitempty"`
	TraceID   string        `json:"trace_id,omitempty"`
	SpanID    string        `json:"span_id,omitempty"`
	UserID    string        `json:"user_id,omitempty"`
	SessionID string        `json:"session_id,omitempty"`
	Logger    string        `json:"logger,omitempty"`
	Message   string        `json:"message"`
	Duration  time.Duration `json:"duration,omitempty"`
	Kind      EventKind     `json:"kind"`
	HTTP      *HTTPInfo     `json:"http,omitempty"`
	Error     *ErrorInfo    `json:"error,omitempty"`
	Raw       string        `json:"raw"`
	Line      int64         `json:"line"`
}

type HTTPInfo struct {
	Method string `json:"method,omitempty"`
	URL    string `json:"url,omitempty"`
	Status int    `json:"status,omitempty"`
}

type ErrorInfo struct {
	Type       string `json:"type,omitempty"`
	RootCause  string `json:"root_cause,omitempty"`
	StackTrace string `json:"stack_trace,omitempty"`
	Location   string `json:"location,omitempty"`
}

type Request struct {
	ID        string        `json:"id"`
	GroupBy   string        `json:"group_by"`
	Events    []Event       `json:"events"`
	Start     time.Time     `json:"start"`
	End       time.Time     `json:"end"`
	Duration  time.Duration `json:"duration"`
	Status    string        `json:"status"`
	Endpoint  string        `json:"endpoint,omitempty"`
	UserID    string        `json:"user_id,omitempty"`
	SessionID string        `json:"session_id,omitempty"`
	TraceID   string        `json:"trace_id,omitempty"`
}

func (r Request) Summary() string {
	if r.Endpoint != "" {
		return r.Endpoint
	}
	if len(r.Events) == 0 {
		return ""
	}
	msg := strings.TrimSpace(r.Events[0].Message)
	if len(msg) > 96 {
		return msg[:93] + "..."
	}
	return msg
}

func (r Request) ContainsLevel(level Level) bool {
	for _, event := range r.Events {
		if event.Level == level {
			return true
		}
	}
	return false
}

func (r Request) ContainsKind(kind EventKind) bool {
	for _, event := range r.Events {
		if event.Kind == kind {
			return true
		}
	}
	return false
}

type Timeline struct {
	Requests []Request `json:"requests"`
	Events   int       `json:"events"`
}

func (t Timeline) Sorted() Timeline {
	out := t
	out.Requests = append([]Request(nil), t.Requests...)
	sort.SliceStable(out.Requests, func(i, j int) bool {
		return out.Requests[i].Start.Before(out.Requests[j].Start)
	})
	return out
}

func NormalizeLevel(v string) Level {
	switch strings.ToUpper(strings.TrimSpace(v)) {
	case "TRACE":
		return LevelTrace
	case "DEBUG":
		return LevelDebug
	case "WARN", "WARNING":
		return LevelWarn
	case "ERROR", "ERR":
		return LevelError
	case "FATAL", "CRITICAL", "PANIC":
		return LevelFatal
	default:
		return LevelInfo
	}
}

func (d Duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

type Duration time.Duration

func FormatDuration(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return d.Round(100 * time.Millisecond).String()
}
