package timeline

import (
	"sort"
	"sync"
	"time"

	"github.com/deepawasthi/logtimeline/internal/models"
)

type Store struct {
	mu     sync.RWMutex
	events []models.Event
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) Add(event models.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, event)
}

func (s *Store) Timeline() models.Timeline {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return Build(s.events)
}

func Build(events []models.Event) models.Timeline {
	groups := make(map[string]*models.Request)
	for _, event := range events {
		key, groupBy := groupKey(event)
		req := groups[key]
		if req == nil {
			req = &models.Request{ID: key, GroupBy: groupBy, Start: event.Timestamp, End: event.Timestamp, Status: "SUCCESS"}
			groups[key] = req
		}
		req.Events = append(req.Events, event)
		if event.Timestamp.Before(req.Start) {
			req.Start = event.Timestamp
		}
		if event.Timestamp.After(req.End) {
			req.End = event.Timestamp
		}
		if event.Duration > req.Duration {
			req.Duration = event.Duration
		}
		if event.HTTP != nil && event.HTTP.Method != "" && event.HTTP.URL != "" && req.Endpoint == "" {
			req.Endpoint = event.HTTP.Method + " " + event.HTTP.URL
		}
		if event.UserID != "" {
			req.UserID = event.UserID
		}
		if event.SessionID != "" {
			req.SessionID = event.SessionID
		}
		if event.TraceID != "" {
			req.TraceID = event.TraceID
		}
		if event.Level == models.LevelError || event.Level == models.LevelFatal || event.Kind == models.KindException {
			req.Status = "ERROR"
		}
		if event.HTTP != nil && event.HTTP.Status >= 500 {
			req.Status = "ERROR"
		} else if event.HTTP != nil && event.HTTP.Status >= 400 && req.Status != "ERROR" {
			req.Status = "FAILED"
		}
	}

	requests := make([]models.Request, 0, len(groups))
	totalEvents := 0
	for _, req := range groups {
		sort.SliceStable(req.Events, func(i, j int) bool {
			if req.Events[i].Timestamp.Equal(req.Events[j].Timestamp) {
				return req.Events[i].Line < req.Events[j].Line
			}
			return req.Events[i].Timestamp.Before(req.Events[j].Timestamp)
		})
		if req.Duration == 0 && !req.End.IsZero() && !req.Start.IsZero() {
			req.Duration = req.End.Sub(req.Start)
		}
		totalEvents += len(req.Events)
		requests = append(requests, *req)
	}
	sort.SliceStable(requests, func(i, j int) bool {
		if requests[i].Start.Equal(requests[j].Start) {
			return requests[i].ID < requests[j].ID
		}
		return requests[i].Start.Before(requests[j].Start)
	})
	return models.Timeline{Requests: requests, Events: totalEvents}
}

func groupKey(event models.Event) (string, string) {
	switch {
	case event.RequestID != "":
		return event.RequestID, "request_id"
	case event.TraceID != "":
		return event.TraceID, "trace_id"
	case event.UserID != "":
		return event.UserID, "user_id"
	case event.SessionID != "":
		return event.SessionID, "session_id"
	default:
		return event.Timestamp.Format(time.RFC3339Nano), "timestamp"
	}
}
