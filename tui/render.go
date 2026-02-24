package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (a *App) renderTopBar() string {
	queueLabel := TopBarQueueStyle.Render(fmt.Sprintf("Queue: %s", a.currentQueue))
	connLabel := TopBarConnStyle.Render(fmt.Sprintf("Connection: %s", a.currentConn))

	gap := a.width - lipgloss.Width(queueLabel) - lipgloss.Width(connLabel) - 2
	if gap < 1 {
		gap = 1
	}

	return TopBarStyle.Width(a.width).Render(
		queueLabel + strings.Repeat(" ", gap) + connLabel,
	)
}

func (a *App) renderDetail(width, height int) string {
	style := DetailStyle
	if a.focus == focusDetail {
		style = DetailFocusedStyle
	}

	// Not connected — show simple status message
	if !a.connected {
		statusMsg := lipgloss.NewStyle().Foreground(ColorMuted).Render(
			"Not connected — press C to switch connections")
		centered := lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, statusMsg)
		return style.Width(width).Height(height).Render(centered)
	}

	// Job detail mode — show detail in the right pane
	if a.showDetail {
		content := a.detailView.ViewInline(width, height)
		return style.Width(width).Height(height).Render(
			padOrTruncate(content, width, height),
		)
	}

	// Dashboard mode — show tab bar + content
	tabBar := a.tabBar.View()

	var tabContent string
	switch a.tabBar.Active() {
	case TabStatus:
		tabContent = a.statusView.View()
	case TabLive:
		tabContent = a.liveView.View()
	case TabOrphaned:
		tabContent = a.orphanedView.View()
	}

	var parts []string
	if a.lastError != nil {
		parts = append(parts, ErrorBannerStyle.Render(fmt.Sprintf("Error: %v", a.lastError)))
	}
	parts = append(parts, tabBar, tabContent)

	inner := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return style.Width(width).Height(height).Render(
		padOrTruncate(inner, width, height),
	)
}

func (a *App) renderToastOverlay(base string) string {
	rendered := ToastStyle.Render(a.toast)
	return lipgloss.Place(a.width, a.height, lipgloss.Right, lipgloss.Top, rendered,
		lipgloss.WithWhitespaceChars(" "),
	)
}

func (a *App) renderPicker(title string, items []string, selected int) string {
	var b strings.Builder

	titleRendered := TitleStyle.Render(title)
	b.WriteString(titleRendered)
	b.WriteString("\n\n")

	for i, item := range items {
		cursor := "  "
		style := lipgloss.NewStyle().Foreground(ColorMuted)
		if i == selected {
			cursor = "► "
			style = lipgloss.NewStyle().Foreground(ColorWhite).Bold(true)
		}
		b.WriteString(style.Render(cursor+item) + "\n")
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("↑↓ navigate · enter select · esc cancel"))

	width := 40
	if width > a.width-10 {
		width = a.width - 10
	}

	return OverlayStyle.Width(width).Render(b.String())
}

func (a *App) renderHelpBar() string {
	var parts []string
	for _, k := range a.keys.HelpKeys() {
		parts = append(parts, fmt.Sprintf("%s %s", k.Help().Key, k.Help().Desc))
	}
	help := strings.Join(parts, " · ")
	return HelpStyle.Width(a.width).MaxHeight(1).Render(help)
}

func (a *App) renderHelpOverlay() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Keyboard Shortcuts"))
	b.WriteString("\n\n")

	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorWhite).Width(14)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#DDDDDD"))

	for _, k := range a.keys.AllKeys() {
		h := k.Help()
		b.WriteString(fmt.Sprintf("  %s %s\n", keyStyle.Render(h.Key), descStyle.Render(h.Desc)))
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("  Press any key to close"))

	width := 40
	if width > a.width-10 {
		width = a.width - 10
	}

	return OverlayStyle.Width(width).Render(b.String())
}
