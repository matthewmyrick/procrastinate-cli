package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all key bindings for the application.
type KeyMap struct {
	Quit        key.Binding
	TabNext     key.Binding
	TabPrev     key.Binding
	FocusNext   key.Binding
	FocusPrev   key.Binding
	Up          key.Binding
	Down        key.Binding
	Enter       key.Binding
	Back        key.Binding
	SwitchQueue key.Binding
	SwitchConn  key.Binding
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
	}
}

// HelpKeys returns key bindings formatted for the help bar.
func (k KeyMap) HelpKeys() []key.Binding {
	return []key.Binding{
		k.Quit, k.FocusNext, k.TabNext, k.TabPrev,
		k.Enter, k.Back, k.SwitchQueue, k.SwitchConn,
	}
}
