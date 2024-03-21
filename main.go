package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

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

type item string

func (i item) FilterValue() string { return string(i) }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type model struct {
	list   list.Model
	err    error
	choice string
}

type newItemsMsg []list.Item

func (m model) Init() tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("git", "for-each-ref", "--format", "%(refname:short)").CombinedOutput()
		if err != nil {
			return errorMsg(err)
		}

		branches := strings.Split(strings.TrimSpace(string(out)), "\n")

		var items []list.Item
		for _, b := range branches {
			items = append(items, item(b))
		}

		return newItemsMsg(items)
	}
}

type (
	doneMsg  struct{}
	errorMsg error
)

func (m model) updateBranch(i item) tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("git", "checkout", string(i)).CombinedOutput()
		if err != nil {
			return errorMsg(fmt.Errorf("unabled to checkout branch: %v; %s", string(out), err))
		}

		return doneMsg{}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
				return m, m.updateBranch(i)
			}

			// Selected item wasn't an item? This shouldn't happen
			return m, nil
		}

	case newItemsMsg:
		m.list.SetItems(msg)
		return m, nil

	case doneMsg:
		return m, tea.Quit

	case errorMsg:
		m.err = msg
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	s := "\n" + m.list.View()

	if m.err != nil {
		s = fmt.Sprintf("%s\nError: %s", s, m.err)
	}

	return s
}

func main() {
	var items []list.Item
	const defaultWidth = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Which branch would you like to switch to?"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := model{list: l}

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
