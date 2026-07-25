### Project Summary
An AI-powered observability agent, written in Go, that correlates logs from Sumo Logic with metrics and traces from New Relic, then uses an LLM to reason about root cause and produce human-readable incident summaries.

### Motivation 
Operations spend a lot of time context-switching between log and APM tools during an incident, manually correlating timestamps, trace IDs, and service names. Obersavbility AI agent automates that first pass: it pulls the relevant data from both platforms, aligns it, and hands a structured summary to an LLM to draft a root-cause hypothesis — before a human even opens a dashboard.

### Core Flow

- **Trigger** webhook from New Relic alert (or manual CLI query) → Go service
- **Fetch** parallel goroutines pull: Sumo Logic search API (logs around incident window), New Relic NerdGraph (metrics + distributed traces for affected service)
- **Correlate** normalize into a common timeline struct (timestamp, source, severity, service)
- **Reason** send condensed context to an LLM (Claude API) with a structured prompt → get probable root cause + confidence + suggested runbook
- **Output** post to Slack/Teams with summary, or serve via a small CLI/HTTP endpoint

### Architecture
```
┌─────────────┐   ┌─────────────┐
│ Sumo Logic   │   │ New Relic    │
│ (Search API) │   │ (NerdGraph)  │
└──────┬───────┘   └──────┬───────┘
       │  logs             │  metrics/traces
       └────────┬──────────┘
                 ▼
        ┌─────────────────┐
        │ Go Collector     │  polling / webhook triggered
        └────────┬─────────┘
                 ▼
        ┌─────────────────┐
        │ Correlator       │  align by trace_id / timestamp / service
        └────────┬─────────┘
                 ▼
        ┌─────────────────┐
        │ LLM Reasoner     │  root cause, anomaly summary
        └────────┬─────────┘
                 ▼
        ┌─────────────────┐
        │ Output           │  Slack / PagerDuty / incident doc
        └─────────────────┘
```

### Tech Stack
- Language Go 1.22+
- Concurrency goroutines + worker pool
- Tracing OpenTelemetry
- Storage SQLite / Postgres (incident history)
- Deployment Kubernetes (Deployment + webhook listener, or CronJob)
- LLM Claude API

### Project Structure 
```
ObservabilityAgentAI/
├── cmd/
│   └── Obs/      # main entrypoint
├── internal/
│   ├── sumologic/          # Sumo Logic API client
│   ├── newrelic/           # New Relic NerdGraph client
│   ├── correlator/         # event correlation logic
│   ├── reasoner/           # LLM prompt construction + calls
│   ├── output/             # Slack/Teams notifiers
│   └── store/              # incident history persistence
├── pkg/
│   └── models/             # shared Event/Incident structs
├── deploy/
│   └── k8s/                # manifests
├── config/
│   └── config.example.yaml
├── go.mod
└── README.md
```

### Configuration

Copy the example config and fill in credentials:
```
cp config/config.example.yaml config/config.yaml
```
```
sumologic:
  access_id: ""
  access_key: ""
  endpoint: "https://api.sumologic.com/api/v1"

newrelic:
  api_key: ""
  account_id: ""

anthropic:
  api_key: ""
  model: "claude-sonnet-4-6"

output:
  slack_webhook_url: ""
  ```

### Run locally
```
  go build -o sentinel ./cmd/sentinel
  ./Observability --config config/config.yaml
```

### Run on Kubernetes
```
  kubectl apply -f deploy/k8s/
```

### Roadmap
- Phase 1 — Data clients: Sumo Logic + New Relic, independently testable.
- Phase 2 — Correlator: normalize both sources into a common Event struct.
- Phase 3 — Reasoner: structured prompt to Claude, JSON-in/JSON-out.
- Phase 4 — Event-driven trigger via NRQL alert webhook.
- Phase 5 — Slack output with dashboard deep links.
- Phase 6 (stretch) — Incident history + RAG-style feedback loop.
