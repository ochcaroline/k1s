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
	line3 := " resource: [pink]" + a.resource + "[-]"
	if subtitle != "" {
		line3 += "  " + subtitle
	} else if a.filter != "" {
		line3 += fmt.Sprintf("  [gray]/[-][white]%s[-]", a.filter)
	}
	a.header.SetText(fmt.Sprintf(
		" [::b]k1s[-]\n context: [cyan]%s[-]\n namespace: [pink]%s[-]\n%s\n",
		ctx, ns, line3))
}

func (a *App) updateHeader() {
	a.renderHeader("")
}

func (a *App) updateHint() {
	if a.view == viewDetail {
		a.hint.SetText("  [gray]Esc/q[-] back  [gray]↑↓[-] scroll  [gray]g[-] top  [gray]G[-] bottom")
		return
	}
	if a.resource == "namespaces" {
		a.hint.SetText("  [gray]:[-] cmd  [gray]↵[-] select ns→pods  [gray]d[-] describe  [gray]y[-] yaml  [gray]c[-] copy  [gray]/[-] filter  [gray]R[-] refresh")
		return
	}
	a.hint.SetText("  [gray]:[-] cmd  [gray]↵/d[-] describe  [gray]y[-] yaml  [gray]l[-] logs  [gray]c[-] copy  [gray]/[-] filter  [gray]a[-] all-ns  [gray]R[-] refresh")
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
