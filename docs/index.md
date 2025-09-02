# Tickity

_A cuteness-first market analytics TUI, with an optional web UI — written in Go._

![Tickity Gopher](static/tickity_gopher.png)

## What is this?

**Tickity** is a fast terminal UI (TUI) for exploring stocks with minimal friction. It also ships a small web server for lightweight charts when you want a browser view.

- ⚡ **TUI**: snappy, keyboard-driven navigation  
- 🖥️ **Optional web UI**: simple HTMX/go-echarts views  
- 🧰 **CLI**: quick commands for quotes and ASCII charts  
- 🔌 **Pluggable data**: e.g., Yahoo Finance fetchers  
- 🧾 **Configurable**: defaults, theme, watchlists

---

## Quickstart (App)

Build and run from source:

```bash
# from repo root
go mod tidy
go build -o bin/ticker-forge ./cmd/ticker-forge
./bin/ticker-forge --help

# Launch terminal UI
./bin/ticker-forge

# Start the web server (default :8080)
./bin/ticker-forge --mode serve --port 8080 --symbol MSFT
```
