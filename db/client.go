package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Client manages PostgreSQL connections for querying.
type Client struct {
	pool    *pgxpool.Pool
	connStr string
}

// NewClient creates a new database client with a connection pool.
func NewClient(connStr string) (*Client, error) {
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("creating connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	return &Client{pool: pool, connStr: connStr}, nil
}

// Close shuts down the connection pool.
func (c *Client) Close() {
	if c.pool != nil {
		c.pool.Close()
	}
}

// Pool returns the underlying connection pool for running queries.
func (c *Client) Pool() *pgxpool.Pool {
	return c.pool
}

// NewListenerConn creates a dedicated connection for LISTEN/NOTIFY.
// This is separate from the pool because WaitForNotification blocks.
func (c *Client) NewListenerConn(ctx context.Context) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, c.connStr)
	if err != nil {
		return nil, fmt.Errorf("creating listener connection: %w", err)
	}
	return conn, nil
}
