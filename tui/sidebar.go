package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/matthewmyrick/procrastinate-cli/db"
)

var (
	filterOptions = []string{"", "doing", "todo", "succeeded", "failed", "cancelled"}
	filterLabels  = []string{"All", "Doing", "Todo", "Succeeded", "Failed", "Cancelled"}
)

// jobItem wraps a db.Job to implement list.Item.
type jobItem struct {
	job db.Job
}

func (j jobItem) FilterValue() string { return j.job.TaskName }

// jobItemDelegate renders each job in the sidebar list.
type jobItemDelegate struct{}

func (d jobItemDelegate) Height() int                             { return 1 }
func (d jobItemDelegate) Spacing() int                            { return 0 }
func (d jobItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d jobItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	ji, ok := item.(jobItem)
	if !ok {
		return
	}

	id := fmt.Sprintf("#%d", ji.job.ID)
	task := ji.job.TaskName
	status := string(ji.job.Status)

	// Truncate task name to fit
	maxTask := m.Width() - len(id) - len(status) - 6
	if maxTask < 4 {
		maxTask = 4
	}
	if len(task) > maxTask {
		task = task[:maxTask-1] + "…"
	}

	statusRendered := StatusStyle(status).Render(status)

	line := fmt.Sprintf(" %s %s %s", id, task, statusRendered)

	if index == m.Index() {
		line = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite).
			Background(ColorDim).
			Width(m.Width()).
			Render(fmt.Sprintf(" ► %s %s %s", id, task, status))
	}

	fmt.Fprint(w, line)
}

// Sidebar is the left-pane job list.
type Sidebar struct {
	list        list.Model
	jobs        []db.Job
	focused     bool
	width       int
	height      int
	filterIndex int // index into filterOptions
}

// NewSidebar creates a new sidebar with the given dimensions.
func NewSidebar(width, height int) Sidebar {
	delegate := jobItemDelegate{}
	l := list.New([]list.Item{}, delegate, width-2, height-3)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.DisableQuitKeybindings()

	return Sidebar{
		list:   l,
		width:  width,
		height: height,
	}
}

// SetJobs updates the sidebar with new job data.
// Returns a tea.Cmd that must be executed (re-filters items if a filter is active).
func (s *Sidebar) SetJobs(jobs []db.Job) tea.Cmd {
	s.jobs = jobs
	items := make([]list.Item, len(jobs))
	for i, j := range jobs {
		items[i] = jobItem{job: j}
	}
	return s.list.SetItems(items)
}

// SelectedJob returns the currently highlighted job, or nil if none.
func (s *Sidebar) SelectedJob() *db.Job {
	item := s.list.SelectedItem()
	if item == nil {
		return nil
	}
	ji := item.(jobItem)
	return &ji.job
}

// CurrentFilter returns the current status filter value (empty string = All).
func (s *Sidebar) CurrentFilter() string {
	return filterOptions[s.filterIndex]
}

// SetSize updates the sidebar dimensions.
func (s *Sidebar) SetSize(width, height int) {
	s.width = width
	s.height = height
	s.list.SetSize(width-2, height-4) // header + filter label
}

// SetFocused sets whether the sidebar has focus.
func (s *Sidebar) SetFocused(f bool) {
	s.focused = f
}

// IsFiltering returns true when the user is actively typing in the filter input.
func (s *Sidebar) IsFiltering() bool {
	return s.list.SettingFilter()
}

// Update handles messages for the sidebar.
func (s Sidebar) Update(msg tea.Msg) (Sidebar, tea.Cmd) {
	// Block key input when unfocused, but always allow non-key messages
	// (e.g., the list's internal FilterMatchesMsg) to pass through.
	if !s.focused {
		if _, isKey := msg.(tea.KeyMsg); isKey {
			return s, nil
		}
	}
	var cmd tea.Cmd
	s.list, cmd = s.list.Update(msg)
	return s, cmd
}

// View renders the sidebar.
func (s Sidebar) View() string {
	style := SidebarStyle
	if s.focused {
		style = SidebarFocusedStyle
	}

	title := TitleStyle.Render(" Jobs ")
	count := lipgloss.NewStyle().Foreground(ColorMuted).Render(
		fmt.Sprintf("(%d)", len(s.jobs)),
	)
	filterLabel := lipgloss.NewStyle().Foreground(ColorWarning).Bold(true).Render(
		filterLabels[s.filterIndex],
	)
	header := title + count + " " + filterLabel

	content := s.list.View()
	if len(s.jobs) == 0 {
		content = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Padding(1, 1).
			Render("No jobs found")
	}

	inner := lipgloss.JoinVertical(lipgloss.Left, header, content)

	return style.Width(s.width).Height(s.height).Render(
		padOrTruncate(inner, s.width, s.height),
	)
}

// padOrTruncate ensures content fits within the given dimensions.
func padOrTruncate(content string, width, height int) string {
	lines := strings.Split(content, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	for len(lines) < height {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}
