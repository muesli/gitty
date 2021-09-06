package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Label struct {
	Name  string
	Color string
}
type Labels []Label

func (l Label) View() string {
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#" + l.Color))

	return labelStyle.Render(fmt.Sprintf("◖%s◗", l.Name))
}

func (ll Labels) View() string {
	var s strings.Builder

	for _, v := range ll {
		s.WriteString(v.View() + " ")
	}

	return strings.TrimSpace(s.String())
}
