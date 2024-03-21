package main

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type pick struct {
	list     list.Model
	err      error
	selected string
}

func newPick() pick {
	var items []list.Item
	const defaultWidth = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Which branch would you like to switch to?"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return pick{list: l}
}

func (p pick) Init() tea.Cmd {
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

func (p pick) updateBranch(i item) tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("git", "checkout", string(i)).CombinedOutput()
		if err != nil {
			return errorMsg(fmt.Errorf("unabled to checkout branch: %v; %s", string(out), err))
		}

		return doneMsg{}
	}
}

func (p pick) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.list.SetWidth(msg.Width)
		return p, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "enter":
			i, ok := p.list.SelectedItem().(item)
			if ok {
				p.selected = string(i)
				return p, p.updateBranch(i)
			}

			// Selected item wasn't an item? This shouldn't happen
			return p, nil
		}

	case newItemsMsg:
		p.list.SetItems(msg)
		return p, nil

	case doneMsg:
		return p, tea.Quit

	case errorMsg:
		p.err = msg
		return p, nil
	}

	var cmd tea.Cmd
	p.list, cmd = p.list.Update(msg)
	return p, cmd
}

func (p pick) View() string {
	s := p.list.View()

	if p.err != nil {
		s += fmt.Sprintf("\nError: %s", p.err)
	}

	return s
}

type newItemsMsg []list.Item

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
