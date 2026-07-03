# LogTimeline

LogTimeline is a production-quality terminal application for turning noisy backend logs into request-oriented timelines. It parses Spring Boot, Logback, JSON, custom timestamped text, and simple text logs, then groups events by request ID, trace ID, user ID, or session ID.

## 1. Prerequisites

Install Go 1.25 or newer:

```sh
go version
```

Optional but recommended:

```sh
make --version
git --version
```

## 2. Get the Project

Clone the repository and enter the project directory:

```sh
git clone <repository-url>
cd LogTimeline
```

If you already have the project locally:

```sh
cd /Users/deepawasthi/Downloads/LogTimeline
```

## 3. Install Dependencies

Download and verify Go module dependencies:

```sh
go mod tidy
go mod verify
```

## 4. Build the App

Build with Make:

```sh
make build
```

Or build directly with Go:

```sh
go build -o bin/logtimeline .
```

Confirm the binary works:

```sh
./bin/logtimeline --help
```

## 5. Run with the Sample Log

Open the included sample log in the interactive UI:

```sh
./bin/logtimeline open sample-logs/application.log
```

Use the keyboard shortcuts in the UI:

| Key | Action |
| --- | --- |
| Up / Down | Move through requests |
| Enter | Open selected request |
| Esc | Go back |
| / | Search |
| f | Filter |
| e | Export visible timeline to `timeline.json` |
| s | Statistics |
| r | Refresh sort |
| q | Quit |

## 6. Run Against Your Own Logs

Open a log file:

```sh
./bin/logtimeline open application.log
```

Follow a log file live:

```sh
./bin/logtimeline tail application.log
```

Search requests and events:

```sh
./bin/logtimeline search kafka --input application.log
./bin/logtimeline search "POST /users" --input application.log
./bin/logtimeline search TimeoutException --input application.log
```

Filter by level or event kind:

```sh
./bin/logtimeline filter ERROR --input application.log
./bin/logtimeline filter SQL --input application.log
./bin/logtimeline filter Kafka --input application.log
./bin/logtimeline filter Redis --input application.log
./bin/logtimeline filter REST --input application.log
./bin/logtimeline filter Exceptions --input application.log
```

Print statistics:

```sh
./bin/logtimeline stats --input application.log
```

Export timelines:

```sh
./bin/logtimeline export timeline.json --input application.log
./bin/logtimeline export timeline.md --input application.log
./bin/logtimeline export timeline.csv --input application.log
```

Force an export format if the output extension is unusual:

```sh
./bin/logtimeline export timeline.out --format json --input application.log
```

## 7. Install Globally

Install from the repository checkout:

```sh
go install .
```

Then run:

```sh
logtimeline open sample-logs/application.log
```

When published as a module, install with:

```sh
go install github.com/deepawasthi/logtimeline@latest
```

## 8. Supported Log Inputs

LogTimeline automatically detects common formats:

```text
2026-07-03 10:15:12 INFO Request received POST /users request_id=7ab45de
2026-07-03 10:15:13 INFO SQL INSERT users request_id=7ab45de duration=45ms
2026-07-03 10:15:14 ERROR KafkaException publish failed request_id=7ab45de
```

```json
{"timestamp":"2026-07-03T10:17:00Z","level":"INFO","request_id":"json-1","trace_id":"trace-abc","logger":"api.PaymentController","message":"POST /payments"}
```

Detected fields include:

| Field | Examples |
| --- | --- |
| Timestamp | `timestamp`, `time`, `@timestamp`, `ts` |
| Level | `INFO`, `DEBUG`, `WARN`, `ERROR`, `FATAL` |
| Request ID | `request_id`, `requestId`, `req_id`, `correlation_id`, `x-request-id` |
| Trace ID | `trace_id`, `traceId`, `trace.id` |
| Span ID | `span_id`, `spanId`, `span.id` |
| User ID | `user_id`, `userId`, `uid` |
| Session ID | `session_id`, `sessionId`, `sid` |
| Duration | `duration`, `elapsed`, `latency`, `took` |
| Logger | `logger`, `logger_name`, `log.logger` |

## 9. Timeline Grouping

Events are grouped in this order:

1. `request_id`
2. `trace_id`
3. `user_id`
4. `session_id`
5. timestamp fallback for uncorrelated lines

Each request timeline includes:

1. request ID
2. grouping field
3. start time
4. end time
5. duration
6. status
7. endpoint
8. ordered events
9. detected errors and exceptions

## 10. Configuration

Create `.logtimeline.yaml` in the project directory or your home directory:

```yaml
log_level: warn
parser:
  workers: 8
  channel_buffer: 8192
  max_line_bytes: 4194304
  stack_window: 2s
tail:
  poll_interval: 500ms
  from_end: false
ui:
  page_size: 30
```

Use a specific config file:

```sh
./bin/logtimeline --config .logtimeline.yaml open application.log
```

## 11. Development Workflow

Format code:

```sh
make fmt
```

Run unit tests:

```sh
make test
```

Run benchmarks:

```sh
make benchmark
```

Run static checks:

```sh
make lint
```

Clean build artifacts:

```sh
make clean
```

## 12. Full Verification Checklist

Before committing or releasing, run:

```sh
go mod tidy
gofmt -w $(find . -name '*.go' -not -path './.git/*')
go test ./...
go vet ./...
make build
./bin/logtimeline stats --input sample-logs/application.log
./bin/logtimeline search kafka --input sample-logs/application.log
./bin/logtimeline export /tmp/logtimeline-smoke.json --input sample-logs/application.log
```

## 13. Project Structure

```text
cmd/                    Cobra CLI commands
internal/config/        Viper configuration loading
internal/export/        JSON, Markdown, and CSV exporters
internal/filters/       Level and kind filters
internal/models/        Shared domain models
internal/parser/        Streaming parser and format detection
internal/search/        Request and event search
internal/stats/         Request statistics
internal/tail/          Live file tailing
internal/timeline/      Request grouping and timeline construction
internal/ui/            Bubble Tea terminal UI
sample-logs/            Example logs
.github/workflows/      CI workflow
```

## 14. Performance Notes

The parser uses buffered scanning, worker goroutines, and channels. This keeps parsing responsive for large files while preserving deterministic timeline ordering after parse completion.

For very large logs:

1. Increase `parser.workers` to match available CPU.
2. Increase `parser.channel_buffer` for high-throughput disks.
3. Increase `parser.max_line_bytes` if logs contain large JSON payloads or stack traces.
4. Use `search`, `filter`, `stats`, and `export` when you do not need the full interactive UI.
5. Use `tail` for live investigation.

## 15. CI

The GitHub Actions workflow runs on pushes to `main` and pull requests:

```sh
go test ./...
go vet ./...
```
