package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const listHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
)

type view int

const (
	pickView view = iota
	demoView
)

type model struct {
	pick        tea.Model
	currentView view
}

type (
	doneMsg  struct{}
	errorMsg error
)

func (m model) Init() tea.Cmd {
	pCmd := m.pick.Init()
	return tea.Batch(pCmd)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		var cmd tea.Cmd
		var cmds []tea.Cmd
		m.pick, cmd = m.pick.Update(msg)
		cmds = append(cmds, cmd)
		// TODO: All views should receive these messages even if they aren't active
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			// INFO: intercepted and handled by this, the parent model
			return m, tea.Quit
		case "*":
			m.currentView = demoView
			return m, nil
		}

	}

	switch m.currentView {
	case pickView:
		var cmd tea.Cmd
		m.pick, cmd = m.pick.Update(msg)
		return m, cmd
	}

	// This means that there wasn't a selected view... how strange
	return m, nil
}

func (m model) View() string {
	if m.currentView == pickView {
		return m.pick.View()
	}

	return "ERR: NO CURRENT VIEW"
}

func main() {
	pv := newPick()

	m := model{currentView: pickView, pick: pv}

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
