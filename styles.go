package main

import "github.com/charmbracelet/lipgloss"

// Centralized Lipgloss styles. Keeping them in one place simplifies future theming.
var (
	// Global UI elements
	backgroundColor = lipgloss.Color("#6f7cbf")

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
			// No background to prevent color bleeding

	selectedTextStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#4a5568")). // Darker, more subtle highlight with better contrast
				Foreground(lipgloss.Color("#e2e8f0"))  // Light foreground for better visibility

	cursorLineStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#2a2d3a")) // Subtle dark background for current line

	wordHighlightStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#264f78")). // VSCode-like blue background
				Foreground(lipgloss.Color("#ffffff"))  // White text for better contrast

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#282a36")). // White cursor matching VS Code default
			Background(lipgloss.Color("#ffffff"))

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
