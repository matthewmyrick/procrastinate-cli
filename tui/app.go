package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/matthewmyrick/procrastinate-cli/config"
	"github.com/matthewmyrick/procrastinate-cli/db"
)

const (
	minWidth  = 80
	minHeight = 24
)

type focusPane int

const (
	focusSidebar focusPane = iota
	focusDetail
)

type overlayMode int

const (
	overlayNone overlayMode = iota
	overlayDetail
	overlayQueuePicker
	overlayConnPicker
)

// App is the root Bubble Tea model.
type App struct {
	config *config.Config

	// DB state — nil until connected
	dbClient  *db.Client
	listener  *db.Listener
	connected bool

	currentQueue  string
	currentConn   string
	queueOverride string // from --queue flag, empty means use connection default
	focus         focusPane
	overlay       overlayMode
	width         int
	height        int
	ready         bool
	lastError     error
	toast         string

	// Child components
	sidebar      Sidebar
	tabBar       TabBar
	statusView   StatusView
	liveView     LiveView
	orphanedView OrphanedView
	detailView   DetailView

	// Picker state
	queues        []string
	pickerIndex   int
	pickerItems   []string
	switchConnFn  func(string) tea.Cmd
	switchQueueFn func(string) tea.Cmd

	keys KeyMap
}

// connectedMsg is sent after a connection attempt completes.
type connectedMsg struct {
	client   *db.Client
	listener *db.Listener
	queue    string
	err      error
}

// NewApp creates the root TUI model. No DB connection yet — that happens on Init.
func NewApp(cfg *config.Config, connName, queueOverride string) *App {
	conn, _ := cfg.GetConnection(connName)
	initialQueue := conn.DefaultQueue
	if queueOverride != "" {
		initialQueue = queueOverride
	}

	return &App{
		config:        cfg,
		currentConn:   connName,
		currentQueue:  initialQueue,
		queueOverride: queueOverride,
		focus:         focusSidebar,
		overlay:       overlayNone,
		sidebar:       NewSidebar(30, 20),
		tabBar:        NewTabBar(TabNames),
		statusView:    NewStatusView(),
		liveView:      NewLiveView(),
		orphanedView:  NewOrphanedView(),
		detailView:    NewDetailView(),
		keys:          DefaultKeyMap(),
	}
}

func (a *App) Init() tea.Cmd {
	return a.connectCmd(a.currentConn, a.currentQueue)
}

// connectCmd attempts to connect to a database in the background.
func (a *App) connectCmd(connName, queue string) tea.Cmd {
	cfg := a.config
	return func() tea.Msg {
		conn, err := cfg.GetConnection(connName)
		if err != nil {
			return connectedMsg{err: err}
		}

		client, err := db.NewClient(config.ConnString(conn))
		if err != nil {
			return connectedMsg{err: fmt.Errorf("connect to %s: %w", connName, err)}
		}

		// Try to set up listener (non-fatal if it fails)
		var listener *db.Listener
		listenerConn, err := client.NewListenerConn(context.Background())
		if err == nil {
			listener = db.NewListener(listenerConn, queue)
			if startErr := listener.Start(context.Background()); startErr != nil {
				listener = nil
			}
		}

		return connectedMsg{client: client, listener: listener, queue: queue}
	}
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.ready = true
		a.recalcLayout()
		return a, nil

	case tea.KeyMsg:
		return a.handleKey(msg)

	case clearToastMsg:
		a.toast = ""

	case connectedMsg:
		if msg.err != nil {
			a.toast = fmt.Sprintf("Connection failed: %s", a.currentConn)
			a.connected = false
			a.dbClient = nil
			a.listener = nil
			cmds = append(cmds, a.clearToastCmd())
		} else {
			a.dbClient = msg.client
			a.listener = msg.listener
			a.connected = true
			a.lastError = nil
			// Start fetching data and polling
			cmds = append(cmds,
				a.fetchJobs(), a.fetchStatusCounts(), a.fetchQueues(),
				a.tickCmd(), a.listenCmd(),
			)
		}

	case jobsLoadedMsg:
		if msg.err != nil {
			a.lastError = msg.err
		} else {
			a.sidebar.SetJobs(msg.jobs)
			a.lastError = nil
		}

	case statusCountsMsg:
		if msg.err != nil {
			a.lastError = msg.err
		} else {
			a.statusView.SetCounts(msg.counts)
			a.lastError = nil
		}

	case recentJobsMsg:
		if msg.err != nil {
			a.lastError = msg.err
		} else {
			a.liveView.SetJobs(msg.jobs)
			a.lastError = nil
		}

	case orphanedJobsMsg:
		if msg.err != nil {
			a.lastError = msg.err
		} else {
			a.orphanedView.SetJobs(msg.jobs)
			a.lastError = nil
		}

	case queuesLoadedMsg:
		if msg.err != nil {
			a.lastError = msg.err
		} else {
			a.queues = msg.queues
		}

	case jobDetailMsg:
		if msg.err != nil {
			a.lastError = msg.err
		} else {
			a.detailView.SetJob(msg.job, msg.events)
			a.detailView.SetSize(a.width, a.height)
			a.overlay = overlayDetail
		}

	case notificationMsg:
		cmds = append(cmds, a.listenCmd(), a.fetchJobs(), a.fetchActiveTabData())

	case tickMsg:
		if a.connected {
			cmds = append(cmds, a.fetchJobs(), a.fetchActiveTabData(), a.tickCmd())
		}
	}

	return a, tea.Batch(cmds...)
}

func (a *App) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if a.overlay != overlayNone {
		return a.handleOverlayKey(msg)
	}

	switch {
	case key.Matches(msg, a.keys.Quit):
		return a, tea.Quit

	case key.Matches(msg, a.keys.FocusNext), key.Matches(msg, a.keys.FocusPrev):
		a.toggleFocus()
		return a, nil

	case key.Matches(msg, a.keys.TabNext):
		a.tabBar.Next()
		if a.connected {
			return a, a.fetchActiveTabData()
		}
		return a, nil

	case key.Matches(msg, a.keys.TabPrev):
		a.tabBar.Prev()
		if a.connected {
			return a, a.fetchActiveTabData()
		}
		return a, nil

	case key.Matches(msg, a.keys.Enter):
		if a.focus == focusSidebar && a.connected {
			if job := a.sidebar.SelectedJob(); job != nil {
				return a, a.fetchJobDetail(job.ID)
			}
		}
		return a, nil

	case key.Matches(msg, a.keys.SwitchQueue):
		a.openQueuePicker()
		return a, nil

	case key.Matches(msg, a.keys.SwitchConn):
		a.openConnPicker()
		return a, nil
	}

	if a.focus == focusSidebar {
		var cmd tea.Cmd
		a.sidebar, cmd = a.sidebar.Update(msg)
		return a, cmd
	}

	cmds := a.updateActiveView(msg)
	return a, tea.Batch(cmds...)
}

func (a *App) handleOverlayKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch a.overlay {
	case overlayDetail:
		if key.Matches(msg, a.keys.Back) {
			a.overlay = overlayNone
			a.detailView.SetVisible(false)
			return a, nil
		}
		var cmd tea.Cmd
		a.detailView, cmd = a.detailView.Update(msg)
		return a, cmd

	case overlayQueuePicker, overlayConnPicker:
		return a.handlePickerKey(msg)
	}

	return a, nil
}

func (a *App) handlePickerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keys.Back):
		a.overlay = overlayNone
		a.switchQueueFn = nil
		a.switchConnFn = nil
		return a, nil

	case key.Matches(msg, a.keys.Up):
		if a.pickerIndex > 0 {
			a.pickerIndex--
		}
		return a, nil

	case key.Matches(msg, a.keys.Down):
		if a.pickerIndex < len(a.pickerItems)-1 {
			a.pickerIndex++
		}
		return a, nil

	case key.Matches(msg, a.keys.Enter):
		if a.pickerIndex >= len(a.pickerItems) {
			return a, nil
		}
		selected := a.pickerItems[a.pickerIndex]
		a.overlay = overlayNone

		if a.switchQueueFn != nil {
			cmd := a.switchQueueFn(selected)
			a.switchQueueFn = nil
			a.switchConnFn = nil
			return a, cmd
		}
		if a.switchConnFn != nil {
			cmd := a.switchConnFn(selected)
			a.switchConnFn = nil
			a.switchQueueFn = nil
			return a, cmd
		}
		return a, nil
	}

	return a, nil
}

func (a *App) openQueuePicker() {
	if len(a.queues) == 0 {
		return
	}
	a.overlay = overlayQueuePicker
	a.pickerItems = a.queues
	a.pickerIndex = 0
	for i, q := range a.queues {
		if q == a.currentQueue {
			a.pickerIndex = i
			break
		}
	}
	a.switchQueueFn = func(queue string) tea.Cmd {
		a.currentQueue = queue
		if a.listener != nil {
			_ = a.listener.SwitchQueue(context.Background(), queue)
		}
		if a.connected {
			return tea.Batch(a.fetchJobs(), a.fetchActiveTabData(), a.fetchQueues())
		}
		return nil
	}
}

func (a *App) openConnPicker() {
	conns := make([]string, len(a.config.Connections))
	for i, c := range a.config.Connections {
		conns[i] = c.Name
	}
	a.overlay = overlayConnPicker
	a.pickerItems = conns
	a.pickerIndex = 0
	for i, c := range conns {
		if c == a.currentConn {
			a.pickerIndex = i
			break
		}
	}
	a.switchConnFn = func(connName string) tea.Cmd {
		conn, err := a.config.GetConnection(connName)
		if err != nil {
			a.lastError = err
			return nil
		}

		// Close old connection first
		if a.dbClient != nil {
			a.dbClient.Close()
			a.dbClient = nil
		}
		if a.listener != nil {
			a.listener.Stop()
			a.listener = nil
		}

		a.currentConn = connName
		a.currentQueue = conn.DefaultQueue
		a.connected = false
		a.lastError = nil
		a.sidebar.SetJobs(nil)

		return a.connectCmd(connName, conn.DefaultQueue)
	}
}

func (a *App) updateActiveView(msg tea.Msg) []tea.Cmd {
	var cmds []tea.Cmd
	switch a.tabBar.Active() {
	case TabStatus:
		var cmd tea.Cmd
		a.statusView, cmd = a.statusView.Update(msg)
		cmds = append(cmds, cmd)
	case TabLive:
		var cmd tea.Cmd
		a.liveView, cmd = a.liveView.Update(msg)
		cmds = append(cmds, cmd)
	case TabOrphaned:
		var cmd tea.Cmd
		a.orphanedView, cmd = a.orphanedView.Update(msg)
		cmds = append(cmds, cmd)
	}
	return cmds
}

func (a *App) View() string {
	if !a.ready {
		return "Initializing..."
	}

	if a.width < minWidth || a.height < minHeight {
		return fmt.Sprintf(
			"Terminal too small (%dx%d). Minimum: %dx%d",
			a.width, a.height, minWidth, minHeight,
		)
	}

	topBar := a.renderTopBar()
	helpBar := a.renderHelpBar()
	chromeH := lipgloss.Height(topBar) + lipgloss.Height(helpBar)
	borderH := 2 // rounded border top + bottom on panes
	contentHeight := a.height - chromeH - borderH

	sidebarWidth := a.calcSidebarWidth()
	detailWidth := a.width - sidebarWidth

	sidebarContent := a.sidebar.View()
	detailContent := a.renderDetail(detailWidth, contentHeight)

	content := lipgloss.JoinHorizontal(lipgloss.Top, sidebarContent, detailContent)
	base := lipgloss.JoinVertical(lipgloss.Left, topBar, content, helpBar)

	switch a.overlay {
	case overlayDetail:
		return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, a.detailView.View())
	case overlayQueuePicker:
		return a.renderOverlay(base, a.renderPicker("Switch Queue", a.pickerItems, a.pickerIndex))
	case overlayConnPicker:
		return a.renderOverlay(base, a.renderPicker("Switch Connection", a.pickerItems, a.pickerIndex))
	}

	if a.toast != "" {
		return a.renderToastOverlay(base)
	}

	return base
}

// --- Layout helpers ---

func (a *App) recalcLayout() {
	chromeH := 2 // top bar + help bar (1 line each)
	borderH := 2 // rounded border top + bottom on panes
	contentHeight := a.height - chromeH - borderH

	sidebarWidth := a.calcSidebarWidth()
	detailWidth := a.width - sidebarWidth

	a.sidebar.SetSize(sidebarWidth, contentHeight)
	a.sidebar.SetFocused(a.focus == focusSidebar)
	a.tabBar.SetWidth(detailWidth - 2)

	tabContentHeight := contentHeight - 2 // tab bar takes ~2 lines
	tabContentWidth := detailWidth - 2
	a.statusView.SetSize(tabContentWidth, tabContentHeight)
	a.liveView.SetSize(tabContentWidth, tabContentHeight)
	a.orphanedView.SetSize(tabContentWidth, tabContentHeight)
	a.detailView.SetSize(a.width, a.height)
}

func (a *App) calcSidebarWidth() int {
	w := a.width * 30 / 100
	if w < 20 {
		w = 20
	}
	if w > 40 {
		w = 40
	}
	return w
}

func (a *App) toggleFocus() {
	if a.focus == focusSidebar {
		a.focus = focusDetail
	} else {
		a.focus = focusSidebar
	}
	a.sidebar.SetFocused(a.focus == focusSidebar)
}

func (a *App) clearToastCmd() tea.Cmd {
	return tea.Tick(4*time.Second, func(t time.Time) tea.Msg {
		return clearToastMsg{}
	})
}

// Render helpers are in render.go
// Data commands are in commands.go
