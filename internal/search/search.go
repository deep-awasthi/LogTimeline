package search

import (
	"strings"

	"github.com/deepawasthi/logtimeline/internal/models"
)

func Search(requests []models.Request, query string) []models.Request {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return requests
	}
	var out []models.Request
	for _, req := range requests {
		if requestMatches(req, query) {
			out = append(out, req)
		}
	}
	return out
}

func requestMatches(req models.Request, query string) bool {
	haystack := strings.ToLower(strings.Join([]string{req.ID, req.GroupBy, req.Endpoint, req.UserID, req.SessionID, req.TraceID, req.Status}, " "))
	if strings.Contains(haystack, query) {
		return true
	}
	for _, event := range req.Events {
		parts := []string{
			string(event.Level), string(event.Kind), event.Thread, event.RequestID, event.TraceID,
			event.SpanID, event.UserID, event.SessionID, event.Logger, event.Message, event.Raw,
		}
		if event.HTTP != nil {
			parts = append(parts, event.HTTP.Method, event.HTTP.URL)
		}
		if event.Error != nil {
			parts = append(parts, event.Error.Type, event.Error.RootCause, event.Error.Location, event.Error.StackTrace)
		}
		if strings.Contains(strings.ToLower(strings.Join(parts, " ")), query) {
			return true
		}
	}
	return false
}
