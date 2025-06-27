# Stampede Shooter – Load-Test Tool Specification

## 1. Purpose
A single-binary, command-line utility that spins up thousands of lightweight virtual users, each of which signs in (via a header-based credential), performs a configurable sequence of HTTP requests against your streaming site, measures performance, and prints a concise report.

## 2. Why Go?
- **Massive concurrency**: Goroutines require only a few kB each, enabling hundreds of thousands of parallel requests.
- **Minimal runtime**: Cross-compiles to macOS, Linux, Windows with no external dependencies.
- **Rich standard library**: HTTP, flags, JSON, time, crypto.
- **Ecosystem**: Mature libraries for CLI, histograms, YAML.

## 3. High-Level Architecture
```text
┌──────────────┐
│ CLI / Config │  ← parse flags or YAML/JSON
└───┬──────────┘
    ▼
┌─────────┐       spawn N workers (goroutines)
│ Orchestr│──────┐
└─────────┘      │
       ▲         │ each worker
       │         ▼
┌─────────────┐  ┌────────────────┐
│ Metrics Chan │←│ HTTP Worker    │
└─────────────┘  │  • optional login call
       ▲         │  • loop request script
       │         └────────────────┘
       ▼
┌────────────────┐
│ Reporter       │  ← aggregates & prints live / final stats
└────────────────┘
```

## 4. Command-Line Interface
```bash
stampede-shooter \
  --users 500 \
  --rps 50            # per virtual user
  --duration 120s     # total test length
  --login-url https://site/auth \
  --login-hdr "x-api-token:XYZ" \
  --script ./actions.yml \
  --out report.json
```

### Key Flags
| Flag | Type | Description |
|------|------|-------------|
| `--users` | int | Concurrent virtual users |
| `--rps` | int | Max requests/second *per user* |
| `--duration` | time | How long to keep users active |
| `--script` | file | Ordered list of HTTP actions with placeholders |
| `--login-url` | string | Optional separate login endpoint |
| `--login-hdr` | string | Header(s) to include in every request post-login |
| `--out` | file | Write JSON/CSV in addition to stdout table |
| `--insecure-tls` | bool | Skip TLS verification (for staging) |

## 5. Request Script Format (YAML)
```yaml
- name: Catalog
  method: GET
  url: https://site/api/v1/catalog?page={{randInt 1 20}}
  expect_status: 200
- name: StreamInit
  method: POST
  url: https://site/api/v1/stream/start
  json_body: |
    {"movie_id": "{{pick movies}}"}
  expect_status: 201
```
Built-in template helpers: `randInt`, `pick`, `epochms`, `userId`, etc. Per-action timeout and expected status code.

## 6. Metrics Captured per Request
- Start / end timestamps
- HTTP method & URL alias
- Status code
- Bytes transferred
- Latency (µs)
- Error category (timeout, non-2xx, network)

### Aggregations (computed centrally)
- p50 / p90 / p99 latency (HDR Histogram)
- Requests/s, failures/s
- Bandwidth MB/s
- Per-action breakdown table

## 7. Reporting
Terminal (end of test):
```text
Action        OK   ERR   p50   p90   p99   rps
──────────── ──── ──── ───── ───── ───── ────
Catalog       9k     0   34ms  75ms  180ms 75
StreamInit    3k    12   82ms 144ms 390ms 24

Totals: 12 012 requests, 99.9 % success, 120 s, 108.3 rps, avg 58 ms
```
- Optional JSON/CSV with raw aggregates for dashboards.
- Live progress bar every second (users active, current rps, error %) when `--verbose`.

## 8. Core Packages / Files
| File | Responsibility |
|------|---------------|
| `main.go` | Flag parsing, config load, start orchestrator |
| `orchestrator.go` | Coordinate workers, stop after duration |
| `worker.go` | Executes login + scripted actions, rate-limits |
| `script.go` | Load/parse templates, expand per request |
| `metrics.go` | Typed struct + channel, HDR histograms |
| `reporter.go` | Live & final summaries, file output |
| `auth.go` | Header store, per-user token wiring |
| `util/ratelimit.go` | Token-bucket implementation |

## 9. Concurrency & Safety
- Each worker owns its `http.Client` (with tuned `Transport`: `MaxIdleConns`, `DisableCompression`).
- Metrics sent over buffered channel to avoid worker blocking.
- `context.Context` with cancel propagates early termination (Ctrl-C).
- `sync.WaitGroup` ensures all goroutines finish before final report.

## 10. Dependencies (≈ 1 MiB compiled-in)
- `github.com/spf13/pflag` – POSIX/GNU flag parsing
- `github.com/HdrHistogram/hdrhistogram-go` – latency percentiles
- `gopkg.in/yaml.v3` – script parsing
(All vendored via `go mod vendor`.)

## 11. Extensibility Hooks
- Plugin interface (`type ActionPlugin`)—load Go plugins or subprocess hooks for custom steps (WebSocket, gRPC, etc.).
- Future Prometheus exporter (`--prometheus :9100`) to feed Grafana.
- Optional OpenTelemetry trace push toggled by flag.

## 12. Build & Run
```bash
# fetch deps
$ go mod download

# build static binary
$ go build -o stampede-shooter ./cmd/shooter

# run smoke test
$ ./stampede-shooter --users 100 --script smoke.yml --duration 30s
```
The resulting static binary (~6 MB) can push >100 k rps from a modern laptop while remaining easy to extend and maintain. 

