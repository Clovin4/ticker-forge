package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"ticker-forge/internal/server"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Mode int

const (
	ModeTUI Mode = iota
	ModeServe
)

type Options struct {
	Mode           Mode
	Port           string
	DefaultSymbol  string
	DefaultRange   string
	DefaultInterval string
}

func Run(opts Options) error {
	if opts.Mode == ModeServe {
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

//
// --- Bubble Tea TUI ---
//
type model struct {
	symbol   string
	port     string
	status   string
	quitting bool
}

var (
	title = lipgloss.NewStyle().Bold(true).MarginBottom(1)
	lbl   = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	ok    = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
)

func initialModel(opts Options) model {
	if opts.Port == "" {
		opts.Port = "8080"
	}
	if opts.DefaultSymbol == "" {
		opts.DefaultSymbol = "AAPL"
	}
	return model{symbol: opts.DefaultSymbol, port: opts.Port}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) View() string {
	if m.quitting {
		return "\n  Bye.\n\n"
	}
	return title.Render("Ticker Forge") +
		fmt.Sprintf("%s Type a ticker and press Enter to launch the web dashboard on http://localhost:%s\n\n", lbl.Render("hint:"), m.port) +
		fmt.Sprintf("Ticker: %s\n\n", m.symbol) +
		fmt.Sprintf("%s %s\n", lbl.Render("status:"), m.status) +
		"\n(Press Ctrl+C to quit)\n"
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			// start server in background, open browser hint
			go func() {
				_ = server.ListenAndServe(server.Options{
					Port:            m.port,
					DefaultSymbol:   m.symbol,
					DefaultRange:    "1d",
					DefaultInterval: "1m",
				})
			}()
			m.status = ok.Render("serving → ") + fmt.Sprintf("http://localhost:%s (symbol=%s)", m.port, m.symbol)
			return m, nil
		default:
			// collect characters for ticker (very simple)
			s := msg.String()
			if len(s) == 1 {
				m.symbol += s
			} else if s == "backspace" && len(m.symbol) > 0 {
				m.symbol = m.symbol[:len(m.symbol)-1]
			}
		}
	}
	return m, nil
}

func runTUI(opts Options) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	p := tea.NewProgram(initialModel(opts))
	go func() {
		<-ctx.Done()
		p.Quit()
	}()
	_, err := p.Run()
	return err
}
