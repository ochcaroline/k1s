package ui

import (
	"fmt"
	"strings"

	"k1s/internal/resource"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *App) loadResource() {
	a.table.Clear()
	a.table.SetCell(0, 0, tview.NewTableCell("  loading...").
		SetTextColor(tcell.ColorGray).SetSelectable(false))

	cols, hasColumns := a.userCfg.ColumnsFor(a.resource)

	if hasColumns {
		a.loadJSON(cols)
	} else {
		a.loadRaw()
	}
}

// loadJSON fetches -o json and renders using column definitions.
func (a *App) loadJSON(cols []resource.Column) {
	// Prepend NAMESPACE column when showing all namespaces for namespaced resources.
	allNS := a.client.Namespace == "all" && !isClusterScoped(a.resource)
	if allNS {
		cols = append([]resource.Column{resource.Field("NAMESPACE", "metadata.namespace")}, cols...)
	}

	go func() {
		items, err := a.client.ListJSON(a.resource)
		a.tapp.QueueUpdateDraw(func() {
			if err != nil {
				a.showError(err.Error())
				return
			}

			headers := make([]string, len(cols))
			for i, c := range cols {
				headers[i] = c.Header
			}

			rows := make([][]string, len(items))
			for i, item := range items {
				row := make([]string, len(cols))
				for j, c := range cols {
					row[j] = c.Value(item)
				}
				rows[i] = row
			}

			a.allCols = headers
			a.allRows = rows
			a.renderTable(headers, rows)
		})
	}()
}

// loadRaw fetches raw kubectl text output (fallback for unknown resources).
func (a *App) loadRaw() {
	go func() {
		header, rows, err := a.client.List(a.resource)
		a.tapp.QueueUpdateDraw(func() {
			if err != nil {
				a.showError(err.Error())
				return
			}

			cols := strings.Fields(header)
			parsed := make([][]string, 0, len(rows))
			for _, line := range rows {
				if strings.TrimSpace(line) == "" {
					continue
				}
				parsed = append(parsed, strings.Fields(line))
			}

			a.allCols = cols
			a.allRows = parsed
			a.renderTable(cols, parsed)
		})
	}()
}

func (a *App) showError(msg string) {
	a.table.Clear()
	a.table.SetCell(0, 0, tview.NewTableCell(
		fmt.Sprintf("  [red]error:[-] %s", msg)).
		SetExpansion(1).SetSelectable(false))
}

func (a *App) renderTable(cols []string, rows [][]string) {
	a.table.Clear()

	if len(cols) == 0 {
		a.table.SetCell(0, 0, tview.NewTableCell("  [gray]no resources found[-]").
			SetExpansion(1).SetSelectable(false))
		return
	}

	// Header row
	for c, col := range cols {
		a.table.SetCell(0, c, tview.NewTableCell(" "+col+" ").
			SetTextColor(tcell.ColorYellow).
			SetSelectable(false).
			SetExpansion(1).
			SetAttributes(tcell.AttrBold))
	}

	// Data rows (filtered)
	r := 1
	for _, row := range a.filterRows(rows) {
		for c, cell := range row {
			a.table.SetCell(r, c, tview.NewTableCell(" "+cell+" ").
				SetExpansion(1).
				SetTextColor(statusColor(cell)))
		}
		r++
	}

	if a.table.GetRowCount() > 1 {
		a.table.Select(1, 0)
		a.table.ScrollToBeginning()
	}
}

func (a *App) filterRows(rows [][]string) [][]string {
	if a.filter == "" {
		return rows
	}
	f := strings.ToLower(a.filter)
	var out [][]string
	for _, row := range rows {
		for _, cell := range row {
			if strings.Contains(strings.ToLower(cell), f) {
				out = append(out, row)
				break
			}
		}
	}
	return out
}

// selectedResource extracts (namespace, name) for the selected row.
// Detects all-namespaces layout by checking whether the first header is "NAMESPACE".
func (a *App) selectedResource() (ns, name string) {
	row, _ := a.table.GetSelection()
	if row < 1 || row >= a.table.GetRowCount() {
		return "", ""
	}

	h0 := a.table.GetCell(0, 0)
	if h0 != nil && strings.EqualFold(strings.TrimSpace(h0.Text), "namespace") {
		if c := a.table.GetCell(row, 0); c != nil {
			ns = strings.TrimSpace(c.Text)
		}
		if c := a.table.GetCell(row, 1); c != nil {
			name = strings.TrimSpace(c.Text)
		}
		return
	}

	if c := a.table.GetCell(row, 0); c != nil {
		name = strings.TrimSpace(c.Text)
	}
	return "", name
}

func isClusterScoped(resource string) bool {
	scoped := map[string]bool{
		"nodes":                     true,
		"namespaces":                true,
		"persistentvolumes":         true,
		"clusterroles":              true,
		"clusterrolebindings":       true,
		"storageclasses":            true,
		"ingressclasses":            true,
		"customresourcedefinitions": true,
		"priorityclasses":           true,
	}
	return scoped[resource]
}
