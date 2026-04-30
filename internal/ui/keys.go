package ui

import (
	"fmt"
	"time"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
)

func (a *App) setupKeys() {
	a.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if a.waitingForDelete {
			switch event.Rune() {
			case 'y', 'Y':
				a.waitingForDelete = false
				a.deleteSelected()
				return nil
			default:
				a.waitingForDelete = false
				a.updateHeader()
				a.updateHint()
				return nil
			}
		}

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
				return a.handleDKey()
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
			case 'c':
				a.createJobFromCronjob()
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

func (a *App) handleDKey() *tcell.EventKey {
	if a.pendingDKey {
		a.pendingDKey = false
		a.confirmDelete()
		return nil
	}

	a.pendingDKey = true
	go func() {
		time.Sleep(500 * time.Millisecond)
		a.tapp.QueueUpdateDraw(func() {
			if a.pendingDKey {
				a.pendingDKey = false
			}
		})
	}()
	return nil
}

func (a *App) confirmDelete() {
	ns, name := a.selectedResource()
	if name == "" {
		return
	}
	a.waitingForDelete = true
	msg := fmt.Sprintf("[red]delete %s %s? (y/n):[-]", a.resource, name)
	if ns != "" {
		msg = fmt.Sprintf("[red]delete %s %s/%s? (y/n):[-]", a.resource, ns, name)
	}
	a.renderHeader(msg)
	a.updateHint()
}

func (a *App) deleteSelected() {
	ns, name := a.selectedResource()
	if name == "" {
		return
	}

	go func() {
		_, err := a.client.Delete(a.resource, ns, name)
		a.tapp.QueueUpdateDraw(func() {
			if err != nil {
				a.renderHeader(fmt.Sprintf("[red]error deleting: %s[-]", err))
			} else {
				a.renderHeader(fmt.Sprintf("[gray]deleted: %s[-]", name))
				a.reload()
			}
			go func() {
				time.Sleep(2000 * time.Millisecond)
				a.tapp.QueueUpdateDraw(a.updateHeader)
			}()
		})
	}()
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

func (a *App) createJobFromCronjob() {
	if a.resource != "cronjobs" {
		a.renderHeader("[yellow]⚠ Can only create jobs from cronjobs[-]")
		return
	}

	ns, name := a.selectedResource()
	if name == "" {
		return
	}

	go func() {
		jobName, err := a.client.CreateJobFromCronjob(ns, name)
		a.tapp.QueueUpdateDraw(func() {
			if err != nil {
				a.renderHeader(fmt.Sprintf("[red]✗ error creating job: %s[-]", err))
			} else {
				a.renderHeader(fmt.Sprintf("[green]✓ created job: %s[-]", jobName))
				time.AfterFunc(2*time.Second, func() {
					a.tapp.QueueUpdateDraw(a.updateHeader)
				})
			}
		})
	}()
}
