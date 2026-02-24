package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/matthewmyrick/procrastinate-cli/db"
)

func (a *App) fetchJobs() tea.Cmd {
	pool := a.dbClient.Pool()
	queue := a.currentQueue
	filter := a.sidebar.CurrentFilter()
	return func() tea.Msg {
		jobs, err := db.ListJobsFiltered(context.Background(), pool, queue, filter, 100, 0)
		return jobsLoadedMsg{jobs: jobs, err: err}
	}
}

func (a *App) fetchStatusCounts() tea.Cmd {
	pool := a.dbClient.Pool()
	queue := a.currentQueue
	return func() tea.Msg {
		counts, err := db.CountJobsByStatus(context.Background(), pool, queue)
		return statusCountsMsg{counts: counts, err: err}
	}
}

func (a *App) fetchOrphanedJobs() tea.Cmd {
	pool := a.dbClient.Pool()
	queue := a.currentQueue
	threshold := a.config.OrphanThreshold
	return func() tea.Msg {
		jobs, err := db.ListOrphanedJobs(context.Background(), pool, queue, threshold)
		return orphanedJobsMsg{jobs: jobs, err: err}
	}
}

func (a *App) fetchRecentJobs() tea.Cmd {
	pool := a.dbClient.Pool()
	queue := a.currentQueue
	return func() tea.Msg {
		since := time.Now().Add(-1 * time.Hour)
		jobs, err := db.ListRecentJobs(context.Background(), pool, queue, since)
		return recentJobsMsg{jobs: jobs, err: err}
	}
}

func (a *App) fetchQueues() tea.Cmd {
	pool := a.dbClient.Pool()
	return func() tea.Msg {
		queues, err := db.ListQueues(context.Background(), pool)
		return queuesLoadedMsg{queues: queues, err: err}
	}
}

func (a *App) fetchJobDetail(id int64) tea.Cmd {
	pool := a.dbClient.Pool()
	return func() tea.Msg {
		job, err := db.GetJob(context.Background(), pool, id)
		if err != nil {
			return jobDetailMsg{err: err}
		}
		events, err := db.GetJobEvents(context.Background(), pool, id)
		return jobDetailMsg{job: job, events: events, err: err}
	}
}

func (a *App) fetchActiveTabData() tea.Cmd {
	switch a.tabBar.Active() {
	case TabStatus:
		return a.fetchStatusCounts()
	case TabLive:
		return a.fetchRecentJobs()
	case TabOrphaned:
		return a.fetchOrphanedJobs()
	}
	return nil
}

func (a *App) tickCmd() tea.Cmd {
	interval := a.config.PollInterval
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (a *App) listenCmd() tea.Cmd {
	if a.listener == nil {
		return nil
	}
	ch := a.listener.Notifications()
	return func() tea.Msg {
		notif, ok := <-ch
		if !ok {
			return nil
		}
		return notificationMsg{notification: notif}
	}
}
