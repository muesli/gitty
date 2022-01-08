package vcs

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Label represents a label.
type Label struct {
	Name  string
	Color string
}

// Labels represents a list of labels.
type Labels []Label

// View returns a string representation of the label.
func (l Label) View() string {
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(l.Color))

	return labelStyle.Render(fmt.Sprintf("◖%s◗", l.Name))
}

// View returns a string representation of the labels.
func (ll Labels) View() string {
	var s strings.Builder

	for _, v := range ll {
		s.WriteString(v.View() + " ")
	}

	return strings.TrimSpace(s.String())
}
