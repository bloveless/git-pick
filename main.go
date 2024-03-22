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

func focusedCmd() tea.Cmd {
	return func() tea.Msg {
		return focusedMsg{}
	}
}

type model struct {
	pick        tea.Model
	demo        tea.Model
	currentView view
}

type (
	doneMsg    struct{}
	focusedMsg struct{}
	errorMsg   error
)

func (m model) Init() tea.Cmd {
	pCmd := m.pick.Init()
	dCmd := m.demo.Init()
	return tea.Batch(pCmd, dCmd)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		newPick, pCmd := m.pick.Update(msg)
		m.pick = newPick

		newDemo, dCmd := m.demo.Update(msg)
		m.demo = newDemo

		return m, tea.Batch(pCmd, dCmd)

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			// INFO: intercepted and handled by this, the parent model
			return m, tea.Quit

		case "*":
			m.currentView = demoView
			return m, focusedCmd()

		case "#":
			m.currentView = pickView
			return m, focusedCmd()
		}
	}

	switch m.currentView {
	case pickView:
		newPick, cmd := m.pick.Update(msg)
		m.pick = newPick

		return m, cmd

	case demoView:
		newDemo, cmd := m.demo.Update(msg)
		m.demo = newDemo

		return m, cmd

	default:
		// This means that there wasn't a selected view... how strange
		return m, nil
	}
}

func (m model) View() string {
	switch m.currentView {
	case pickView:
		return m.pick.View()

	case demoView:
		return m.demo.View()

	default:
		return "ERR: NO CURRENT VIEW"
	}
}

func main() {
	pv := newPick()
	dv := newDemo()

	m := model{currentView: pickView, pick: pv, demo: dv}

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
