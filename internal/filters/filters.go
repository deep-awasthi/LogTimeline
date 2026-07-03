package filters

import (
	"strings"

	"github.com/deepawasthi/logtimeline/internal/models"
)

func Apply(requests []models.Request, filter string) []models.Request {
	filter = strings.TrimSpace(strings.ToUpper(filter))
	if filter == "" || filter == "ALL" {
		return requests
	}
	var out []models.Request
	for _, req := range requests {
		if matches(req, filter) {
			out = append(out, req)
		}
	}
	return out
}

func matches(req models.Request, filter string) bool {
	switch filter {
	case "TRACE", "DEBUG", "INFO", "WARN", "WARNING", "ERROR", "FATAL":
		return req.ContainsLevel(models.NormalizeLevel(filter))
	case "SQL", "ONLY SQL":
		return req.ContainsKind(models.KindSQL)
	case "KAFKA", "ONLY KAFKA":
		return req.ContainsKind(models.KindKafka)
	case "REDIS", "ONLY REDIS":
		return req.ContainsKind(models.KindRedis)
	case "REST", "HTTP", "ONLY REST":
		return req.ContainsKind(models.KindHTTP)
	case "EXCEPTION", "EXCEPTIONS", "ONLY EXCEPTIONS":
		return req.ContainsKind(models.KindException)
	case "SUCCESS":
		return req.Status == "SUCCESS"
	case "FAILED":
		return req.Status == "FAILED" || req.Status == "ERROR"
	default:
		return strings.Contains(strings.ToUpper(req.Summary()), filter)
	}
}
