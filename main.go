package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

const listHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
)

type item struct {
	name  string
	short string
}

func (i item) FilterValue() string { return i.short }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.short)

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
	choice string
	err    error
}

type newItemsMsg []list.Item

func (m model) Init() tea.Cmd {
	return func() tea.Msg {
		repo, err := git.PlainOpenWithOptions("", &git.PlainOpenOptions{
			DetectDotGit: true,
		})
		if errors.Is(err, git.ErrRepositoryNotExists) {
			return tea.Quit
		}
		if err != nil {
			panic(err)
		}

		b, err := repo.Branches()
		if err != nil {
			panic(err)
		}
		defer b.Close()

		var items []list.Item
		for {
			ref, err := b.Next()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				panic(err)
			}

			items = append(items, item{
				name:  ref.Name().String(),
				short: ref.Name().Short(),
			})
		}

		return newItemsMsg(items)
	}
}

type doneMsg struct{}
type errorMsg error

func (m model) updateBranch(i item) tea.Cmd {
	return func() tea.Msg {
		r, err := git.PlainOpenWithOptions("", &git.PlainOpenOptions{
			DetectDotGit: true,
		})
		if err != nil {
			return errorMsg(err)
		}

		wt, err := r.Worktree()
		if err != nil {
			return errorMsg(err)
		}

		err = wt.Checkout(&git.CheckoutOptions{
			Branch: plumbing.ReferenceName(i.name),
			Force:  true,
		})
		if err != nil {
			return errorMsg(err)
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
				m.choice = string(i.name)
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
		s = fmt.Sprintf("Error: %s\n\n%s", m.err, s)
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
