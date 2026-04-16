package ui

import (
	"strings"

	"k1s/internal/kubectl"

	"github.com/gdamore/tcell/v2"
)

func (a *App) openInput(mode inputMode, initial string) {
	a.imode = mode

	label := ":"
	if mode == inputFilter || mode == inputDetailSearch {
		label = "/"
	}

	a.cmdInput.SetLabel(label).SetText(initial)
	a.root.ResizeItem(a.cmdInput, 1, 0)
	a.root.ResizeItem(a.hint, 0, 0)

	// Live highlighting for detail search.
	if mode == inputDetailSearch {
		a.cmdInput.SetChangedFunc(func(text string) {
			a.applyDetailSearch(text)
		})
	} else {
		a.cmdInput.SetChangedFunc(nil)
	}

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
		case inputDetailSearch:
			// already applied live via ChangedFunc
		}
	})

	a.cmdInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			if a.imode == inputDetailSearch {
				a.applyDetailSearch("")
			}
			a.closeInput()
			return nil
		}
		return event
	})

	a.tapp.SetFocus(a.cmdInput)
}

func (a *App) closeInput() {
	a.cmdInput.SetChangedFunc(nil)
	a.root.ResizeItem(a.cmdInput, 0, 0)
	a.root.ResizeItem(a.hint, 1, 0)
	if a.view == viewDetail {
		a.tapp.SetFocus(a.detail)
	} else {
		a.tapp.SetFocus(a.table)
	}
}

func (a *App) openDetailSearch() {
	a.openInput(inputDetailSearch, a.detailSearch)
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

	// :ctx <name>  →  switch context
	if len(parts) == 2 && (parts[0] == "ctx" || parts[0] == "context") {
		if err := a.client.SwitchContext(parts[1]); err == nil {
			a.client.Context = parts[1]
			a.client.Namespace = a.client.CurrentNamespace()
			a.resource = "pods"
			a.filter = ""
			a.updateHeader()
			a.updateHint()
			a.loadResource()
		}
		return
	}

	a.resource = kubectl.ResolveResource(cmd)
	a.filter = ""
	a.updateHeader()
	a.updateHint()
	a.loadResource()
}

func (a *App) selectContext() {
	row, _ := a.table.GetSelection()
	if row < 1 {
		return
	}
	// kubectl get-contexts output: rows with current context have "*" in col 0,
	// rows without it have the context name in col 0 (strings.Fields strips indent).
	cell0 := strings.TrimSpace(a.table.GetCell(row, 0).Text)
	var name string
	if cell0 == "*" {
		if c := a.table.GetCell(row, 1); c != nil {
			name = strings.TrimSpace(c.Text)
		}
	} else {
		name = cell0
	}
	if name == "" {
		return
	}
	if err := a.client.SwitchContext(name); err != nil {
		return
	}
	a.client.Context = name
	a.client.Namespace = a.client.CurrentNamespace()
	a.resource = "pods"
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
