package ui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	up      key.Binding
	down    key.Binding
	enter   key.Binding
	esc     key.Binding
	search  key.Binding
	filter  key.Binding
	export  key.Binding
	stats   key.Binding
	refresh key.Binding
	quit    key.Binding
}

func defaultKeys() keyMap {
	return keyMap{
		up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "move up")),
		down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "move down")),
		enter:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open")),
		esc:     key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		search:  key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
		filter:  key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "filter")),
		export:  key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "export")),
		stats:   key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "stats")),
		refresh: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
		quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.up, k.down, k.enter, k.esc, k.search, k.filter, k.stats, k.refresh, k.quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.up, k.down, k.enter, k.esc}, {k.search, k.filter, k.export, k.stats, k.refresh, k.quit}}
}
