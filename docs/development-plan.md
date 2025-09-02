# Ticker Forge — Development Plan

_A pragmatic roadmap for a Go‑based market analytics **TUI + optional web UI**._  
Owner: Christian (Ticker Forge) • Last updated: 2025-09-02

> **Goals**: keep it fast, keyboard‑first, and simple to maintain. Ship value every sprint.

---

## 0) Guiding Principles
- **Keyboard‑first** UX; mouse is optional (web only).
- **Single binary**: no external runtime required for the app.
- **Security first** for API keys and (future) brokerage integrations.
- **Modular** by design (data sources, renderers, UI panes).
- **Docs or it didn’t happen**: keep MkDocs current; autogenerate API docs.
- **Small, shippable slices**: each milestone delivers visible value.

---

## 1) Architecture (high‑level)
```
cmd/ticker-forge (main)
└─ internal/
   ├─ cli/           # CLI entry, flags, subcommands
   ├─ ui/            # TUI (Bubble Tea), models, views, keymap
   ├─ server/        # web server (Gin), handlers, templates
   ├─ chart/         # renderers (ASCII, web charts) + studies
   ├─ data/          # sources; e.g., yahoo, fmp (interfaces below)
   ├─ cfg/           # config load/save, keyring/env integration
   └─ core/          # app state, domain models, events, caching
```
**Data source contract**:
```go
// internal/data/source.go
type PriceBar struct {
    T int64   // unix ms
    O, H, L, C float64
    V float64
}

type PriceFeed interface {
    Intraday(symbol string, interval string, lookback int) ([]PriceBar, error)
    Daily(symbol string, lookback int) ([]PriceBar, error)
    Quote(symbol string) (map[string]any, error)
    Fundamentals(symbol string) (map[string]any, error) // roadmap
    News(symbol string, limit int) ([]NewsItem, error)   // roadmap
    SourceName() string
}
```
Swap implementations (e.g., `yahoo`, `fmp`) without touching UI.

---

## 2) Milestones & Deliverables

### M1 — Baseline TUI + ASCII Charts (✅ foundation)
**Scope**
- TUI skeleton (Bubble Tea or your current approach) with panes: watchlist, chart, status.
- ASCII line/candlestick renderers.
- Yahoo fetcher (intraday + daily) with simple in‑memory cache.
- Config file (default ticker, theme, watchlists).

**Acceptance Criteria**
- `ticker-forge tui` loads, renders a chart for default ticker, moves via keybindings.
- `ticker-forge chart AAPL` prints ASCII line/candle.
- Config resolved from `~/.config/tickerforge/config.yaml` (Win: `%APPDATA%\tickerforge\`).
- MkDocs site serves locally via `make serve-docs`.

---

### M2 — Web Server (HTMX + go‑echarts)
**Scope**
- Minimal Gin server: `/` dashboard (chart + quote), `/quote/:symbol`, `/chart/:symbol`.
- Static assets (CSS) + templates; embed with `go:embed` for single binary.
- Reuse data layer; ensure identical candles between TUI and Web.

**Acceptance Criteria**
- `ticker-forge web --port 8080` serves chart and latest quote for a symbol.
- Lighthouse sanity: page interactive < 2s locally.
- Docs page **Usage > Web** with screenshots.

---

### M3 — Config & Secrets Hardening
**Scope**
- Config schema versioning; validation with defaults.
- Secrets sources: **env** > **OS keyring** > **plaintext config** (discouraged).
- Helper CLI: `ticker-forge keys set fmp` (stores in keyring), `... keys print` (warns).

**Acceptance Criteria**
- Setting an API key stores in keyring (macOS Keychain, wincred, libsecret).
- App never logs secrets; redact on panic/log.
- Security section in docs; threat model & non‑goals.

---

### M4 — Overlays & Indicators
**Scope**
- Studies (SMA/EMA/RSI/MACD/Bollinger) as composable functions in `internal/chart/studies`.
- **Overlays toggle** in TUI (`i`) with small modal.
- **Macro overlays (roadmap)**: SPX, FF rate, CPI, unemployment (pluggable).

**Acceptance Criteria**
- Toggle 2–3 indicators on/off; chart updates live.
- Unit tests for indicator math.
- Docs page with indicator definitions.

---

### M5 — Watchlists & Persistence
**Scope**
- CRUD watchlists in config or `~/.local/share/tickerforge/db.json` (simple bolt/BBolt later).
- Quick search (`/`) with filtered list.
- Per‑ticker preferences (interval, overlays).

**Acceptance Criteria**
- Add/remove tickers from a watchlist in‑app; persists across runs.
- Search finds by symbol or name substring.
- Docs updated with watchlist format.

---

### M6 — Fundamentals & News (opt‑in keys)
**Scope**
- Fundamentals view (key ratios, valuation snapshot).
- News feed: headline list + open link (web) or preview (TUI).

**Acceptance Criteria**
- Switching to Fundamentals tab fetches & renders key metrics.
- News shows recent headlines per ticker; errors handled gracefully.
- Rate limiting and polite backoffs.

---

### M7 — Packaging, CI, and Releases
**Scope**
- GitHub Actions: lint/test, `goreleaser` (darwin/linux/windows).
- Homebrew tap (later), scoop/choco (optional).
- Docs: install section for each OS.

**Acceptance Criteria**
- Publish versioned binaries on GitHub Releases.
- MkDocs deploy on push to `main` (Pages).
- Changelog generated (Keep a Changelog).

---

## 3) Backlog (nice‑to‑have)
- **Plugin system** for data feeds via `go plugin` (linux‑only) or build tags.
- **Alert rules**: price crosses/indicator events → desktop/OS notification.
- **Export**: CSV/PNG chart export; copy to clipboard.
- **Theme system** for TUI (light/dark/high‑contrast).
- **Multi‑symbol compare** overlay.
- **Offline cache** with TTL & size limits (BoltDB / SQLite).

---

## 4) Security & Privacy
- Key storage priority: **OS keyring** > env vars > config (discouraged).
- Never write secrets to logs; add a redactor logger.
- Respect robots/ToS of data sources; add rate‑limiters + jitter.
- Optional proxy config for corp networks.

---

## 5) Observability & Quality
- **Logging**: leveled logger; `--log-level` flag.
- **Metrics**: internal counters/timers (expose `/metrics` in web mode; optional).
- **Tracing**: simple spans around fetch/render (disabled by default).
- **Tests**: core studies math, data transforms, config loading, handlers.
- **Benchmarks**: ASCII chart render at target sizes (e.g., 120x30 within <10ms/sample).

---

## 6) Docs & DX
- **MkDocs** with `uvx`:
  - `make serve-docs` and `make build-docs` (no venv).
- **API docs**: `gomarkdoc` → `docs/reference/` via `scripts/gen_ref_docs.sh`.
- **Usage pages**: CLI, TUI, Web; screenshots/gifs.
- **ADR notes**: short Architecture Decision Records for key choices (TUI lib, data lib, secrets).

---

## 7) Milestone Timeline (target)
| Milestone | Duration | Notes |
| --- | --- | --- |
| M1 Baseline TUI | 1–2 weeks | solid charts + config |
| M2 Web Server | 1 week | reuse data & chart logic |
| M3 Secrets | 0.5–1 week | keyring & redaction |
| M4 Indicators | 1–1.5 weeks | SMA/EMA/RSI/MACD |
| M5 Watchlists | 0.5–1 week | CRUD + search |
| M6 Fundamentals/News | 1 week | opt‑in APIs |
| M7 Packaging/CI | 0.5 week | goreleaser + Pages |

---

## 8) Definition of Done
- Feature has tests (where practical), docs updated, and help text (`--help`) accurate.
- No plaintext secrets in repo or logs.
- Cross‑platform check (macOS/Linux/Windows) for CLI/TUI basics.
- Reasonable performance for ASCII charts at common terminal sizes.

---

## 9) Risks & Mitigations
- **API terms/rate limits** → cache + backoff + clearly document keys/limits.
- **Cross‑platform TUI quirks** → CI matrix + minimal terminal dependencies.
- **Secret storage differences per OS** → graceful fallback & docs.
- **Scope creep** → milestone gates; keep web UI intentionally simple.

---

## 10) Make Targets (suggested)
```makefile
run-tui:
	./bin/ticker-forge tui

run-web:
	./bin/ticker-forge web --port 8080

gen-ref:
	scripts/gen_ref_docs.sh

serve-docs:
	uvx --from mkdocs-material mkdocs serve

build-docs:
	uvx --from mkdocs-material mkdocs build --strict

test:
	go test ./...
```

---

## 11) Issue Labels (triage)
- `type/feature`, `type/bug`, `type/chore`, `type/docs`
- `area/tui`, `area/web`, `area/data`, `area/chart`, `area/security`
- `good-first-issue`, `help-wanted`, `blocked`

---

## 12) Immediate Next Steps
- [ ] Lock in TUI keybindings in code and mirror in docs.
- [ ] Implement `internal/cfg` with keyring + env precedence.
- [ ] Add `internal/chart/studies` with SMA + EMA + RSI.
- [ ] Create `docs/usage/cli.md` and `docs/usage/tui.md` with GIFs.
- [ ] Set up CI: lint/test + MkDocs build on PRs.
