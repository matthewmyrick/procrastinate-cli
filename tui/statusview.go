package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/matthewmyrick/procrastinate-cli/db"
)

// StatusView displays job counts grouped by status.
type StatusView struct {
	counts   []db.StatusCount
	total    int64
	viewport viewport.Model
	width    int
	height   int
}

// NewStatusView creates a new status breakdown view.
func NewStatusView() StatusView {
	return StatusView{}
}

// SetCounts updates the status counts and re-renders content.
func (s *StatusView) SetCounts(counts []db.StatusCount) {
	s.counts = counts
	s.total = 0
	for _, c := range counts {
		s.total += c.Count
	}
	s.viewport.SetContent(s.renderContent())
}

// SetSize updates the viewport dimensions.
func (s *StatusView) SetSize(width, height int) {
	s.width = width
	s.height = height
	s.viewport.Width = width
	s.viewport.Height = height
	s.viewport.SetContent(s.renderContent())
}

// Update handles messages for the status view.
func (s StatusView) Update(msg tea.Msg) (StatusView, tea.Cmd) {
	var cmd tea.Cmd
	s.viewport, cmd = s.viewport.Update(msg)
	return s, cmd
}

// View renders the status view.
func (s StatusView) View() string {
	return s.viewport.View()
}

func (s *StatusView) renderContent() string {
	if len(s.counts) == 0 {
		return lipgloss.NewStyle().
			Foreground(ColorMuted).
			Padding(1, 1).
			Render("No status data available")
	}

	var b strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorWhite)
	b.WriteString(headerStyle.Render(fmt.Sprintf("  %-14s %8s   %s", "Status", "Count", "Bar")))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(
		"  " + strings.Repeat("─", s.width-4)))
	b.WriteString("\n")

	// Build a complete count map with all statuses
	countMap := make(map[db.JobStatus]int64)
	for _, c := range s.counts {
		countMap[c.Status] = c.Count
	}

	maxBarWidth := s.width - 30
	if maxBarWidth < 10 {
		maxBarWidth = 10
	}

	for _, status := range db.AllStatuses() {
		count := countMap[status]
		style := StatusStyle(string(status))

		// Bar proportional to total
		barWidth := 0
		if s.total > 0 && count > 0 {
			barWidth = int(float64(count) / float64(s.total) * float64(maxBarWidth))
			if barWidth < 1 {
				barWidth = 1
			}
		}
		bar := style.Render(strings.Repeat("█", barWidth))

		statusLabel := style.Render(fmt.Sprintf("%-14s", status))
		countStr := lipgloss.NewStyle().Foreground(ColorWhite).Render(fmt.Sprintf("%8d", count))

		b.WriteString(fmt.Sprintf("  %s %s   %s\n", statusLabel, countStr, bar))
	}

	// Separator and total
	b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(
		"  " + strings.Repeat("─", s.width-4)))
	b.WriteString("\n")

	totalStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorWhite)
	b.WriteString(totalStyle.Render(fmt.Sprintf("  %-14s %8d", "Total", s.total)))
	b.WriteString("\n")

	return b.String()
}
