package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ListQueues returns all distinct queue names.
func ListQueues(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	rows, err := pool.Query(ctx,
		`SELECT DISTINCT queue_name FROM procrastinate_jobs ORDER BY queue_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var queues []string
	for rows.Next() {
		var q string
		if err := rows.Scan(&q); err != nil {
			return nil, err
		}
		queues = append(queues, q)
	}
	return queues, rows.Err()
}

// ListJobs returns jobs for a queue, ordered by id DESC, with limit/offset.
func ListJobs(ctx context.Context, pool *pgxpool.Pool, queue string, limit, offset int) ([]Job, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, queue_name, task_name, priority, lock, queueing_lock,
		       args, status, scheduled_at, attempts, abort_requested, worker_id
		FROM procrastinate_jobs
		WHERE queue_name = $1
		ORDER BY id DESC
		LIMIT $2 OFFSET $3`,
		queue, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanJobs(rows)
}

// GetJob returns a single job by ID.
func GetJob(ctx context.Context, pool *pgxpool.Pool, id int64) (*Job, error) {
	row := pool.QueryRow(ctx, `
		SELECT id, queue_name, task_name, priority, lock, queueing_lock,
		       args, status, scheduled_at, attempts, abort_requested, worker_id
		FROM procrastinate_jobs
		WHERE id = $1`, id)

	var j Job
	err := row.Scan(
		&j.ID, &j.QueueName, &j.TaskName, &j.Priority, &j.Lock, &j.QueueingLock,
		&j.Args, &j.Status, &j.ScheduledAt, &j.Attempts, &j.AbortRequested, &j.WorkerID,
	)
	if err != nil {
		return nil, err
	}
	return &j, nil
}

// GetJobEvents returns events for a job, ordered by timestamp.
func GetJobEvents(ctx context.Context, pool *pgxpool.Pool, jobID int64) ([]JobEvent, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, job_id, type, at
		FROM procrastinate_events
		WHERE job_id = $1
		ORDER BY at ASC`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []JobEvent
	for rows.Next() {
		var e JobEvent
		if err := rows.Scan(&e.ID, &e.JobID, &e.Type, &e.At); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// CountJobsByStatus returns job counts grouped by status for a queue.
func CountJobsByStatus(ctx context.Context, pool *pgxpool.Pool, queue string) ([]StatusCount, error) {
	rows, err := pool.Query(ctx, `
		SELECT status, COUNT(*)
		FROM procrastinate_jobs
		WHERE queue_name = $1
		GROUP BY status
		ORDER BY status`, queue)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counts []StatusCount
	for rows.Next() {
		var sc StatusCount
		if err := rows.Scan(&sc.Status, &sc.Count); err != nil {
			return nil, err
		}
		counts = append(counts, sc)
	}
	return counts, rows.Err()
}

// ListRecentJobs returns jobs that were created after the given timestamp.
func ListRecentJobs(ctx context.Context, pool *pgxpool.Pool, queue string, since time.Time) ([]Job, error) {
	rows, err := pool.Query(ctx, `
		SELECT j.id, j.queue_name, j.task_name, j.priority, j.lock, j.queueing_lock,
		       j.args, j.status, j.scheduled_at, j.attempts, j.abort_requested, j.worker_id
		FROM procrastinate_jobs j
		JOIN procrastinate_events e ON e.job_id = j.id
		WHERE j.queue_name = $1
		  AND e.type = 'deferred'
		  AND e.at >= $2
		ORDER BY j.id DESC
		LIMIT 200`, queue, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanJobs(rows)
}

// ListOrphanedJobs returns jobs that appear stuck or abandoned.
// It combines two strategies:
// 1. Jobs in 'doing' with a dead/missing worker (stale heartbeat)
// 2. Jobs in 'todo' sitting too long without progress (excluding future-scheduled)
func ListOrphanedJobs(ctx context.Context, pool *pgxpool.Pool, queue string, threshold time.Duration) ([]Job, error) {
	rows, err := pool.Query(ctx, `
		-- Doing jobs with dead or missing worker
		SELECT j.id, j.queue_name, j.task_name, j.priority, j.lock, j.queueing_lock,
		       j.args, j.status, j.scheduled_at, j.attempts, j.abort_requested, j.worker_id
		FROM procrastinate_jobs j
		LEFT JOIN procrastinate_workers w ON j.worker_id = w.id
		WHERE j.queue_name = $1
		  AND j.status = 'doing'
		  AND (w.id IS NULL OR w.last_heartbeat < NOW() - $2::interval)

		UNION ALL

		-- Todo jobs sitting too long
		SELECT j.id, j.queue_name, j.task_name, j.priority, j.lock, j.queueing_lock,
		       j.args, j.status, j.scheduled_at, j.attempts, j.abort_requested, j.worker_id
		FROM procrastinate_jobs j
		WHERE j.queue_name = $1
		  AND j.status = 'todo'
		  AND (j.scheduled_at IS NULL OR j.scheduled_at <= NOW())
		  AND NOT EXISTS (
		    SELECT 1 FROM procrastinate_events e
		    WHERE e.job_id = j.id AND e.at > NOW() - $2::interval
		  )
		ORDER BY id ASC`,
		queue, threshold.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanJobs(rows)
}

// scanJobs is a helper that scans job rows into a slice.
func scanJobs(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]Job, error) {
	var jobs []Job
	for rows.Next() {
		var j Job
		if err := rows.Scan(
			&j.ID, &j.QueueName, &j.TaskName, &j.Priority, &j.Lock, &j.QueueingLock,
			&j.Args, &j.Status, &j.ScheduledAt, &j.Attempts, &j.AbortRequested, &j.WorkerID,
		); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}
