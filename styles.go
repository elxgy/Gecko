package main

import "github.com/charmbracelet/lipgloss"

// Centralized Lipgloss styles. Keeping them in one place simplifies future theming.
var (
	// Global UI elements
	statusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#6f7cbf")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	modifiedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	editorStyle = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("#6f7cbf")).
			Padding(0, 1)

	lineNumberStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Width(4).
			Align(lipgloss.Right)

	selectedTextStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#a600a0")).
				Foreground(lipgloss.Color("#f8f8f2"))

	cursorLineStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#282a36"))

	wordHighlightStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#3e4451")) // Subtle background for word highlight

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f8f8f2")). // Consistent visible cursor
			Background(lipgloss.Color("#282a36"))

	helpBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#6f7cbf")).
			AlignHorizontal(lipgloss.Center).
			Padding(1, 2).
			Margin(1, 0)

	helpTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6f7cbf")).
			Bold(true).
			Underline(true)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89b4fa")).
			Bold(true)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cdd6f4"))

	// Minibuffer & search styles
	minibufferStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#44475a")).
			Foreground(lipgloss.Color("#f8f8f2")).
			Padding(0, 1)

	minibufferPromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#89b4fa")).
				Bold(true)

	minibufferInputStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#f8f8f2"))

	minibufferCursorStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#f8f8f2")).
				Foreground(lipgloss.Color("#282a36"))

	searchResultSelectedStyle = lipgloss.NewStyle().
					Background(lipgloss.Color("#b889fa")).
					Foreground(lipgloss.Color("#1e1e2e")).
					Bold(true)

	searchResultNormalStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00ffe1"))

	// Flash message styles
	flashSuccessStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ff00b3"))

	flashErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#800024"))

	flashWarningStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#e3e094"))
)
