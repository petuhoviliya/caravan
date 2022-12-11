package main

import (
	"fmt"
	"os"
	"time"

	//  "github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const refreshInterval = 1000 * time.Millisecond

var (
	app *tview.Application

	textMap     *tview.TextView
	textTown    *tview.TextView
	textCaravan *tview.TextView
	textLog     *tview.TextView
)

func refresh() {

	tick := time.NewTicker(refreshInterval)

	for {
		select {
		case <-tick.C:
			t := time.Now().Format("15:04:05")

			app.QueueUpdateDraw(func() {
				textMap.SetText(fmt.Sprintf("Map ticker at %s\n", t))
				//fmt.Fprintf(textMap, "Map ticker at %s\n", t)
				fmt.Fprintf(textLog, "Log ticker at %s\n", t)
			})
		}
	}
}

func main() {

	var proc float64 = 0.075
	var days float64 = 365.0

	fmt.Printf("В день: %.5f%, %.4f", proc/days, proc/days*10782.24)

	os.Exit(0)

	app = tview.NewApplication()

	textMap = tview.NewTextView().
		SetScrollable(true).
		SetWrap(true).
		SetWordWrap(true)

	textMap.
		SetBorder(true).
		SetTitleAlign(tview.AlignLeft).
		SetTitle("Map")

	textLog = tview.NewTextView().
		SetScrollable(true).
		SetWrap(true).
		SetWordWrap(true)

	textLog.
		SetBorder(true).
		SetTitleAlign(tview.AlignLeft).
		SetTitle("Log")

	fmt.Fprintf(textMap, "Map start\n")

	/*	boxMap = tview.NewBox().
			SetBorder(true).
			SetTitle("Map").
			SetTitleAlign(tview.AlignLeft)

	  boxTowns := tview.NewBox().
			SetBorder(true).
			SetTitle("Towns").
			SetTitleAlign(tview.AlignLeft)

		boxCaravans := tview.NewBox().
			SetBorder(true).
			SetTitle("Caravans").
			SetTitleAlign(tview.AlignLeft)

		boxLog := tview.NewTextView().
			SetText("Log start\n").
			SetChangedFunc(func() {app.Draw()})

		boxLog.
			SetBorder(true).

			SetTitle("Log").
			SetTitleAlign(tview.AlignLeft)

	  grid := tview.NewGrid().
			SetRows(-1, -1).
			SetColumns(-1, -1 ,-1).
			SetMinSize(15, 20).
			SetBorders(false)

		grid.AddItem(boxMap,    0, 0, 1, 2, 0, 0, false).
			AddItem(boxLog,       0, 2, 2, 1, 0, 0, false).
			AddItem(boxTowns,     1, 0, 1, 1, 0, 0, false).
			AddItem(boxCaravans,  1, 1, 1, 1, 0, 0, false)

		fmt.Printf("%#v\n", boxLog)
	*/

	grid := tview.NewGrid().
		SetRows(-1, -1).
		SetColumns(-1, -1, -1).
		SetMinSize(15, 20).
		SetBorders(false)

	grid.AddItem(textMap, 0, 0, 1, 2, 0, 0, false).
		AddItem(textLog, 0, 2, 2, 1, 0, 0, false)

	go refresh()

	if err := app.SetRoot(grid, true).SetFocus(grid).Run(); err != nil {
		panic(err)
	}

	os.Exit(0)
}
