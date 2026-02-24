package db

import (
	"encoding/json"
	"time"
)

// JobStatus represents the procrastinate_job_status enum.
type JobStatus string

const (
	StatusTodo      JobStatus = "todo"
	StatusDoing     JobStatus = "doing"
	StatusSucceeded JobStatus = "succeeded"
	StatusFailed    JobStatus = "failed"
	StatusCancelled JobStatus = "cancelled"
	StatusAborting  JobStatus = "aborting"
	StatusAborted   JobStatus = "aborted"
)

// AllStatuses returns all known job statuses in display order.
func AllStatuses() []JobStatus {
	return []JobStatus{
		StatusTodo, StatusDoing, StatusSucceeded,
		StatusFailed, StatusCancelled, StatusAborting, StatusAborted,
	}
}

// Job represents a row from procrastinate_jobs.
type Job struct {
	ID             int64
	QueueName      string
	TaskName       string
	Priority       int
	Lock           *string
	QueueingLock   *string
	Args           json.RawMessage
	Status         JobStatus
	ScheduledAt    *time.Time
	Attempts       int
	AbortRequested bool
	WorkerID       *int64
}

// JobEvent represents a row from procrastinate_events.
type JobEvent struct {
	ID    int64
	JobID int64
	Type  string
	At    time.Time
}

// StatusCount holds a status and its count for the breakdown view.
type StatusCount struct {
	Status JobStatus
	Count  int64
}

// Notification represents a parsed LISTEN/NOTIFY payload.
type Notification struct {
	Type  string `json:"type"`
	JobID int64  `json:"job_id"`
}
