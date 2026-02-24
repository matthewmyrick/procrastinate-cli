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

// OrphanedView shows jobs that appear stuck or abandoned.
type OrphanedView struct {
	jobs     []db.Job
	viewport viewport.Model
	width    int
	height   int
}

// NewOrphanedView creates a new orphaned jobs view.
func NewOrphanedView() OrphanedView {
	return OrphanedView{}
}

// SetJobs updates the orphaned job data.
func (o *OrphanedView) SetJobs(jobs []db.Job) {
	o.jobs = jobs
	o.viewport.SetContent(o.renderContent())
}

// SetSize updates the viewport dimensions.
func (o *OrphanedView) SetSize(width, height int) {
	o.width = width
	o.height = height
	o.viewport.Width = width
	o.viewport.Height = height
	o.viewport.SetContent(o.renderContent())
}

// Update handles messages for the orphaned view.
func (o OrphanedView) Update(msg tea.Msg) (OrphanedView, tea.Cmd) {
	var cmd tea.Cmd
	o.viewport, cmd = o.viewport.Update(msg)
	return o, cmd
}

// View renders the orphaned view.
func (o OrphanedView) View() string {
	return o.viewport.View()
}

func (o *OrphanedView) renderContent() string {
	if len(o.jobs) == 0 {
		return lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Padding(1, 1).
			Render("No orphaned jobs detected")
	}

	var b strings.Builder

	// Warning header
	warningStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorWarning)
	b.WriteString(warningStyle.Render(
		fmt.Sprintf("  ⚠ %d orphaned job(s) found", len(o.jobs))))
	b.WriteString("\n\n")

	// Column headers
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorWhite)
	b.WriteString(headerStyle.Render(
		fmt.Sprintf("  %-8s %-20s %-10s %-10s %s", "ID", "Task", "Status", "Stuck For", "Worker")))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(
		"  " + strings.Repeat("─", o.width-4)))
	b.WriteString("\n")

	now := time.Now()
	for _, job := range o.jobs {
		id := fmt.Sprintf("#%d", job.ID)

		task := job.TaskName
		maxTask := 20
		if len(task) > maxTask {
			task = task[:maxTask-1] + "…"
		}

		status := StatusStyle(string(job.Status)).Render(fmt.Sprintf("%-10s", job.Status))

		stuckFor := lipgloss.NewStyle().Foreground(ColorWarning).Render("--")
		if job.ScheduledAt != nil {
			d := now.Sub(*job.ScheduledAt)
			stuckFor = lipgloss.NewStyle().Foreground(ColorWarning).
				Render(fmt.Sprintf("%-10s", formatDuration(d)))
		}

		worker := lipgloss.NewStyle().Foreground(ColorMuted).Render("-")
		if job.WorkerID != nil {
			worker = lipgloss.NewStyle().Foreground(ColorError).
				Render(fmt.Sprintf("#%d (dead)", *job.WorkerID))
		}

		b.WriteString(fmt.Sprintf("  %-8s %-20s %s %s %s\n", id, task, status, stuckFor, worker))
	}

	return b.String()
}
