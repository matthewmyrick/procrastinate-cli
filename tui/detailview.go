package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/matthewmyrick/procrastinate-cli/db"
)

// DetailView shows full details for a selected job as an overlay.
type DetailView struct {
	job      *db.Job
	events   []db.JobEvent
	viewport viewport.Model
	visible  bool
	width    int
	height   int
}

// NewDetailView creates a new detail view.
func NewDetailView() DetailView {
	return DetailView{}
}

// SetJob populates the detail view with job data and events.
func (d *DetailView) SetJob(job *db.Job, events []db.JobEvent) {
	d.job = job
	d.events = events
	d.visible = true
	d.viewport.SetContent(d.renderContent())
	d.viewport.GotoTop()
}

// SetVisible controls whether the overlay is shown.
func (d *DetailView) SetVisible(v bool) {
	d.visible = v
}

// IsVisible returns whether the overlay is showing.
func (d *DetailView) IsVisible() bool {
	return d.visible
}

// SetSize updates the overlay dimensions.
func (d *DetailView) SetSize(width, height int) {
	d.width = width
	d.height = height
	vw, vh := d.overlayDimensions()
	frameW := OverlayStyle.GetHorizontalFrameSize()
	frameH := OverlayStyle.GetVerticalFrameSize()
	d.viewport.Width = vw - frameW
	d.viewport.Height = vh - frameH - 1 // -1 for footer
	if d.job != nil {
		d.viewport.SetContent(d.renderContent())
	}
}

func (d *DetailView) overlayDimensions() (int, int) {
	vw := d.width * 80 / 100
	vh := d.height * 80 / 100
	if vw < 60 {
		vw = 60
	}
	if vw > d.width-4 {
		vw = d.width - 4
	}
	if vh < 15 {
		vh = 15
	}
	if vh > d.height-4 {
		vh = d.height - 4
	}
	return vw, vh
}

// Update handles messages for the detail view.
func (d DetailView) Update(msg tea.Msg) (DetailView, tea.Cmd) {
	if !d.visible {
		return d, nil
	}
	var cmd tea.Cmd
	d.viewport, cmd = d.viewport.Update(msg)
	return d, cmd
}

// View renders the detail overlay.
func (d DetailView) View() string {
	if !d.visible || d.job == nil {
		return ""
	}

	vw, vh := d.overlayDimensions()

	content := d.viewport.View()

	footer := lipgloss.NewStyle().
		Foreground(ColorMuted).
		Render(fmt.Sprintf("  ↑↓ scroll · esc close  (%.0f%%)", d.viewport.ScrollPercent()*100))

	inner := lipgloss.JoinVertical(lipgloss.Left, content, footer)

	return OverlayStyle.
		Width(vw).
		Height(vh).
		Render(inner)
}

func (d *DetailView) renderContent() string {
	if d.job == nil {
		return ""
	}

	var b strings.Builder
	j := d.job

	// Title
	title := TitleStyle.Render(fmt.Sprintf("Job #%d", j.ID))
	b.WriteString(title)
	b.WriteString("\n\n")

	// Fields
	d.writeField(&b, "Task", j.TaskName)
	d.writeField(&b, "Queue", j.QueueName)
	d.writeField(&b, "Status", StatusStyle(string(j.Status)).Render(string(j.Status)))
	d.writeField(&b, "Priority", fmt.Sprintf("%d", j.Priority))
	d.writeField(&b, "Attempts", fmt.Sprintf("%d", j.Attempts))

	if j.Lock != nil {
		d.writeField(&b, "Lock", *j.Lock)
	}
	if j.QueueingLock != nil {
		d.writeField(&b, "Queueing Lock", *j.QueueingLock)
	}
	if j.ScheduledAt != nil {
		d.writeField(&b, "Scheduled At", j.ScheduledAt.Local().Format("2006-01-02 15:04:05"))
	}
	d.writeField(&b, "Abort Requested", fmt.Sprintf("%v", j.AbortRequested))
	if j.WorkerID != nil {
		d.writeField(&b, "Worker ID", fmt.Sprintf("%d", *j.WorkerID))
	} else {
		d.writeField(&b, "Worker ID", "-")
	}

	// Args (pretty-printed JSON)
	b.WriteString("\n")
	argsLabel := LabelStyle.Render("Args:")
	b.WriteString(argsLabel)
	b.WriteString("\n")

	var prettyArgs json.RawMessage
	if err := json.Unmarshal(j.Args, &prettyArgs); err == nil {
		pretty, err := json.MarshalIndent(prettyArgs, "  ", "  ")
		if err == nil {
			b.WriteString(lipgloss.NewStyle().Foreground(ColorSecondary).Render("  " + string(pretty)))
		} else {
			b.WriteString("  " + string(j.Args))
		}
	} else {
		b.WriteString("  " + string(j.Args))
	}
	b.WriteString("\n")

	// Events
	if len(d.events) > 0 {
		b.WriteString("\n")
		eventsLabel := LabelStyle.Render("Events:")
		b.WriteString(eventsLabel)
		b.WriteString("\n")

		for _, e := range d.events {
			ts := lipgloss.NewStyle().Foreground(ColorMuted).
				Render(e.At.Local().Format("2006-01-02 15:04:05"))
			eventType := lipgloss.NewStyle().Foreground(ColorWhite).Bold(true).
				Render(fmt.Sprintf("%-22s", e.Type))
			b.WriteString(fmt.Sprintf("  %s  %s\n", eventType, ts))
		}
	}

	return b.String()
}

func (d *DetailView) writeField(b *strings.Builder, label, value string) {
	l := LabelStyle.Render(label + ":")
	v := ValueStyle.Render(value)
	b.WriteString(fmt.Sprintf("%s %s\n", l, v))
}
