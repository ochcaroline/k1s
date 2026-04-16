package ui

import "fmt"

func (a *App) openDescribe() {
	ns, name := a.selectedResource()
	if name == "" {
		return
	}
	a.switchToDetail(fmt.Sprintf("[yellow]describe[-]  %s", name))
	go func() {
		out, _ := a.client.Describe(a.resource, ns, name)
		a.tapp.QueueUpdateDraw(func() {
			a.detail.SetText(out)
			a.detail.ScrollToBeginning()
		})
	}()
}

func (a *App) openYAML() {
	ns, name := a.selectedResource()
	if name == "" {
		return
	}
	a.switchToDetail(fmt.Sprintf("[cyan]yaml[-]  %s", name))
	go func() {
		out, _ := a.client.YAML(a.resource, ns, name)
		a.tapp.QueueUpdateDraw(func() {
			a.detail.SetText(out)
			a.detail.ScrollToBeginning()
		})
	}()
}

func (a *App) openLogs() {
	if a.resource != "pods" {
		return
	}
	ns, name := a.selectedResource()
	if name == "" {
		return
	}
	a.switchToDetail(fmt.Sprintf("[green]logs[-]  %s", name))
	go func() {
		out, _ := a.client.Logs(ns, name)
		a.tapp.QueueUpdateDraw(func() {
			a.detail.SetText(out)
			a.detail.ScrollToEnd()
		})
	}()
}

func (a *App) switchToDetail(subtitle string) {
	a.view = viewDetail
	a.detail.SetText("  loading...")
	a.pages.SwitchToPage("detail")
	a.tapp.SetFocus(a.detail)
	a.updateHint()
	a.renderHeader(subtitle)
}

func (a *App) switchToList() {
	a.view = viewList
	a.pages.SwitchToPage("list")
	a.tapp.SetFocus(a.table)
	a.updateHeader()
	a.updateHint()
}
