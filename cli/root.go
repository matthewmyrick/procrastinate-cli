package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/matthewmyrick/procrastinate-cli/config"
	"github.com/matthewmyrick/procrastinate-cli/tui"
)

var (
	configPath string
	queue      string
	connection string
)

var rootCmd = &cobra.Command{
	Use:   "procrastinate-cli",
	Short: "TUI monitor for Procrastinate PostgreSQL task queue",
	RunE:  runTUI,
}

func init() {
	rootCmd.Flags().StringVarP(&configPath, "config", "c", "", "path to config file")
	rootCmd.Flags().StringVarP(&queue, "queue", "q", "", "queue to monitor (overrides connection default)")
	rootCmd.Flags().StringVarP(&connection, "connection", "n", "", "connection name to use (defaults to first in config)")
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runTUI(cmd *cobra.Command, args []string) error {
	cfgPath, err := config.FindConfigPath(configPath)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	// Resolve which connection to start with: flag override, or first in list
	connName := cfg.Connections[0].Name
	if connection != "" {
		connName = connection
		if _, err := cfg.GetConnection(connName); err != nil {
			return fmt.Errorf("connection %q: %w", connName, err)
		}
	}

	// Queue override from flag (empty means use connection's default_queue)
	queueOverride := queue

	// TUI boots immediately — DB connection happens inside the TUI
	app := tui.NewApp(cfg, connName, queueOverride)
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui: %w", err)
	}

	return nil
}
