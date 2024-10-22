package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fgrosse/cli"
)

const (
	padding  = 2
	maxWidth = 80
)

var (
	textStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#5A56E0")).Render
	helpStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render
	progressGradient = [2]string{"#5A56E0", "#EE6FF8"}
)

func main() {
	n := flag.Int("workers", runtime.NumCPU(), "Number of workers to run")
	flag.Parse()
	log.SetPrefix("")
	log.SetFlags(0)

	ctx := cli.Context()
	m := NewModel(ctx, *n)

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		log.Fatalf("ERROR: %v", err)
	}

}

type Model struct {
	numWorkers  int
	progress    progress.Model
	stopWorkers context.CancelFunc
}

type tickMsg time.Time

func NewModel(ctx context.Context, numWorkers int) Model {
	ctx, cancel := context.WithCancel(cli.Context())

	m := Model{
		progress: progress.New(
			progress.WithGradient(progressGradient[0], progressGradient[1]),
			progress.WithoutPercentage(),
		),
		numWorkers:  numWorkers,
		stopWorkers: cancel,
	}

	go runWorkers(ctx, m.numWorkers)

	return m
}

func (m Model) Init() tea.Cmd {
	return tickCmd()
}

func runWorkers(ctx context.Context, n int) {
	// log.Printf("Stressing CPU with %d workers", n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go worker(ctx, &wg)
	}

	wg.Wait()
}

func worker(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			fibonacci(35)
		}

	}
}

// Fibonacci function without recursion, using iteration instead.
func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	prev, curr := 0, 1
	for i := 2; i <= n; i++ {
		prev, curr = curr, prev+curr
	}
	return curr
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.stopWorkers()
			return m, tea.Quit
		default:
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - padding*2 - 4
		if m.progress.Width > maxWidth {
			m.progress.Width = maxWidth
		}
		return m, nil

	case tickMsg:
		if m.progress.Percent() == 1.0 {
			return m, nil
		}

		// Note that you can also use progress.Model.SetPercent to set the
		// percentage value explicitly, too.
		cmd := m.progress.IncrPercent(0.25)
		return m, tea.Batch(tickCmd(), cmd)

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	default:
		return m, nil
	}
}

func (m Model) View() string {
	pad := strings.Repeat(" ", padding)
	return "\n" +
		textStyle("Stressing CPUs") +
		pad + m.progress.View() + "\n\n" +
		pad + helpStyle(fmt.Sprintf("Using %d workers. Press q or ctrl+c to quit", m.numWorkers))
}

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
