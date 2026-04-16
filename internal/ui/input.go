package ui

import (
	"strings"

	"k1s/internal/kubectl"

	"github.com/gdamore/tcell/v2"
)

func (a *App) openInput(mode inputMode, initial string) {
	a.imode = mode

	label := ":"
	if mode == inputFilter {
		label = "/"
	}

	a.cmdInput.SetLabel(label).SetText(initial)
	a.root.ResizeItem(a.cmdInput, 1, 0)
	a.root.ResizeItem(a.hint, 0, 0)

	a.cmdInput.SetDoneFunc(func(key tcell.Key) {
		text := strings.TrimSpace(a.cmdInput.GetText())
		a.closeInput()
		if key != tcell.KeyEnter {
			return
		}
		switch a.imode {
		case inputCommand:
			a.executeCommand(text)
		case inputFilter:
			a.filter = text
			a.renderTable(a.allCols, a.allRows)
			a.updateHeader()
		}
	})

	a.cmdInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			a.closeInput()
			return nil
		}
		return event
	})

	a.tapp.SetFocus(a.cmdInput)
}

func (a *App) closeInput() {
	a.root.ResizeItem(a.cmdInput, 0, 0)
	a.root.ResizeItem(a.hint, 1, 0)
	a.tapp.SetFocus(a.table)
}

func (a *App) executeCommand(cmd string) {
	if cmd == "" {
		return
	}
	switch cmd {
	case "q", "quit":
		a.tapp.Stop()
		return
	}

	// :ns <name>  →  switch namespace
	parts := strings.Fields(cmd)
	if len(parts) == 2 && (parts[0] == "ns" || parts[0] == "namespace") {
		a.client.Namespace = parts[1]
		a.updateHeader()
		a.reload()
		return
	}

	a.resource = kubectl.ResolveResource(cmd)
	a.filter = ""
	a.updateHeader()
	a.updateHint()
	a.loadResource()
}

func (a *App) selectNamespace() {
	_, name := a.selectedResource()
	if name == "" {
		return
	}
	a.client.Namespace = name
	a.prevNamespace = name
	a.resource = "pods"
	a.filter = ""
	a.updateHeader()
	a.updateHint()
	a.loadResource()
}

func (a *App) toggleAllNS() {
	if a.client.Namespace == "all" {
		a.client.Namespace = a.prevNamespace
	} else {
		a.prevNamespace = a.client.Namespace
		a.client.Namespace = "all"
	}
	a.reload()
}

func (a *App) reload() {
	a.updateHeader()
	a.loadResource()
}
