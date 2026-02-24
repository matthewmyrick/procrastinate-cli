package tui

import (
	"time"

	"github.com/matthewmyrick/procrastinate-cli/db"
)

// Data refresh messages returned by fetch commands.
// Each carries a gen field to detect stale results from old connections/queues.

type jobsLoadedMsg struct {
	jobs []db.Job
	err  error
	gen  uint64
}

type statusCountsMsg struct {
	counts []db.StatusCount
	err    error
	gen    uint64
}

type orphanedJobsMsg struct {
	jobs []db.Job
	err  error
	gen  uint64
}

type recentJobsMsg struct {
	jobs []db.Job
	err  error
	gen  uint64
}

type jobDetailMsg struct {
	job    *db.Job
	events []db.JobEvent
	err    error
	gen    uint64
}

type queuesLoadedMsg struct {
	queues []string
	err    error
	gen    uint64
}

// notificationMsg wraps a LISTEN/NOTIFY event.
type notificationMsg struct {
	notification db.Notification
}

// tickMsg fires on each poll interval.
type tickMsg time.Time

// clearToastMsg dismisses the toast notification.
type clearToastMsg struct{}

// errMsg carries an error to display.
type errMsg struct {
	err error
}
