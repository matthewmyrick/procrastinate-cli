package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

// Listener manages a dedicated PostgreSQL connection for LISTEN/NOTIFY.
type Listener struct {
	conn     *pgx.Conn
	queue    string
	notifyCh chan Notification
	cancel   context.CancelFunc
}

// NewListener creates a listener bound to a queue.
// The provided connection must be a raw pgx.Conn (not from a pool).
func NewListener(conn *pgx.Conn, queue string) *Listener {
	return &Listener{
		conn:     conn,
		queue:    queue,
		notifyCh: make(chan Notification, 64),
	}
}

// Start begins listening for notifications. It blocks internally and sends
// parsed notifications to the channel returned by Notifications().
// Call Stop() to shut it down.
func (l *Listener) Start(ctx context.Context) error {
	ctx, l.cancel = context.WithCancel(ctx)

	if err := l.subscribe(ctx); err != nil {
		return err
	}

	go l.loop(ctx)
	return nil
}

// Stop cancels the listener and closes the connection.
func (l *Listener) Stop() {
	if l.cancel != nil {
		l.cancel()
	}
	if l.conn != nil {
		l.conn.Close(context.Background())
	}
	close(l.notifyCh)
}

// Notifications returns the channel that receives parsed notifications.
func (l *Listener) Notifications() <-chan Notification {
	return l.notifyCh
}

// SwitchQueue unsubscribes from the old queue and subscribes to the new one.
func (l *Listener) SwitchQueue(ctx context.Context, newQueue string) error {
	if err := l.unsubscribe(ctx); err != nil {
		return err
	}
	l.queue = newQueue
	return l.subscribe(ctx)
}

func (l *Listener) subscribe(ctx context.Context) error {
	queueChannel := fmt.Sprintf("procrastinate_queue_v1#%s", l.queue)
	_, err := l.conn.Exec(ctx, fmt.Sprintf("LISTEN %s", pgx.Identifier{queueChannel}.Sanitize()))
	if err != nil {
		return fmt.Errorf("LISTEN %s: %w", queueChannel, err)
	}

	_, err = l.conn.Exec(ctx, "LISTEN procrastinate_any_queue_v1")
	if err != nil {
		return fmt.Errorf("LISTEN procrastinate_any_queue_v1: %w", err)
	}

	return nil
}

func (l *Listener) unsubscribe(ctx context.Context) error {
	_, err := l.conn.Exec(ctx, "UNLISTEN *")
	return err
}

func (l *Listener) loop(ctx context.Context) {
	for {
		notification, err := l.conn.WaitForNotification(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return // context cancelled, clean shutdown
			}
			log.Printf("listener error: %v", err)
			return
		}

		var n Notification
		if err := json.Unmarshal([]byte(notification.Payload), &n); err != nil {
			log.Printf("failed to parse notification payload: %v", err)
			continue
		}

		select {
		case l.notifyCh <- n:
		case <-ctx.Done():
			return
		}
	}
}
