package search

import (
	"testing"

	"github.com/deepawasthi/logtimeline/internal/models"
)

func TestSearchMatchesMessage(t *testing.T) {
	requests := []models.Request{{ID: "req-1", Events: []models.Event{{Message: "Kafka Event Published"}}}}
	results := Search(requests, "kafka")
	if len(results) != 1 {
		t.Fatalf("results = %d, want 1", len(results))
	}
}
