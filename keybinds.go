package main

import (
	"github.com/charmbracelet/bubbles/key"
)

type KeyMap struct {
	Save       key.Binding
	Quit       key.Binding
	Help       key.Binding
	Copy       key.Binding
	Cut        key.Binding
	Paste      key.Binding
	Undo       key.Binding
	Redo       key.Binding
	SelectAll  key.Binding
	GoToLine   key.Binding
	Find       key.Binding
	FindNext   key.Binding
	FindPrev   key.Binding
	Delete     key.Binding
	ShiftLeft  key.Binding
	ShiftDown  key.Binding
	ShiftUp    key.Binding
	ShiftRight key.Binding
	AltLeft    key.Binding
	AltRight   key.Binding
}

var keys = KeyMap{
	Save: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save file"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+q"),
		key.WithHelp("ctrl+q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("ctrl+h"),
		key.WithHelp("ctrl+h", "toggle help"),
	),
	Copy: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "copy"),
	),
	Cut: key.NewBinding(
		key.WithKeys("ctrl+x"),
		key.WithHelp("ctrl+x", "cut"),
	),
	Paste: key.NewBinding(
		key.WithKeys("ctrl+v"),
		key.WithHelp("ctrl+v", "paste"),
	),
	Undo: key.NewBinding(
		key.WithKeys("ctrl+z"),
		key.WithHelp("ctrl+z", "undo"),
	),
	Redo: key.NewBinding(
		key.WithKeys("ctrl+y"),
		key.WithHelp("ctrl+y", "redo"),
	),
	SelectAll: key.NewBinding(
		key.WithKeys("ctrl+a"),
		key.WithHelp("ctrl+a", "select all"),
	),
	GoToLine: key.NewBinding(
		key.WithKeys("ctrl+g"),
		key.WithHelp("ctrl+g", "go to line"),
	),
	Find: key.NewBinding(
		key.WithKeys("ctrl+f"),
		key.WithHelp("ctrl+f", "find"),
	),
	FindNext: key.NewBinding(
		key.WithKeys("ctrl+n"),
		key.WithHelp("ctrl+n", "find next"),
	),
	FindPrev: key.NewBinding(
		key.WithKeys("ctrl+l"),
		key.WithHelp("ctrl+l", "find previous"),
	),
	Delete: key.NewBinding(
		key.WithKeys("backspace"),
		key.WithHelp("backspace", "delete"),
	),
	ShiftLeft: key.NewBinding(
		key.WithKeys("shift+left"),
		key.WithHelp("shift+left", "select text left"),
	),
	ShiftRight: key.NewBinding(
		key.WithKeys("shift+right"),
		key.WithHelp("shift+right", "select text right"),
	),
	ShiftUp: key.NewBinding(
		key.WithKeys("shift+up"),
		key.WithHelp("shift+up", "select text up"),
	),
	ShiftDown: key.NewBinding(
		key.WithKeys("shift+down"),
		key.WithHelp("shift+down", "select text down"),
	),
	AltLeft: key.NewBinding(
		key.WithKeys("alt+left"),
		key.WithHelp("alt+left", "select previous word"),
	),
	AltRight: key.NewBinding(
		key.WithKeys("alt+right"),
		key.WithHelp("alt+right", "select next word"),
	),
}
