package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/matthewmyrick/procrastinate-cli/db"
)

const maxLiveJobs = 200

// LiveView shows a live tail of jobs arriving in the queue.
type LiveView struct {
	jobs     []db.Job
	viewport viewport.Model
	width    int
	height   int
}

// NewLiveView creates a new live job stream view.
func NewLiveView() LiveView {
	return LiveView{}
}

// SetJobs sets the job list from a polling result.
func (l *LiveView) SetJobs(jobs []db.Job) {
	l.jobs = jobs
	if len(l.jobs) > maxLiveJobs {
		l.jobs = l.jobs[:maxLiveJobs]
	}
	l.viewport.SetContent(l.renderContent())
}

// AddJob prepends a single job (from LISTEN/NOTIFY).
func (l *LiveView) AddJob(job db.Job) {
	// Deduplicate by ID
	for _, existing := range l.jobs {
		if existing.ID == job.ID {
			return
		}
	}
	l.jobs = append([]db.Job{job}, l.jobs...)
	if len(l.jobs) > maxLiveJobs {
		l.jobs = l.jobs[:maxLiveJobs]
	}
	l.viewport.SetContent(l.renderContent())
}

// SetSize updates the viewport dimensions.
func (l *LiveView) SetSize(width, height int) {
	l.width = width
	l.height = height
	l.viewport.Width = width
	l.viewport.Height = height
	l.viewport.SetContent(l.renderContent())
}

// Update handles messages for the live view.
func (l LiveView) Update(msg tea.Msg) (LiveView, tea.Cmd) {
	var cmd tea.Cmd
	l.viewport, cmd = l.viewport.Update(msg)
	return l, cmd
}

// View renders the live view.
func (l LiveView) View() string {
	return l.viewport.View()
}

func (l *LiveView) renderContent() string {
	if len(l.jobs) == 0 {
		return lipgloss.NewStyle().
			Foreground(ColorMuted).
			Padding(1, 1).
			Render("No recent jobs")
	}

	var b strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorWhite)
	b.WriteString(headerStyle.Render(
		fmt.Sprintf("  %-8s %-24s %-12s %s", "ID", "Task", "Status", "Age")))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(
		"  " + strings.Repeat("─", l.width-4)))
	b.WriteString("\n")

	now := time.Now()
	for _, job := range l.jobs {
		id := fmt.Sprintf("#%d", job.ID)

		task := job.TaskName
		maxTask := 24
		if len(task) > maxTask {
			task = task[:maxTask-1] + "…"
		}

		status := StatusStyle(string(job.Status)).Render(fmt.Sprintf("%-12s", job.Status))
		age := formatAge(now, job)

		b.WriteString(fmt.Sprintf("  %-8s %-24s %s %s\n", id, task, status, age))
	}

	return b.String()
}

// formatAge calculates a human-readable age for a job.
func formatAge(now time.Time, job db.Job) string {
	if job.ScheduledAt == nil {
		return lipgloss.NewStyle().Foreground(ColorMuted).Render("--")
	}

	d := now.Sub(*job.ScheduledAt)
	return lipgloss.NewStyle().Foreground(ColorMuted).Render(formatDuration(d))
}

// formatDuration returns a short human-readable duration.
func formatDuration(d time.Duration) string {
	if d < 0 {
		return "in " + formatDuration(-d)
	}
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
	default:
		days := int(d.Hours()) / 24
		return fmt.Sprintf("%dd %dh", days, int(d.Hours())%24)
	}
}
