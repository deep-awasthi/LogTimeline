package stats

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/deepawasthi/logtimeline/internal/models"
)

type Report struct {
	TotalRequests       int
	AverageDuration     time.Duration
	SlowestRequest      models.Request
	FastestRequest      models.Request
	MostCommonException string
	ErrorCount          int
	RequestsPerMinute   float64
	TopEndpoints        []Count
	ExceptionCounts     []Count
}

type Count struct {
	Name  string
	Count int
}

func Compute(requests []models.Request) Report {
	report := Report{TotalRequests: len(requests)}
	if len(requests) == 0 {
		return report
	}
	var total time.Duration
	endpointCounts := map[string]int{}
	exceptionCounts := map[string]int{}
	first := requests[0].Start
	last := requests[0].End
	for i, req := range requests {
		total += req.Duration
		if i == 0 || req.Duration > report.SlowestRequest.Duration {
			report.SlowestRequest = req
		}
		if i == 0 || req.Duration < report.FastestRequest.Duration {
			report.FastestRequest = req
		}
		if req.Status == "ERROR" || req.Status == "FAILED" {
			report.ErrorCount++
		}
		if req.Endpoint != "" {
			endpointCounts[req.Endpoint]++
		}
		if req.Start.Before(first) {
			first = req.Start
		}
		if req.End.After(last) {
			last = req.End
		}
		for _, event := range req.Events {
			if event.Error != nil {
				name := event.Error.Type
				if name == "" {
					name = "Unknown"
				}
				exceptionCounts[name]++
			}
		}
	}
	report.AverageDuration = total / time.Duration(len(requests))
	window := last.Sub(first).Minutes()
	if window <= 0 {
		report.RequestsPerMinute = float64(len(requests))
	} else {
		report.RequestsPerMinute = float64(len(requests)) / window
	}
	report.TopEndpoints = top(endpointCounts, 5)
	report.ExceptionCounts = top(exceptionCounts, 5)
	if len(report.ExceptionCounts) > 0 {
		report.MostCommonException = report.ExceptionCounts[0].Name
	}
	return report
}

func (r Report) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Total requests: %d\n", r.TotalRequests)
	fmt.Fprintf(&b, "Average duration: %s\n", models.FormatDuration(r.AverageDuration))
	if r.SlowestRequest.ID != "" {
		fmt.Fprintf(&b, "Slowest request: %s (%s)\n", r.SlowestRequest.ID, models.FormatDuration(r.SlowestRequest.Duration))
	}
	if r.FastestRequest.ID != "" {
		fmt.Fprintf(&b, "Fastest request: %s (%s)\n", r.FastestRequest.ID, models.FormatDuration(r.FastestRequest.Duration))
	}
	fmt.Fprintf(&b, "Most common exception: %s\n", valueOrNone(r.MostCommonException))
	fmt.Fprintf(&b, "Error count: %d\n", r.ErrorCount)
	fmt.Fprintf(&b, "Requests per minute: %.2f\n", r.RequestsPerMinute)
	b.WriteString("Top endpoints:\n")
	for _, item := range r.TopEndpoints {
		fmt.Fprintf(&b, "  %s: %d\n", item.Name, item.Count)
	}
	if len(r.TopEndpoints) == 0 {
		b.WriteString("  none\n")
	}
	b.WriteString("Exceptions:\n")
	for _, item := range r.ExceptionCounts {
		fmt.Fprintf(&b, "  %s: %d\n", item.Name, item.Count)
	}
	if len(r.ExceptionCounts) == 0 {
		b.WriteString("  none\n")
	}
	return b.String()
}

func top(values map[string]int, limit int) []Count {
	counts := make([]Count, 0, len(values))
	for name, count := range values {
		counts = append(counts, Count{Name: name, Count: count})
	}
	sort.SliceStable(counts, func(i, j int) bool {
		if counts[i].Count == counts[j].Count {
			return counts[i].Name < counts[j].Name
		}
		return counts[i].Count > counts[j].Count
	})
	if len(counts) > limit {
		counts = counts[:limit]
	}
	return counts
}

func valueOrNone(value string) string {
	if value == "" {
		return "none"
	}
	return value
}
