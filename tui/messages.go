package tui

import (
	"time"

	"github.com/matthewmyrick/procrastinate-cli/db"
)

// Data refresh messages returned by fetch commands.

type jobsLoadedMsg struct {
	jobs []db.Job
	err  error
}

type statusCountsMsg struct {
	counts []db.StatusCount
	err    error
}

type orphanedJobsMsg struct {
	jobs []db.Job
	err  error
}

type recentJobsMsg struct {
	jobs []db.Job
	err  error
}

type jobDetailMsg struct {
	job    *db.Job
	events []db.JobEvent
	err    error
}

type queuesLoadedMsg struct {
	queues []string
	err    error
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
