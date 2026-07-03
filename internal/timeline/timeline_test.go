package timeline

import (
	"testing"
	"time"

	"github.com/deepawasthi/logtimeline/internal/models"
)

func TestBuildGroupsByRequestID(t *testing.T) {
	ts := time.Date(2026, 7, 3, 10, 15, 12, 0, time.UTC)
	tl := Build([]models.Event{
		{Timestamp: ts, RequestID: "a", Level: models.LevelInfo, Message: "GET /users", Kind: models.KindHTTP, HTTP: &models.HTTPInfo{Method: "GET", URL: "/users"}},
		{Timestamp: ts.Add(time.Second), RequestID: "a", Level: models.LevelError, Message: "TimeoutException", Kind: models.KindException},
	})
	if len(tl.Requests) != 1 {
		t.Fatalf("requests = %d, want 1", len(tl.Requests))
	}
	req := tl.Requests[0]
	if req.Status != "ERROR" {
		t.Fatalf("status = %s, want ERROR", req.Status)
	}
	if req.Duration != time.Second {
		t.Fatalf("duration = %s, want 1s", req.Duration)
	}
	if req.Endpoint != "GET /users" {
		t.Fatalf("endpoint = %s", req.Endpoint)
	}
}
