package ui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

func (a *App) displayNS() string {
	if a.client.Namespace == "all" {
		return "*"
	}
	return a.client.Namespace
}

func (a *App) renderHeader(subtitle string) {
	ctx := a.client.CurrentContext()
	ns := a.displayNS()

	line3 := " [::b]resource:[:-] [#ff5faf]" + a.resource + "[-]"
	if subtitle != "" {
		line3 += "  " + subtitle
	} else if a.filter != "" {
		line3 += fmt.Sprintf("  [gray]filter:[:-] [white]%s[-]", a.filter)
	}

	allNSIndicator := ""
	if ns == "*" {
		allNSIndicator = " [yellow]⚠ all-namespaces[-]"
	}

	a.header.SetText(fmt.Sprintf(
		"\n context: [white]%s[-]\n namespace: [#ff5f87]%s[-]%s\n%s\n",
		ctx, ns, allNSIndicator, line3))
}

func (a *App) updateHeader() {
	a.renderHeader("")
}

func (a *App) updateHint() {
	if a.waitingForDelete {
		a.hint.SetText("  [red]⚠ DELETE CONFIRMATION[:-] [red]y[-] yes  [gray]N[-] cancel")
		return
	}
	if a.view == viewDetail {
		wrapLabel := "[gray]w[-] wrap"
		if a.detailWrap {
			wrapLabel = "[green]✓ w[-] wrap"
		}
		if a.detailSearch != "" && a.detailMatchCnt > 0 {
			a.hint.SetText(fmt.Sprintf(
				"  [gray]│[:-] Esc clear  [gray]│[:-] ↑↓ scroll  [gray]│[:-] n/N jump  [yellow]🔍 %d/%d[-]  %s",
				a.detailMatchIdx+1, a.detailMatchCnt, wrapLabel))
		} else if a.detailSearch != "" {
			a.hint.SetText(fmt.Sprintf(
				`  [gray]│[:-] Esc clear  [red]✗ no matches for "%s"[-]  %s`, a.detailSearch, wrapLabel))
		} else {
			a.hint.SetText(formatHint(
				"Esc/q", "back",
				"↑↓", "scroll",
				"g/G", "top/bottom",
				"y", "yaml",
				"/", "search",
				"w", wrapLabel))
		}
		return
	}

	// Default hints per resource type
	var hints []string
	hints = append(hints, ":", "cmd", "↵", "describe", "l", "logs", "e", "edit", "y", "copy", "/", "filter", "R", "refresh")

	switch a.resource {
	case "_contexts":
		hints = []string{":", "cmd", "↵", "switch", "d", "describe", "y", "copy", "/", "filter", "R", "refresh"}
	case "namespaces":
		hints = []string{":", "cmd", "↵", "select", "d", "describe", "y", "copy", "/", "filter", "R", "refresh"}
	case "cronjobs":
		hints = []string{":", "cmd", "↵/d", "describe", "l", "logs", "e", "edit", "y", "copy", "c", "create job", "/", "filter", "R", "refresh"}
	default:
		hints = append(hints, "a", "all-ns")
	}

	a.hint.SetText(formatHint(hints...))
}

// formatHint joins key-value hint pairs with separators.
func formatHint(pairs ...string) string {
	if len(pairs)%2 != 0 {
		return ""
	}
	var parts []string
	for i := 0; i < len(pairs); i += 2 {
		key := pairs[i]
		value := pairs[i+1]
		parts = append(parts, fmt.Sprintf("[gray]%s[-] %s", key, value))
	}
	return "  " + strings.Join(parts, "  [gray]│[:-] ")
}

// statusColor maps common kubectl status strings to display colors.
func statusColor(s string) tcell.Color {
	switch strings.ToLower(s) {
	case "running", "active", "bound", "true", "ready", "healthy":
		return tcell.ColorGreen
	case "pending", "waiting", "unknown", "unschedulable":
		return tcell.ColorYellow
	case "failed", "error", "false", "evicted", "oomkilled",
		"crashloopbackoff", "imagepullbackoff", "errimagepull", "errcreatecontainerconfigerror":
		return tcell.ColorRed
	case "completed", "succeeded":
		return tcell.ColorLightBlue
	case "terminating":
		return tcell.Color208 // orange
	default:
		return tcell.ColorWhite
	}
}
