package main

import (
	"fmt"
	"time"

	"git-pick/graph"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	spinners = []spinner.Spinner{
		spinner.Line,
		spinner.Dot,
		spinner.MiniDot,
		spinner.Jump,
		spinner.Pulse,
		spinner.Points,
		spinner.Globe,
		spinner.Moon,
		spinner.Monkey,
	}
	modelStyle = lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center).
			BorderStyle(lipgloss.HiddenBorder()).
			Render
	focusedModelStyle = lipgloss.NewStyle().
				Align(lipgloss.Center, lipgloss.Center).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("69")).
				Render
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
)

type tickMsg time.Time

type focusedModel int

const (
	focusedProgress focusedModel = iota
	focusedSpinner
	focusedInput
)

type demo struct {
	name           string
	textInput      textinput.Model
	spinner        spinner.Model
	progress       progress.Model
	graph          graph.Model
	focusedModel   focusedModel
	currentSpinner int
}

func newDemo() demo {
	ti := textinput.New()
	ti.Placeholder = "Your name please"
	d := demo{
		textInput: ti,
		progress:  progress.New(progress.WithDefaultGradient()),
		graph:     graph.New(),
	}

	d.resetSpinner()

	return d
}

func (d *demo) resetSpinner() {
	if d.currentSpinner < 0 {
		d.currentSpinner = len(spinners) - 1
	}

	if d.currentSpinner > len(spinners)-1 {
		d.currentSpinner = 0
	}

	d.spinner = spinner.New(spinner.WithStyle(spinnerStyle), spinner.WithSpinner(spinners[d.currentSpinner]))
}

func (d demo) Init() tea.Cmd { return nil }

func (d demo) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.progress.Width = msg.Width/2 - 2
		d.textInput.Width = msg.Width/2 - 2
		d.spinner.Style.Padding(0, 10)
		return d, nil

	case tea.KeyMsg:
		if d.focusedModel == focusedInput {
			if msg.String() == "enter" {
				d.name = d.textInput.Value()
				return d, nil
			}

			if msg.String() == "esc" {
				d.focusedModel = focusedProgress
				return d, nil
			}

			ti, tCmd := d.textInput.Update(msg)
			d.textInput = ti
			return d, tCmd
		}

		switch msg.String() {
		case "h":
			switch d.focusedModel {
			case focusedInput:
				d.focusedModel = focusedSpinner
				d.textInput.Blur()
			case focusedSpinner:
				d.focusedModel = focusedProgress
				d.textInput.Blur()
			case focusedProgress:
				d.focusedModel = focusedInput
				return d, d.textInput.Focus()
			}

		case "l":
			switch d.focusedModel {
			case focusedProgress:
				d.focusedModel = focusedSpinner
				d.textInput.Blur()
			case focusedSpinner:
				d.focusedModel = focusedInput
				return d, d.textInput.Focus()
			case focusedInput:
				d.focusedModel = focusedProgress
				d.textInput.Blur()
			}

		case "k":
			d.currentSpinner++
			d.resetSpinner()
			return d, d.spinner.Tick

		case "j":
			d.currentSpinner--
			d.resetSpinner()
			return d, d.spinner.Tick
		}

		return d, nil

	case focusedMsg:
		pCmd := d.progress.SetPercent(0.0)
		return d, tea.Batch(pCmd, tickCmd(), d.spinner.Tick)

	case tickMsg:
		// Note that you can also use progress.Model.SetPercent to set the
		// percentage value explicitly, too.
		cmd := d.progress.IncrPercent(0.25)
		return d, tea.Batch(cmd, tickCmd())

	case spinner.TickMsg:
		newSpinner, cmd := d.spinner.Update(msg)
		d.spinner = newSpinner
		return d, cmd

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		progressModel, cmd := d.progress.Update(msg)
		d.progress = progressModel.(progress.Model)
		return d, cmd

	default:
		return d, nil
	}
}

func (d demo) styleWidget(model focusedModel, widget string) string {
	if model == d.focusedModel {
		return focusedModelStyle(widget)
	}

	return modelStyle(widget)
}

func (d demo) View() string {
	row1 := lipgloss.JoinHorizontal(lipgloss.Top, d.styleWidget(focusedProgress, d.progress.View()), d.styleWidget(focusedSpinner, d.spinner.View()))

	hello := ""
	if d.name != "" {
		hello = fmt.Sprintf("\n Hello %s", d.name)
	}

	row2 := lipgloss.JoinHorizontal(lipgloss.Top, d.styleWidget(focusedInput, d.textInput.View()), hello)
	return lipgloss.JoinVertical(lipgloss.Left, row1, row2) + "\n\n" + d.graph.View()
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
