package ui

import (
	"k1s/internal/kubectl"
	"k1s/internal/resource"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type viewMode int
type inputMode int

const (
	viewList   viewMode = iota
	viewDetail          // describe / yaml / logs
)

const (
	inputCommand inputMode = iota
	inputFilter
	inputDetailSearch
)

// App holds the entire TUI state.
type App struct {
	tapp     *tview.Application
	client   *kubectl.Client
	userCfg  resource.UserConfig
	resource string

	// Loaded data (unfiltered, normalized)
	allCols []string   // column headers
	allRows [][]string // cell values per row

	filter        string
	prevNamespace string // namespace saved before switching to all-namespaces

	// State
	view  viewMode
	imode inputMode

	// Detail view search state
	detailRaw      string
	detailIsLogs   bool
	detailWrap     bool
	detailSearch   string
	detailMatchCnt int
	detailMatchIdx int

	// Deletion confirmation
	waitingForDelete bool
	pendingDKey      bool

	// UI widgets
	root     *tview.Flex
	header   *tview.TextView
	table    *tview.Table
	detail   *tview.TextView
	cmdInput *tview.InputField
	hint     *tview.TextView
	pages    *tview.Pages
}

// New creates and initialises the TUI application.
func New(namespace, context string) (*App, error) {
	cfg, err := resource.LoadConfig()
	if err != nil {
		// non-fatal: bad config → use defaults
		cfg = nil
	}
	a := &App{
		tapp:     tview.NewApplication(),
		client:   &kubectl.Client{Namespace: namespace, Context: context},
		userCfg:  cfg,
		resource: "pods",
	}
	a.buildUI()
	return a, nil
}

func (a *App) buildUI() {
	// Use terminal's own background everywhere.
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault

	a.header = tview.NewTextView().
		SetDynamicColors(true)
	a.header.SetBackgroundColor(tcell.ColorDefault)

	a.table = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0).
		SetSelectedStyle(tcell.StyleDefault.
			Background(tcell.Color162).
			Foreground(tcell.ColorWhite))
	a.table.SetBackgroundColor(tcell.ColorDefault)

	a.detail = tview.NewTextView().
		SetScrollable(true).
		SetWrap(false).
		SetDynamicColors(false)
	a.detail.SetBackgroundColor(tcell.ColorDefault)

	a.cmdInput = tview.NewInputField().
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldTextColor(tcell.ColorWhite).
		SetLabelColor(tcell.Color204)
	a.cmdInput.SetBackgroundColor(tcell.ColorDefault)

	a.hint = tview.NewTextView().
		SetDynamicColors(true)
	a.hint.SetBackgroundColor(tcell.ColorDefault)

	sep := tview.NewBox().
		SetBackgroundColor(tcell.ColorDefault).
		SetDrawFunc(func(screen tcell.Screen, x, y, width, _ int) (int, int, int, int) {
			for col := x; col < x+width; col++ {
				screen.SetContent(col, y, '─', nil,
					tcell.StyleDefault.Foreground(tcell.ColorGray))
			}
			return x, y, width, 1
		})

	a.pages = tview.NewPages().
		AddPage("list", a.table, true, true).
		AddPage("detail", a.detail, true, false)

	// cmdInput starts hidden (fixedSize=0, proportion=0 = no space allocated)
	a.root = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.header, 4, 0, false).
		AddItem(sep, 1, 0, false).
		AddItem(a.pages, 0, 1, true).
		AddItem(a.cmdInput, 0, 0, false).
		AddItem(a.hint, 1, 0, false)

	a.setupKeys()
	a.tapp.SetRoot(a.root, true).EnableMouse(false)
}

// Run starts the TUI event loop.
func (a *App) Run() error {
	if a.client.Namespace == "" {
		a.client.Namespace = a.client.CurrentNamespace()
	}
	a.updateHeader()
	a.updateHint()
	a.loadResource()
	return a.tapp.Run()
}
