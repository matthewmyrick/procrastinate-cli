-- Setup script for local testing of procrastinate-cli
-- Run with: psql -U <your_user> -d <your_db> -f scripts/setup_test_db.sql

-- Create the Procrastinate enum type
DO $$ BEGIN
    CREATE TYPE procrastinate_job_status AS ENUM (
        'todo', 'doing', 'succeeded', 'failed', 'cancelled', 'aborting', 'aborted'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE procrastinate_job_event_type AS ENUM (
        'deferred', 'started', 'deferred_for_retry', 'failed', 'succeeded',
        'cancelled', 'abort_requested', 'aborted', 'scheduled', 'retried'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Workers table
CREATE TABLE IF NOT EXISTS procrastinate_workers (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    last_heartbeat timestamptz DEFAULT NOW()
);

-- Jobs table
CREATE TABLE IF NOT EXISTS procrastinate_jobs (
    id bigserial PRIMARY KEY,
    queue_name varchar(128) NOT NULL,
    task_name varchar(128) NOT NULL,
    priority integer DEFAULT 0 NOT NULL,
    lock text,
    queueing_lock text,
    args jsonb DEFAULT '{}' NOT NULL,
    status procrastinate_job_status DEFAULT 'todo' NOT NULL,
    scheduled_at timestamptz,
    attempts integer DEFAULT 0 NOT NULL,
    abort_requested boolean DEFAULT false NOT NULL,
    worker_id bigint REFERENCES procrastinate_workers(id)
);

-- Events table
CREATE TABLE IF NOT EXISTS procrastinate_events (
    id bigserial PRIMARY KEY,
    job_id bigint NOT NULL REFERENCES procrastinate_jobs(id) ON DELETE CASCADE,
    type procrastinate_job_event_type NOT NULL,
    at timestamptz DEFAULT NOW() NOT NULL
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_procrastinate_jobs_queue_status ON procrastinate_jobs(queue_name, status);
CREATE INDEX IF NOT EXISTS idx_procrastinate_events_job_id ON procrastinate_events(job_id);

-- LISTEN/NOTIFY function (simplified version of what Procrastinate creates)
CREATE OR REPLACE FUNCTION procrastinate_notify_job()
RETURNS trigger AS $$
BEGIN
    PERFORM pg_notify('procrastinate_any_queue_v1',
        json_build_object('type', 'job_inserted', 'job_id', NEW.id)::text);
    PERFORM pg_notify('procrastinate_queue_v1#' || NEW.queue_name,
        json_build_object('type', 'job_inserted', 'job_id', NEW.id)::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS procrastinate_job_notify ON procrastinate_jobs;
CREATE TRIGGER procrastinate_job_notify
    AFTER INSERT ON procrastinate_jobs
    FOR EACH ROW EXECUTE FUNCTION procrastinate_notify_job();

-- ============================================================
-- SEED DATA
-- ============================================================

-- Insert some workers (one alive, one dead)
INSERT INTO procrastinate_workers (last_heartbeat) VALUES
    (NOW()),                          -- worker 1: alive
    (NOW() - interval '2 hours');     -- worker 2: dead

-- Insert jobs in "default" queue
INSERT INTO procrastinate_jobs (queue_name, task_name, args, status, scheduled_at, attempts, worker_id) VALUES
    ('default', 'send_email',        '{"to": "alice@example.com", "subject": "Welcome"}',  'succeeded', NOW() - interval '30 minutes', 1, 1),
    ('default', 'send_email',        '{"to": "bob@example.com", "subject": "Reset"}',      'succeeded', NOW() - interval '25 minutes', 1, 1),
    ('default', 'process_payment',   '{"order_id": 1001, "amount": 49.99}',                'succeeded', NOW() - interval '20 minutes', 1, 1),
    ('default', 'process_payment',   '{"order_id": 1002, "amount": 129.00}',               'failed',    NOW() - interval '18 minutes', 3, 1),
    ('default', 'sync_inventory',    '{"warehouse": "us-east"}',                            'doing',     NOW() - interval '2 hours', 1, 2),  -- orphaned: dead worker
    ('default', 'generate_report',   '{"report": "monthly", "month": "2025-01"}',          'doing',     NOW() - interval '15 minutes', 1, 1),
    ('default', 'send_email',        '{"to": "carol@example.com", "subject": "Invoice"}',  'todo',      NOW() - interval '10 minutes', 0, NULL),
    ('default', 'process_payment',   '{"order_id": 1003, "amount": 75.50}',                'todo',      NOW() - interval '5 minutes', 0, NULL),
    ('default', 'cleanup_sessions',  '{"older_than": "7d"}',                               'todo',      NOW() - interval '2 minutes', 0, NULL),
    ('default', 'send_notification', '{"user_id": 42, "type": "promo"}',                   'todo',      NOW(), 0, NULL),
    ('default', 'process_payment',   '{"order_id": 1004, "amount": 200.00}',               'cancelled', NOW() - interval '1 hour', 0, NULL),
    ('default', 'send_email',        '{"to": "dave@example.com", "subject": "Reminder"}',  'aborted',   NOW() - interval '45 minutes', 2, 1);

-- Insert jobs in "emails" queue
INSERT INTO procrastinate_jobs (queue_name, task_name, args, status, scheduled_at, attempts) VALUES
    ('emails', 'send_bulk_email',    '{"campaign": "winter_sale", "batch": 1}',  'succeeded', NOW() - interval '1 hour', 1),
    ('emails', 'send_bulk_email',    '{"campaign": "winter_sale", "batch": 2}',  'doing',     NOW() - interval '50 minutes', 1),
    ('emails', 'send_bulk_email',    '{"campaign": "winter_sale", "batch": 3}',  'todo',      NOW() - interval '45 minutes', 0);

-- Insert jobs in "reports" queue (some old orphaned ones)
INSERT INTO procrastinate_jobs (queue_name, task_name, args, status, scheduled_at, attempts, worker_id) VALUES
    ('reports', 'generate_pdf',      '{"report_id": 501}',   'doing',     NOW() - interval '3 hours', 1, 2),  -- orphaned: dead worker
    ('reports', 'generate_pdf',      '{"report_id": 502}',   'todo',      NOW() - interval '2 hours', 0, NULL);  -- orphaned: stale todo

-- Insert events for the jobs
INSERT INTO procrastinate_events (job_id, type, at)
SELECT id, 'deferred', scheduled_at FROM procrastinate_jobs WHERE status IN ('todo', 'doing', 'succeeded', 'failed', 'cancelled', 'aborted');

INSERT INTO procrastinate_events (job_id, type, at)
SELECT id, 'started', scheduled_at + interval '1 second' FROM procrastinate_jobs WHERE status IN ('doing', 'succeeded', 'failed', 'aborted');

INSERT INTO procrastinate_events (job_id, type, at)
SELECT id, 'succeeded', scheduled_at + interval '5 seconds' FROM procrastinate_jobs WHERE status = 'succeeded';

INSERT INTO procrastinate_events (job_id, type, at)
SELECT id, 'failed', scheduled_at + interval '3 seconds' FROM procrastinate_jobs WHERE status = 'failed';

INSERT INTO procrastinate_events (job_id, type, at)
SELECT id, 'cancelled', scheduled_at + interval '2 seconds' FROM procrastinate_jobs WHERE status = 'cancelled';

INSERT INTO procrastinate_events (job_id, type, at)
SELECT id, 'aborted', scheduled_at + interval '10 seconds' FROM procrastinate_jobs WHERE status = 'aborted';

-- Summary
SELECT
    queue_name,
    status,
    COUNT(*) as count
FROM procrastinate_jobs
GROUP BY queue_name, status
ORDER BY queue_name, status;
