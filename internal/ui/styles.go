package ui

import "github.com/charmbracelet/lipgloss"

// Styles holds every lipgloss style used by the menu. Colors are adaptive
// so the menu stays readable on both light and dark terminal themes.
type Styles struct {
	Title        lipgloss.Style
	Meta         lipgloss.Style
	Divider      lipgloss.Style
	ColumnHeader lipgloss.Style
	Muted        lipgloss.Style
	Good         lipgloss.Style
	Accent       lipgloss.Style
	Selected     lipgloss.Style
	Separator    lipgloss.Style
}

// NewStyles builds the default style set.
func NewStyles() Styles {
	accent := lipgloss.AdaptiveColor{Light: "25", Dark: "39"}
	good := lipgloss.AdaptiveColor{Light: "28", Dark: "42"}
	muted := lipgloss.AdaptiveColor{Light: "243", Dark: "245"}

	return Styles{
		Title:        lipgloss.NewStyle().Bold(true),
		Meta:         lipgloss.NewStyle().Foreground(muted),
		Divider:      lipgloss.NewStyle().Foreground(muted),
		ColumnHeader: lipgloss.NewStyle().Bold(true).Foreground(muted),
		Muted:        lipgloss.NewStyle().Foreground(muted),
		Good:         lipgloss.NewStyle().Foreground(good),
		Accent:       lipgloss.NewStyle().Foreground(accent).Bold(true),
		Selected:     lipgloss.NewStyle().Reverse(true),
		Separator:    lipgloss.NewStyle().Foreground(muted),
	}
}
