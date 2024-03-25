package graph

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/guptarohit/asciigraph"
)

type Model struct {
	data []float64
}

func New() Model {
	return Model{
		data: []float64{3, 4, 9, 6, 2, 4, 5, 8, 5, 10, 2, 7, 2, 5, 6},
	}
}

func (g Model) Init() tea.Cmd {
	return nil
}

func (g Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return g, nil
}

func (g Model) View() string {
	return asciigraph.Plot(g.data)
}
