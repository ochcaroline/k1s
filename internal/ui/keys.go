package ui

import (
	"time"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
)

func (a *App) setupKeys() {
	a.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case ':':
				a.openInput(inputCommand, "")
				return nil
			case '/':
				a.openInput(inputFilter, a.filter)
				return nil
			case 'e':
				a.openEdit()
				return nil
			case 'd':
				a.openDescribe()
				return nil
			case 'l':
				a.openLogs()
				return nil
			case 'y':
				a.copyName()
				return nil
			case 'R':
				a.reload()
				return nil
			case 'a':
				a.toggleAllNS()
				return nil
			}
		case tcell.KeyEnter:
			if a.resource == "namespaces" {
				a.selectNamespace()
			} else if a.resource == "_contexts" {
				a.selectContext()
			} else {
				a.openDescribe()
			}
			return nil
		case tcell.KeyEsc:
			if a.filter != "" {
				a.filter = ""
				a.renderTable(a.allCols, a.allRows)
				a.updateHeader()
			}
			return nil
		case tcell.KeyCtrlC:
			a.tapp.Stop()
			return nil
		}
		return event
	})

	a.detail.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			a.switchToList()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				a.switchToList()
				return nil
			case '/':
				a.openDetailSearch()
				return nil
			case 'n':
				a.nextMatch(1)
				return nil
			case 'N':
				a.nextMatch(-1)
				return nil
			case 'w':
				a.toggleWrap()
				return nil
			case 'y':
				a.openYAML()
				return nil
			case 'g':
				a.detail.ScrollToBeginning()
				return nil
			case 'G':
				a.detail.ScrollToEnd()
				return nil
			}
		case tcell.KeyCtrlC:
			a.tapp.Stop()
			return nil
		}
		return event
	})
}

func (a *App) copyName() {
	_, name := a.selectedResource()
	if name == "" {
		return
	}
	_ = clipboard.WriteAll(name)

	a.renderHeader("[gray]copied:[-] " + name)

	go func() {
		time.Sleep(1500 * time.Millisecond)
		a.tapp.QueueUpdateDraw(a.updateHeader)
	}()
}
