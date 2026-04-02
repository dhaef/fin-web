package tui

import (
	tea "charm.land/bubbletea/v2"
)

type Text struct {
	value string
}

func NewText(value string) Text {
	return Text{value: value}
}

func (t Text) Init() tea.Cmd {
	return nil
}

func (t Text) View() tea.View {
	return tea.NewView(t.value)
}

func (t Text) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return t, nil
}
