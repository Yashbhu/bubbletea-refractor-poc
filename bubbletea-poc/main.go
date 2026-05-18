package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	stateLoading state = iota
	stateReady
	stateError
)

type dataLoadedMsg struct {
	rows []table.Row
	err  error
}

type Model struct {
	spinner  spinner.Model
	table    table.Model
	state    state
	errorMsg string
	columns  []table.Column
	simulate bool
}

func newModel(simulateError bool) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	columns := []table.Column{
		{Title: "NAME", Width: 20},
		{Title: "TAG", Width: 12},
		{Title: "DIGEST", Width: 20},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(5),
	)

	s2 := table.DefaultStyles()
	s2.Header = s2.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		Bold(false)
	t.SetStyles(s2)

	return Model{
		spinner:  s,
		table:    t,
		state:    stateLoading,
		columns:  columns,
		simulate: simulateError,
	}
}

func fetchCmd(simulateError bool) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(2 * time.Second)

		if simulateError {
			return dataLoadedMsg{err: fmt.Errorf("harbor: connection refused")}
		}

		return dataLoadedMsg{
			rows: []table.Row{
				{"nginx", "latest", "sha256:abc123"},
				{"alpine", "3.18", "sha256:def456"},
				{"redis", "7.0", "sha256:ghi789"},
			},
		}
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, fetchCmd(m.simulate))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case dataLoadedMsg:
		if msg.err != nil {
			m.state = stateError
			m.errorMsg = msg.err.Error()
			return m, tea.Quit
		}
		m.table.SetRows(msg.rows)
		m.state = stateReady
		return m, tea.Quit

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m Model) View() string {
	switch m.state {
	case stateLoading:
		return fmt.Sprintf("\n  %s Fetching artifacts...\n\n", m.spinner.View())
	case stateError:
		return fmt.Sprintf("\n  Error: %s\n\n", m.errorMsg)
	case stateReady:
		return fmt.Sprintf("\n%s\n", m.table.View())
	}
	return ""
}

func main() {
	simulateError := flag.Bool("error", false, "simulate API error")
	flag.Parse()

	m := newModel(*simulateError)
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}
