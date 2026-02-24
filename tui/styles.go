package tui

import "github.com/charmbracelet/lipgloss"

// Colors used throughout the TUI.
var (
	ColorPrimary   = lipgloss.Color("#7D56F4")
	ColorSecondary = lipgloss.Color("#04B575")
	ColorError     = lipgloss.Color("#FF4444")
	ColorWarning   = lipgloss.Color("#FFAA00")
	ColorMuted     = lipgloss.Color("#626262")
	ColorWhite     = lipgloss.Color("#FFFFFF")
	ColorDim       = lipgloss.Color("#4A4A4A")

	// Status-specific colors
	ColorTodo      = lipgloss.Color("#04B575")
	ColorDoing     = lipgloss.Color("#FFAA00")
	ColorSucceeded = lipgloss.Color("#4A9EFF")
	ColorFailed    = lipgloss.Color("#FF4444")
	ColorCancelled = lipgloss.Color("#888888")
	ColorAborted   = lipgloss.Color("#CC6699")
)

// Layout styles
var (
	TopBarStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1)

	TopBarQueueStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorSecondary)

	TopBarConnStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	SidebarStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorMuted).
			Padding(0, 0)

	SidebarFocusedStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary).
				Padding(0, 0)

	DetailStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorMuted).
			Padding(0, 0)

	DetailFocusedStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary).
				Padding(0, 0)
)

// Tab styles
var (
	ActiveTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite).
			Background(ColorPrimary).
			Padding(0, 2)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Padding(0, 2)

	TabBarStyle = lipgloss.NewStyle().
			Padding(0, 0).
			MarginBottom(0)
)

// Content styles
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	LabelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite).
			Width(16)

	ValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DDDDDD"))

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Padding(0, 1)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	ErrorBannerStyle = lipgloss.NewStyle().
				Foreground(ColorError).
				Padding(0, 1)

	ToastStyle = lipgloss.NewStyle().
			Foreground(ColorWhite).
			Background(lipgloss.Color("#CC3333")).
			Padding(0, 2).
			Bold(true)

	OverlayStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.DoubleBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2)
)

// StatusStyle returns the appropriate style for a job status.
func StatusStyle(status string) lipgloss.Style {
	base := lipgloss.NewStyle().Bold(true)
	switch status {
	case "todo":
		return base.Foreground(ColorTodo)
	case "doing":
		return base.Foreground(ColorDoing)
	case "succeeded":
		return base.Foreground(ColorSucceeded)
	case "failed":
		return base.Foreground(ColorFailed)
	case "cancelled":
		return base.Foreground(ColorCancelled)
	case "aborting", "aborted":
		return base.Foreground(ColorAborted)
	default:
		return base.Foreground(ColorMuted)
	}
}
