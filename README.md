# procrastinate-cli

A TUI monitor for [Procrastinate](https://github.com/procrastinate-org/procrastinate) PostgreSQL task queues. Gives you real-time visibility into your queues without leaving the terminal.

## Features

- **Live job stream** — watch jobs arrive in real-time via PostgreSQL LISTEN/NOTIFY
- **Status breakdown** — see job counts by status (todo, doing, succeeded, failed, etc.) with visual bars
- **Orphaned job detection** — find stuck jobs with dead workers or stale todo items
- **Job detail view** — inspect any job's args, events, and metadata
- **Multi-queue support** — switch between queues at runtime
- **Multiple connections** — switch between database connections on the fly

## Installation

```bash
go install github.com/matthewmyrick/procrastinate-cli@latest
```

Or build from source:

```bash
git clone https://github.com/matthewmyrick/procrastinate-cli.git
cd procrastinate-cli
go build -o procrastinate-cli .
```

## Configuration

Create a config file at `~/.config/procrastinate-cli/config.yaml`:

```yaml
default_queue: "default"
poll_interval: 5s
orphan_threshold: 30m
default_connection: "local-dev"

connections:
  - name: "local-dev"
    host: "localhost"
    port: 5432
    database: "myapp"
    username: "postgres"
    password: "secret"
    sslmode: "disable"
```

See `config.yaml.example` for a full example with multiple connections.

Config file search order:
1. `--config` flag
2. `$PROCRASTINATE_CONFIG` environment variable
3. `~/.config/procrastinate-cli/config.yaml`
4. `./config.yaml`

## Usage

```bash
# Launch with default config
procrastinate-cli

# Specify config file
procrastinate-cli --config /path/to/config.yaml

# Override queue and connection
procrastinate-cli --queue emails --connection staging-readonly
```

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `j` / `k` | Navigate job list |
| `Tab` | Switch focus between sidebar and detail pane |
| `[` / `]` | Switch tabs (Status / Live / Orphaned) |
| `Enter` | View job details |
| `Esc` | Close overlay / go back |
| `Q` | Switch queue |
| `C` | Switch connection |
| `q` | Quit |

## Project Structure

```
cli/       — Cobra CLI commands
config/    — YAML config parsing
db/        — PostgreSQL queries, connection management, LISTEN/NOTIFY
tui/       — Bubble Tea TUI components
```
