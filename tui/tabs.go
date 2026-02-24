package tui

import "github.com/charmbracelet/lipgloss"

// Tab indices
const (
	TabStatus   = 0
	TabLive     = 1
	TabOrphaned = 2
)

// TabNames are the display names for each tab.
var TabNames = []string{"Status", "Live", "Orphaned"}

// TabBar manages the tab strip in the detail pane.
type TabBar struct {
	tabs      []string
	activeTab int
	width     int
}

// NewTabBar creates a tab bar with the given tab names.
func NewTabBar(tabs []string) TabBar {
	return TabBar{
		tabs:      tabs,
		activeTab: 0,
	}
}

// Active returns the currently active tab index.
func (t *TabBar) Active() int {
	return t.activeTab
}

// SetActive sets the active tab by index.
func (t *TabBar) SetActive(i int) {
	if i >= 0 && i < len(t.tabs) {
		t.activeTab = i
	}
}

// Next moves to the next tab, wrapping around.
func (t *TabBar) Next() {
	t.activeTab = (t.activeTab + 1) % len(t.tabs)
}

// Prev moves to the previous tab, wrapping around.
func (t *TabBar) Prev() {
	t.activeTab = (t.activeTab - 1 + len(t.tabs)) % len(t.tabs)
}

// SetWidth sets the available width for rendering.
func (t *TabBar) SetWidth(w int) {
	t.width = w
}

// View renders the tab bar.
func (t TabBar) View() string {
	var rendered []string
	for i, name := range t.tabs {
		if i == t.activeTab {
			rendered = append(rendered, ActiveTabStyle.Render(name))
		} else {
			rendered = append(rendered, InactiveTabStyle.Render(name))
		}
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, rendered...)

	return TabBarStyle.Width(t.width).Render(row)
}
