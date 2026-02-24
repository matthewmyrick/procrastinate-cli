package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all key bindings for the application.
type KeyMap struct {
	Quit         key.Binding
	TabNext      key.Binding
	TabPrev      key.Binding
	FocusNext    key.Binding
	FocusPrev    key.Binding
	FocusLeft    key.Binding
	FocusRight   key.Binding
	Up           key.Binding
	Down         key.Binding
	Enter        key.Binding
	Back         key.Binding
	SwitchQueue  key.Binding
	SwitchConn   key.Binding
	FilterStatus key.Binding
	Dashboard    key.Binding
	Help         key.Binding
}

// DefaultKeyMap returns the default key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		TabNext: key.NewBinding(
			key.WithKeys("]"),
			key.WithHelp("]", "next tab"),
		),
		TabPrev: key.NewBinding(
			key.WithKeys("["),
			key.WithHelp("[", "prev tab"),
		),
		FocusNext: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch pane"),
		),
		FocusPrev: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "switch pane"),
		),
		FocusLeft: key.NewBinding(
			key.WithKeys("ctrl+g"),
			key.WithHelp("ctrl+g", "left pane"),
		),
		FocusRight: key.NewBinding(
			key.WithKeys("ctrl+h"),
			key.WithHelp("ctrl+h", "right pane"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("j", "down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "details"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		SwitchQueue: key.NewBinding(
			key.WithKeys("Q"),
			key.WithHelp("Q", "switch queue"),
		),
		SwitchConn: key.NewBinding(
			key.WithKeys("C"),
			key.WithHelp("C", "switch conn"),
		),
		FilterStatus: key.NewBinding(
			key.WithKeys("f", "F"),
			key.WithHelp("f", "filter"),
		),
		Dashboard: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "dashboard"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
	}
}

// HelpKeys returns a short set of key bindings for the bottom help bar.
func (k KeyMap) HelpKeys() []key.Binding {
	return []key.Binding{
		k.Help, k.Quit, k.Enter, k.Dashboard, k.FilterStatus, k.SwitchConn,
	}
}

// AllKeys returns all key bindings for the full help overlay.
func (k KeyMap) AllKeys() []key.Binding {
	return []key.Binding{
		k.Quit, k.Help,
		k.FocusNext, k.FocusPrev, k.FocusLeft, k.FocusRight,
		k.Up, k.Down, k.Enter, k.Back,
		k.TabNext, k.TabPrev, k.Dashboard,
		k.FilterStatus, k.SwitchQueue, k.SwitchConn,
	}
}
