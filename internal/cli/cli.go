package cli

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
	"unicode"

	"ticker-forge/internal/chart"
	"ticker-forge/internal/server"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Mode int

const (
	ModeTUI Mode = iota
	ModeServe
)

type Options struct {
	Mode            Mode
	Port            string
	DefaultSymbol   string
	DefaultRange    string
	DefaultInterval string
	// Auto-refresh seconds in TUI (0 = off)
	RefreshSeconds int
}

func Run(opts Options) error {
	if opts.Mode == ModeServe {
		// Serve mode uses the web server; keep as-is in your project
		return serve(opts)
	}
	return runTUI(opts)
}

func serve(opts Options) error {
	return server.ListenAndServe(server.Options{
		Port:            opts.Port,
		DefaultSymbol:   opts.DefaultSymbol,
		DefaultRange:    opts.DefaultRange,
		DefaultInterval: opts.DefaultInterval,
	})
}


/* ---------------- TUI MODEL ---------------- */

type ViewMode int
const (
	ViewLine ViewMode = iota
	ViewCandles
)

type model struct {
	symbol   string
	rng      string
	interval string

	width  int
	height int

	loading   bool
	err       error
	times     []time.Time
	closes    []float64
	lastFetch time.Time

	// UI bits
	inputMode bool
	input     textinput.Model

	// refresh
	refreshEvery time.Duration
	ticker       *time.Ticker
	cancel       context.CancelFunc

	ticks []chart.Tick

	view ViewMode
}

var (
	titleStyle = lipgloss.NewStyle().Bold(true)
	subtle     = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)
	hintStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Italic(true)
)

func initialModel(opts Options) model {
	if opts.DefaultSymbol == "" {
		opts.DefaultSymbol = "AAPL"
	}
	if opts.DefaultRange == "" {
		opts.DefaultRange = "1d"
	}
	if opts.DefaultInterval == "" {
		opts.DefaultInterval = "1m"
	}
	ti := textinput.New()
	ti.Prompt = "> "
	ti.Placeholder = "Ticker Symbol (e.g. AAPL)"
	ti.CharLimit = 16
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	ti.TextStyle = lipgloss.NewStyle().Bold(true)
	ti.Validate = func(s string) error {
		// allow letters, digits, dot, hyphen; empty is allowed while typing
		for _, r := range s {
			if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '-' {
				continue
			}
			return fmt.Errorf("invalid char: %q", r)
		}
		return nil
	}

	var refresh time.Duration
	if opts.RefreshSeconds > 0 {
		refresh = time.Duration(opts.RefreshSeconds) * time.Second
	}

	return model{
		symbol:       strings.ToUpper(opts.DefaultSymbol),
		rng:          opts.DefaultRange,
		interval:     opts.DefaultInterval,
		input:        ti,
		refreshEvery: refresh,
		loading:      true,
	}
}

func (m model) Init() tea.Cmd {
	return fetchCmd(m.symbol, m.rng, m.interval)
}

type fetchedMsg struct {
	times  []time.Time
	closes []float64
	err    error
}

func fetchCmd(symbol, rng, interval string) tea.Cmd {
	return func() tea.Msg {
		t, c, err := chart.FetchIntraday(symbol, rng, interval)
		return fetchedMsg{times: t, closes: c, err: err}
	}
}

type tickMsg struct{}

func tickCmd(d time.Duration) tea.Cmd {
	if d <= 0 {
		return nil
	}
	return tea.Tick(d, func(time.Time) tea.Msg { return tickMsg{} })
}

// makeTicksFromCloses builds OHLC from consecutive closes (prev→open; hi/lo = max/min)
func makeTicksFromCloses(ts []time.Time, closes []float64) []chart.Tick {
	if len(closes) < 2 {
		return nil
	}
	n := len(closes) - 1
	out := make([]chart.Tick, 0, n)
	for i := 1; i < len(closes); i++ {
		o := closes[i-1]
		c := closes[i]
		h, l := o, o
		if c > h { h = c }
		if c < l { l = c }
		var t time.Time
		if i < len(ts) { t = ts[i] }
		out = append(out, chart.Tick{T: t, O: o, H: h, L: l, C: c})
	}
	return out
}


func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.inputMode {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				val := strings.ToUpper(strings.TrimSpace(m.input.Value()))
				m.input.Blur()
				m.inputMode = false
				if val != "" && val != m.symbol {
					m.symbol = val
					m.loading = true
					return m, fetchCmd(m.symbol, m.rng, m.interval)
				}
				return m, nil
			case "esc":
				m.input.Blur()
				m.inputMode = false
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		m.input.SetValue(strings.ToUpper(m.input.Value()))
		return m, cmd
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case fetchedMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.times = msg.times
			m.closes = msg.closes
			m.lastFetch = time.Now()
			m.ticks = makeTicksFromCloses(m.times, m.closes)
			if len(m.closes) < 2 {
				// keep a helpful status instead of trying to render
				m.err = fmt.Errorf("no datapoints returned (try another interval/range)")
			}	
		}
		// keep ticking if enabled
		return m, tickCmd(m.refreshEvery)

	case tickMsg:
		// periodic refresh
		m.loading = true
		return m, fetchCmd(m.symbol, m.rng, m.interval)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Sequence(tea.ExitAltScreen, tea.Quit)

		case "r": // refresh now
			m.loading = true
			return m, fetchCmd(m.symbol, m.rng, m.interval)

		case "/": // edit ticker
			m.inputMode = true
			m.input.SetValue(m.symbol)
			m.input.CursorEnd()
			m.input.Focus()
			return m, nil
			
		case "1":
			m.interval = "1m"
			m.loading = true
			return m, fetchCmd(m.symbol, m.rng, m.interval)
		case "2":
			m.interval = "5m"
			m.loading = true
			return m, fetchCmd(m.symbol, m.rng, m.interval)
		case "3":
			m.interval = "15m"
			m.loading = true
			return m, fetchCmd(m.symbol, m.rng, m.interval)
		case "d":
			m.rng = "1d"
			m.loading = true
			return m, fetchCmd(m.symbol, m.rng, m.interval)
		case "w":
			m.rng = "5d"
			m.loading = true
			return m, fetchCmd(m.symbol, m.rng, m.interval)
		case "c":
			if m.view == ViewLine {
				m.view = ViewCandles
			} else {
				m.view = ViewLine
			}
			return m, nil
		}
		return m, nil
	}
	return m, nil
}

func (m model) View() string {
	// header
	header := titleStyle.Render("Ticker Forge") + "\n" +
		fmt.Sprintf("%s  %s  %s  %s  %s  %s\n",
		  subtle.Render("(/) change ticker"),
		  subtle.Render("[1]=1m"),
		  subtle.Render("[2]=5m"),
		  subtle.Render("[3]=15m"),
		  subtle.Render("[d]=1d, [w]=5d"),
		  subtle.Render("[c]=candles/line"),
		)
	// input mode
	if m.inputMode {
		return header + "\n" +
			"Symbol: " + m.input.View() + "\n\n" +
			hintStyle.Render("Press Enter to apply, Esc to cancel")
	}

	// error / loading
	if m.err != nil {
		return header + "\n" + errStyle.Render("error: "+m.err.Error()) + "\n"
	}
	if m.loading {
		return header + "\n" + hintStyle.Render("loading…") + "\n"
	}
	if len(m.closes) < 2 {
		return header + "\n" + hintStyle.Render("no data yet (try 'r' to refresh or change ticker with '/')") + "\n"
	}
	

	// shared bits for either view
	w := m.width
	h := m.height
	if w <= 0 {
		w = 100
	}
	if h <= 0 {
		h = 30
	}
	last := m.closes[len(m.closes)-1]
	caption := fmt.Sprintf("%s  %s/%s   last: %.2f   fetched: %s",
		m.symbol, m.rng, m.interval, last, m.lastFetch.Format("15:04:05"))
	footer := "\n" + hintStyle.Render("r=refresh • /=ticker • c=candles/line • q=quit")

	if m.view == ViewCandles {
		if len((m.ticks)) < 2 {
			return header + "\n" + hintStyle.Render("no ticks to render yet") + "\n"
		}
		return chart.RenderCandlesASCII(m.ticks, w, h, header, caption, footer)
	}
	return chart.RenderLineASCII(m.closes, w, h, header, caption, footer)
}


func runTUI(opts Options) error {
	model := initialModel(opts)
	log.Printf("Model: %+v\n", model)

	altScreen := tea.WithAltScreen()
	log.Printf("AltScreen Option: %+v\n", altScreen)

	p := tea.NewProgram(model, altScreen)
	log.Printf("Program: %+v\n", p)
	_, err := p.Run()
	return err
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
