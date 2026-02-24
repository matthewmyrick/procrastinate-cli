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

	// Connected — show tab bar + content
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
	w := lipgloss.Width(rendered)

	x := a.width - w - 2
	if x < 0 {
		x = 0
	}
	y := 1 // just below the top bar

	return placeOverlay(x, y, rendered, base)
}

func (a *App) renderOverlay(base, overlay string) string {
	if overlay == "" {
		return base
	}

	overlayWidth := lipgloss.Width(overlay)
	overlayHeight := lipgloss.Height(overlay)

	x := (a.width - overlayWidth) / 2
	y := (a.height - overlayHeight) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	return placeOverlay(x, y, overlay, base)
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

// placeOverlay places an overlay string on top of a background string at x, y.
func placeOverlay(x, y int, overlay, background string) string {
	bgLines := strings.Split(background, "\n")
	olLines := strings.Split(overlay, "\n")

	for i, olLine := range olLines {
		bgY := y + i
		if bgY >= len(bgLines) {
			break
		}

		bgLine := bgLines[bgY]
		bgRunes := []rune(bgLine)
		olRunes := []rune(olLine)

		var newLine []rune
		if x > 0 {
			if x <= len(bgRunes) {
				newLine = append(newLine, bgRunes[:x]...)
			} else {
				newLine = append(newLine, bgRunes...)
				for len(newLine) < x {
					newLine = append(newLine, ' ')
				}
			}
		}

		newLine = append(newLine, olRunes...)

		endX := x + len(olRunes)
		if endX < len(bgRunes) {
			newLine = append(newLine, bgRunes[endX:]...)
		}

		bgLines[bgY] = string(newLine)
	}

	return strings.Join(bgLines, "\n")
}
