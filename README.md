# Health Check on Go

Welcome to the community-driven playground for automating service health checks with Go.  
This repo mixes a Go scheduler/runtime, a lightweight Fiber API, and a React dashboard to help teams experiment with HTTP and SFTP availability checks while the project is still in a heavy, fast-moving development phase.

## Why this exists
- **Cron-powered HTTP monitors** – define hosts + endpoints and let `robfig/cron` run them on custom schedules, storing every run under `results/`.
- **Pluggable SFTP probes** – verify auth, stat/list/read remote paths, and display colorful CLI summaries powered by `pkg/sftp` + `fatih/color`.
- **Self-service config API**
- **React admin (frontend/)** – Vite + React Router UI that speaks to the API (CORS-ready) for editing checks and reviewing statuses.
- **Single binary CLI** – `go run . <command>` (or build) to run schedulers, fire off SFTP tests, or host the management API.

> **Heads up:** Everything here is still evolving. Interfaces may change and bugs are expected—please open issues or PRs when you bump into them.

## Repository tour
| Path | Purpose |
|------|---------|
| `main.go` | Loads config, wires HTTP/SFTP services, hands off to the Cobra CLI (`cmd.go`). |
| `configs/health_checks.json` | Default JSON file that drives both HTTP and SFTP monitors. |
| `libs/health_check/` | Core engines: cron HTTP runner, SFTP command harness, shared structs. |
| `libs/server/` | Fiber server (port 3000) exposing config + status endpoints consumed by the UI. |
| `frontend/` | React Router app for managing checks and visualizing `results/`. |
| `results/` | Auto-generated run artifacts (`results/<path>/<timestamp>` + `index` history). |

## Quick start
1. **Install prerequisites**
   - Go 1.24+
   - Node 20+/npm (or Bun) for `frontend/`
2. **Clone & install deps**
   ```bash
   go mod download
   cd frontend && npm install
   ```
3. **Configure monitors** – edit `configs/health_checks.json` (see below).
4. **Run schedulers**
   ```bash
   go run . health-check    # kicks off HTTP cron jobs (and SFTP checks when enabled)
   ```
5. **Launch the API server** (for UI + external integrations)
   ```bash
   go run . server          # serves on http://localhost:3000
   ```
6. **Start the React dashboard**
   ```bash
   cd frontend
   npm run dev              # Vite dev server on http://localhost:5173 (CORS already allowed)
   ```

### CLI commands
| Command | What it does |
|---------|--------------|
| `go run . health-check` | Runs every configured HTTP cron + SFTP test concurrently and keeps the process alive. |
| `go run . server` | Starts the Fiber API (`/config`, `/statuses`) so the UI or scripts can manage configs/results. |

## Configuration reference (`configs/health_checks.json`)
```json
{
  "http": [
    {
      "name": "Cat facts",
      "host": "https://catfact.ninja/",
      "endpoints": [
        {
          "path": "facts",
          "method": "GET",
          "expectedStatus": 200,
          "schedule": "*/1 * * * *"
        }
      ]
    }
  ],
  "sftp": [
    {
      "name": "Docs bucket",
      "host": "sftp.example.com",
      "port": 22,
      "username": "ci",
      "password": "secret",
      "path": "/var/log/app.log",
      "mode": "read"
    }
  ]
}
```

- **HTTP entries** map a `host` to one or more `endpoints`. Each endpoint can override method, headers (via code), schedule (cron syntax), and expected status. Results land in `results/<path>/` as JSON payloads with status, body snippet, and timestamp.
- **SFTP entries** describe how to connect and what to do (`stat`, `list`, `read`). Commands run sequentially per server, showing PASS/FAIL, latency, and optional assertions (coming soon).

### REST API surface (Fiber)
| Route | Method | Description |
|-------|--------|-------------|
| `/config` | GET | Dumps the current config file (HTTP + SFTP arrays). |
| `/config/http` | POST | Body-parsed JSON `HTTPDomainConfig` is validated and appended, then saved. |
| `/config/sftp` | POST | Accepts form data for new SFTP checks and persists it. |
| `/statuses` | GET | Reads every `results/*/index` file and returns arrays of historical `Result` entries. |