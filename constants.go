package main

// Color constants
const (
	// Primary theme colors
	ColorPrimary     = "#6f7cbf"
	ColorBackground  = "#282a36"
	ColorForeground  = "#f8f8f2"
	ColorSelection   = "#44475a"
	ColorHighlight   = "#a600a0"
	ColorAccent      = "#89b4fa"
	ColorText        = "#cdd6f4"
	ColorBorder      = "#6f7cbf"
	
	// Status message colors
	ColorSuccess     = "#ff00b3"
	ColorError       = "#800024"
	ColorWarning     = "#e3e094"
	ColorInfo        = "#00ffe1"
	
	// UI element colors
	ColorMinibuffer  = "#b889fa"
	ColorMinibufferBg = "#1e1e2e"
	ColorWhitespace  = "236"
	ColorMuted       = "240"
	ColorDim         = "241"
	ColorBright      = "230"
	ColorAlert       = "196"
)

// UI layout constants
const (
	KeyColumnWidth        = 18
	HelpMaxWidthPercent   = 40
	HelpMinWidth          = 50
	StatusBarSections     = 3
	SearchContextLength   = 30
	LinePreviewMaxLength  = 70
	LinePreviewTruncate   = 67
	
	// Minimum terminal size constraints
	MinTerminalWidth      = 40
	MinTerminalHeight     = 10
	MinLineNumberWidth    = 6  // "  999 " format
	MinContentWidth       = 20 // Minimum usable content area
)

// Editor behavior constants
const (
	MaxHistorySize        = 100
	TickIntervalMs        = 500
	FilePermissions       = 0644
)

// ANSI escape sequences
const (
	ClearScreen = "\033[2J\033[H"
)