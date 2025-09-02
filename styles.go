package main

import "github.com/charmbracelet/lipgloss"

var (
	StatusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(ColorPrimary)).
			Foreground(lipgloss.Color(ColorBright)).
			Padding(0, 1)

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorDim))

	ModifiedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(ColorPrimary)).
			Foreground(lipgloss.Color(ColorAlert)).
			Bold(true)

	EditorStyle = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color(ColorBorder)).
			Padding(0, 1)

	LineNumberStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorMuted)).
			Width(4).
			Align(lipgloss.Right)

	SelectedTextStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(ColorHighlight)).
			Foreground(lipgloss.Color(ColorForeground))

	CurrentWordStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(ColorSelection)).
			Foreground(lipgloss.Color(ColorForeground))

	CursorLineStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(ColorBackground))

	HelpBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorBorder)).
			AlignHorizontal(lipgloss.Center).
			Padding(1, 2).
			Margin(1, 0)

	HelpTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorPrimary)).
			Bold(true).
			Underline(true)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorAccent)).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorText))

	// Minibuffer styles
	MinibufferStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(ColorSelection)).
			Foreground(lipgloss.Color(ColorForeground)).
			Padding(0, 1)

	MinibufferPromptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorAccent)).
			Bold(true)

	MinibufferInputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorForeground))

	MinibufferCursorStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(ColorForeground)).
			Foreground(lipgloss.Color(ColorBackground))

	SearchResultSelectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(ColorMinibuffer)).
			Foreground(lipgloss.Color(ColorMinibufferBg)).
			Bold(true)

	SearchResultNormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorInfo))

	// Message styles
	SuccessMessageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSuccess))

	ErrorMessageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorError))

	WarningMessageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorWarning))

	// Status bar section styles
	StatusBarLeftStyle = func(width int) lipgloss.Style {
		return lipgloss.NewStyle().
			Width(width / StatusBarSections).
			Align(lipgloss.Left).
			Background(lipgloss.Color(ColorPrimary))
	}

	StatusBarCenterStyle = func(width int) lipgloss.Style {
		return lipgloss.NewStyle().
			Width(width / StatusBarSections).
			Align(lipgloss.Center).
			Background(lipgloss.Color(ColorPrimary))
	}

	StatusBarRightStyle = func(width int) lipgloss.Style {
		return lipgloss.NewStyle().
			Width(width / StatusBarSections).
			Align(lipgloss.Right).
			Background(lipgloss.Color(ColorPrimary))
	}
)