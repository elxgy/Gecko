package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Editor struct {
	app        *tview.Application
	textArea   *tview.TextArea
	statusBar  *tview.TextView
	mainLayout *tview.Flex
	filename   string
	modified   bool
}

func NewEditor() *Editor {
	app := tview.NewApplication()

	textArea := tview.NewTextArea().
		SetPlaceholder("Start typing... Press Ctrl+G for help").
		SetWrap(false)

	statusBar := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[white]No file | Press Ctrl+G for help")

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(textArea, 0, 1, true).
		AddItem(statusBar, 1, 0, false)

	editor := &Editor{
		app:        app,
		textArea:   textArea,
		statusBar:  statusBar,
		mainLayout: flex,
		filename:   "",
		modified:   false,
	}

	app.SetRoot(flex, true)
	editor.setupKeyBindings()

	return editor
}

func (e *Editor) setupKeyBindings() {
	e.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return e.handleKeyPress(event)
	})

	e.textArea.SetChangedFunc(func() {
		e.modified = true
		e.updateStatusBar()
	})
}

func (e *Editor) handleKeyPress(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyCtrlS:
		e.saveFile()
		return nil
	case tcell.KeyCtrlO:
		e.showOpenDialog()
		return nil
	case tcell.KeyCtrlQ:
		if e.modified {
			e.showQuitDialog()
		} else {
			e.app.Stop()
		}
		return nil
	case tcell.KeyCtrlG:
		e.showHelp()
		return nil
	case tcell.KeyCtrlF:
		e.showFindDialog()
		return nil
	case tcell.KeyCtrlA:
		e.textArea.Select(0, len(e.textArea.GetText()))
		return nil
	case tcell.KeyCtrlC, tcell.KeyCtrlX, tcell.KeyCtrlV:
		return event
	}

	return event
}

func (e *Editor) saveFile() {
	if e.filename == "" {
		e.showSaveAsDialog()
		return
	}

	content := e.textArea.GetText()
	err := os.WriteFile(e.filename, []byte(content), 0644)
	if err != nil {
		e.showError(fmt.Sprintf("Error saving file: %v", err))
		return
	}

	e.modified = false
	e.updateStatusBar()
	e.showMessage("File saved successfully")
}

func (e *Editor) showSaveAsDialog() {
	e.filename = "untitled.txt"
	e.saveFile()
}

func (e *Editor) showOpenDialog() {
	e.showMessage("Open dialog - TODO: Implement file browser")
}

func (e *Editor) showFindDialog() {
	e.showMessage("Find dialog - TODO: Implement search functionality")
}

func (e *Editor) showQuitDialog() {
	modal := tview.NewModal().
		SetText("File has been modified. Save before quitting?").
		AddButtons([]string{"Save & Quit", "Quit without saving", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			switch buttonIndex {
			case 0:
				e.saveFile()
				e.app.Stop()
			case 1:
				e.app.Stop()
			case 2:
				e.app.SetRoot(e.mainLayout, true)
			}
		})

	e.app.SetRoot(modal, true)
}

func (e *Editor) showMessage(message string) {
	e.statusBar.SetText(fmt.Sprintf("[green]%s[white]", message))
	go func() {
		time.Sleep(2 * time.Second)
		e.updateStatusBar()
	}()
}

func (e *Editor) showError(message string) {
	e.statusBar.SetText(fmt.Sprintf("[red]%s[white]", message))
	go func() {
		time.Sleep(3 * time.Second)
		e.updateStatusBar()
	}()
}

func (e *Editor) updateStatusBar() {
	modifiedIndicator := ""
	if e.modified {
		modifiedIndicator = " [red]●[white]"
	}

	filename := e.filename
	if filename == "" {
		filename = "No file"
	}

	text := e.textArea.GetText()
	lines := len(strings.Split(text, "\n"))

	status := fmt.Sprintf("%s%s | %d lines | Ctrl+G for help",
		filename,
		modifiedIndicator,
		lines)

	e.statusBar.SetText(status)
}

func (e *Editor) showHelp() {
	helpText := `
Go IDE - Micro-style Editor

Keyboard Shortcuts:
  Ctrl+S - Save file
  Ctrl+O - Open file
  Ctrl+Q - Quit
  Ctrl+G - Show this help
  Ctrl+F - Find text
  Ctrl+A - Select all
  Ctrl+C - Copy
  Ctrl+X - Cut  
  Ctrl+V - Paste

Navigation:
  Arrow Keys - Move cursor
  Home/End - Start/End of line
  Ctrl+Home - Start of file
  Ctrl+End - End of file
  Page Up/Down - Scroll page

Other:
  Just start typing to edit!

Press OK to continue...
`

	modal := tview.NewModal().
		SetText(helpText).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			e.app.SetRoot(e.mainLayout, true)
		})

	e.app.SetRoot(modal, true)
}

func (e *Editor) LoadFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	e.filename = filename
	e.textArea.SetText(string(content), false)
	e.modified = false
	e.updateStatusBar()

	return nil
}

func (e *Editor) SaveFile(filename string) error {
	content := e.textArea.GetText()
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return err
	}

	e.filename = filename
	e.modified = false
	e.updateStatusBar()

	return nil
}

func (e *Editor) Run() error {
	return e.app.Run()
}

func main() {
	editor := NewEditor()

	if len(os.Args) > 1 {
		filename := os.Args[1]
		if err := editor.LoadFile(filename); err != nil {
			log.Printf("Could not load file %s: %v", filename, err)
		}
	}

	if err := editor.Run(); err != nil {
		log.Fatal(err)
	}
}
