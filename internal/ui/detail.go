package ui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rivo/tview"
)

func (a *App) openDescribe() {
	ns, name := a.selectedResource()
	if name == "" {
		return
	}
	a.switchToDetail(fmt.Sprintf("[yellow]describe[-]  %s", name))
	go func() {
		out, _ := a.client.Describe(a.resource, ns, name)
		a.tapp.QueueUpdateDraw(func() {
			a.detailRaw = out
			a.detailIsLogs = false
			a.renderDetailContent()
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
			a.detailRaw = out
			a.detailIsLogs = false
			a.renderDetailContent()
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
			a.detailRaw = out
			a.detailIsLogs = true
			a.renderDetailContent()
			a.detail.ScrollToEnd()
		})
	}()
}

func (a *App) openEdit() {
	ns, name := a.selectedResource()
	if name == "" {
		return
	}
	a.tapp.Suspend(func() {
		_ = a.client.Edit(a.resource, ns, name)
	})
}

func (a *App) switchToDetail(subtitle string) {
	a.view = viewDetail
	a.detailRaw = ""
	a.detailSearch = ""
	a.detailWrap = false
	a.detailMatchCnt = 0
	a.detailMatchIdx = 0
	a.detail.SetRegions(false)
	a.detail.Highlight()
	a.detail.SetDynamicColors(false)
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

// toggleWrap flips line-wrapping in the detail view (useful for long log lines).
func (a *App) toggleWrap() {
	a.detailWrap = !a.detailWrap
	a.detail.SetWrap(a.detailWrap)
	a.updateHint()
}

// applyDetailSearch updates the active search term and re-renders.
// Safe to call from the main goroutine (e.g. input ChangedFunc or key handlers).
func (a *App) applyDetailSearch(search string) {
	a.detailSearch = search
	a.renderDetailContent()
}

// nextMatch advances (dir=+1) or retreats (dir=-1) to the next search match.
func (a *App) nextMatch(dir int) {
	if a.detailMatchCnt == 0 {
		return
	}
	a.detailMatchIdx = (a.detailMatchIdx + dir + a.detailMatchCnt) % a.detailMatchCnt
	a.detail.Highlight(fmt.Sprintf("m%d", a.detailMatchIdx)).ScrollToHighlight()
	a.updateHint()
}

// renderDetailContent re-renders the detail view from raw text with the current search term.
func (a *App) renderDetailContent() {
	if a.detailSearch == "" {
		a.detail.SetRegions(false)
		a.detail.Highlight()
		if a.detailIsLogs {
			a.detail.SetDynamicColors(true)
			a.detail.SetText(formatLogs(a.detailRaw))
		} else {
			a.detail.SetDynamicColors(false)
			a.detail.SetText(a.detailRaw)
		}
		a.detailMatchCnt = 0
		a.detailMatchIdx = 0
		a.updateHint()
		return
	}

	a.detail.SetDynamicColors(true)
	a.detail.SetRegions(true)

	var text string
	var count int
	if a.detailIsLogs {
		text, count = formatLogsWithSearch(a.detailRaw, a.detailSearch)
	} else {
		text, count = injectHighlights(a.detailRaw, a.detailSearch)
	}

	a.detailMatchCnt = count
	a.detail.SetText(text)

	if count > 0 {
		a.detailMatchIdx = 0
		a.detail.Highlight("m0").ScrollToHighlight()
	} else {
		a.detailMatchIdx = 0
		a.detail.Highlight()
	}
	a.updateHint()
}

// injectHighlights escapes text for tview dynamic colors and wraps every
// case-insensitive match in a highlight color tag + region tag.
func injectHighlights(text, search string) (string, int) {
	if search == "" {
		return tview.Escape(text), 0
	}
	lText := strings.ToLower(text)
	lSearch := strings.ToLower(search)
	var buf strings.Builder
	pos, count := 0, 0
	for {
		idx := strings.Index(lText[pos:], lSearch)
		if idx == -1 {
			buf.WriteString(tview.Escape(text[pos:]))
			break
		}
		abs := pos + idx
		buf.WriteString(tview.Escape(text[pos:abs]))
		match := tview.Escape(text[abs : abs+len(search)])
		buf.WriteString(fmt.Sprintf(`["m%d"][black:yellow:b]%s[-:-:-][""]`, count, match))
		count++
		pos = abs + len(search)
	}
	return buf.String(), count
}

// formatLogsWithSearch is like formatLogs but also injects search highlights.
func formatLogsWithSearch(raw, search string) (string, int) {
	lines := strings.Split(raw, "\n")
	lSearch := strings.ToLower(search)
	var buf strings.Builder
	total, matchIdx := 0, 0
	for i, line := range lines {
		levelColor := logLevelColor(detectLogLevel(line))
		lLine := strings.ToLower(line)
		pos := 0
		var lb strings.Builder
		for {
			idx := strings.Index(lLine[pos:], lSearch)
			if idx == -1 {
				break
			}
			abs := pos + idx
			seg := tview.Escape(line[pos:abs])
			if levelColor != "" && seg != "" {
				lb.WriteString(fmt.Sprintf("[%s]%s[-]", levelColor, seg))
			} else {
				lb.WriteString(seg)
			}
			matchStr := tview.Escape(line[abs : abs+len(search)])
			lb.WriteString(fmt.Sprintf(`["m%d"][black:yellow:b]%s[-:-:-][""]`, matchIdx, matchStr))
			matchIdx++
			total++
			pos = abs + len(search)
		}
		trailing := tview.Escape(line[pos:])
		if levelColor != "" && trailing != "" {
			lb.WriteString(fmt.Sprintf("[%s]%s[-]", levelColor, trailing))
		} else {
			lb.WriteString(trailing)
		}
		buf.WriteString(lb.String())
		if i < len(lines)-1 {
			buf.WriteByte('\n')
		}
	}
	return buf.String(), total
}

func formatLogs(raw string) string {
	lines := strings.Split(raw, "\n")
	var buf strings.Builder
	for i, line := range lines {
		buf.WriteString(coloredLogLine(line))
		if i < len(lines)-1 {
			buf.WriteByte('\n')
		}
	}
	return buf.String()
}

func coloredLogLine(line string) string {
	color := logLevelColor(detectLogLevel(line))
	if color == "" {
		return tview.Escape(line)
	}
	return fmt.Sprintf("[%s]%s[-]", color, tview.Escape(line))
}

func detectLogLevel(line string) string {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "{") {
		var m map[string]any
		if json.Unmarshal([]byte(trimmed), &m) == nil {
			for _, key := range []string{"level", "severity", "lvl", "Level", "Severity"} {
				if v, ok := m[key]; ok {
					if s, ok := v.(string); ok {
						return strings.ToUpper(s)
					}
				}
			}
		}
	}
	upper := strings.ToUpper(line)
	for _, kw := range []string{"FATAL", "CRITICAL", "ERROR", "WARNING", "WARN", "INFO", "DEBUG", "TRACE"} {
		if strings.Contains(upper, kw) {
			return kw
		}
	}
	return ""
}

func logLevelColor(level string) string {
	switch level {
	case "FATAL", "CRITICAL", "ERROR":
		return "red"
	case "WARN", "WARNING":
		return "yellow"
	case "INFO":
		return "green"
	case "DEBUG":
		return "gray"
	case "TRACE":
		return "#808080"
	default:
		return ""
	}
}
