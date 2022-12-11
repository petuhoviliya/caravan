package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
	//"github.com/jroimartin/gocui"
	//"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const refreshInterval = 1000 * time.Millisecond

type PositionDef struct {
	X int
	Y int
}

type MapDef struct {
	Width  int
	Height int
}

type CaravanDef struct {
	Name       string
	Status     string
	X          int
	Y          int
	Target     int
	PrevTarget int
}

type TownDef struct {
	Name           string
	Tier           int
	X              int
	Y              int
	WarehouseLimit float64
	Wares          []WareGood
}

type FreeCell struct {
	X int
	Y int
}

type Resources struct {
	Id int
	RequiredPerUnit int
}


type TradingGood struct {
	Id          int
	Tier        int
	Name        string
	PriceMin    float64
	PriceMax    float64
	SellingUnit string
	UnitVolume  float64
	UnitWeight  float64
	Resources 	[]Resources
	Consumables []Resources
}

type WareGood struct {
	Id       int
	Quantity float64
}

var (
	GlobalMap  MapDef
	Goods      []TradingGood
	Towns      []TownDef
	Caravan    CaravanDef
	GlobalStep int

	app *tview.Application

	textMap     *tview.TextView
	textLog     *tview.TextView
	textTown    *tview.TextView
	textCaravan *tview.TextView
)

func PrintMap(Map MapDef, Towns []TownDef, Caravan CaravanDef) string {
	// ╗ ╝ ╚ ╔ ╩ ╦ ╠ ═ ║ ╬ ╣ - borders
	// │ ┤ ┐ └ ┴ ┬ ├ ─ ┼ ┘ ┌ - roads
	// @ - caravan

	var PrintableMap string

	for posX := 0; posX <= Map.Width+1; posX++ {

		for posY := 0; posY <= Map.Height+1; posY++ {

			if posY == 0 {

				if posX == 0 { // left upper corner

					PrintableMap = PrintableMap + "╔"

				} else if posX == (Map.Width + 1) { // left bottom corner

					PrintableMap = PrintableMap + "╚"

				} else {

					PrintableMap = PrintableMap + "║" // left border
				}

			} else if posY == (Map.Height + 1) {

				if posX == 0 { // right upper corner

					PrintableMap = PrintableMap + "╗\n"

				} else if posX == (Map.Width + 1) { // right  bottom corner

					PrintableMap = PrintableMap + "╝\n"

				} else {

					PrintableMap = PrintableMap + "║\n" // right border
				}

			} else {

				if posX == (Map.Width+1) || posX == 0 { // up and bottom border

					PrintableMap = PrintableMap + "═"

				} else { // fill space inside map borders and place towns

					var mapObject string = " "

					for _, town := range Towns {

						if town.X == posX && town.Y == posY {
							mapObject = fmt.Sprintf("[%s]%s[%s]", "red", town.Name[0:1], "white")
						}
					}

					if Caravan.X == posX && Caravan.Y == posY {
						mapObject = "@"
					}

					PrintableMap = PrintableMap + mapObject
				}
			}
		}
	}
	return PrintableMap
}

func MoveToPoint(Caravan *CaravanDef, DestX int, DestY int) {

	modX := Caravan.X - DestX
	modY := Caravan.Y - DestY

	if modX < 0 {
		Caravan.X++
	} else if modX > 0 {
		Caravan.X--
	}

	if modY < 0 {
		Caravan.Y++
	} else if modY > 0 {
		Caravan.Y--
	}
}

func RndRange(Min int, Max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(Max-Min+1) + Min
}

func Rnd(Max int) int {
	return RndRange(1, Max)
}

func GenerateRandomPosition(MaxX int, MaxY int) (X int, Y int) {

	X = Rnd(MaxX)
	Y = Rnd(MaxY)

	return X, Y
}

func PutTownsOnMap(Map MapDef, BitMap *[][]byte, TownCount int, MinDistance int) []TownDef {

	var Towns []TownDef
	var Wares []WareGood

	if TownCount > 26 {
		TownCount = 26
	}

	alphabet := [26]string{"Alpha", "Bravo", "Charlie", "Delta", "Echo", "Foxtrot", "Golf", "Hotel", "India", "Juliet", "Kilo", "Lima", "Mike", "November", "Oscar", "Papa", "Quebec", "Romeo", "Sierra", "Tango", "Uniform", "Victor", "Whiskey", "X-ray", "Yankee", "Zulu"}

	for i := 0; i < TownCount-1; i++ {

		FreeCells := GetFreeCellsOnMap((*BitMap))

		if len(FreeCells) == 0 {
			break
		}

		//fmt.Printf("Free: %d\n", len(FreeCells))

		NextFreeCell := RndRange(0, len(FreeCells)-1)

		PutTownOnBitMap(Map, &(*BitMap), FreeCells[NextFreeCell].X, FreeCells[NextFreeCell].Y, MinDistance)

		Wares = nil

		for i, Good := range Goods {
			if i == 0 {
				continue
			}
			Wares = append(Wares, WareGood{Id: Good.Id, Quantity: float64(Rnd(500))})

		}

		Towns = append(Towns, TownDef{Name: alphabet[i], X: FreeCells[NextFreeCell].X, Y: FreeCells[NextFreeCell].Y, WarehouseLimit: 500.0, Wares: Wares})
	}
	return Towns
}

func PutTownOnBitMap(Map MapDef, BitMap *[][]byte, TownX int, TownY int, Distance int) bool {

	for i := -Distance; i <= Distance; i++ {
		for j := -Distance; j <= Distance; j++ {

			A := math.Abs(float64(0 - i))
			B := math.Abs(float64(0 - j))
			C := int(math.Sqrt(math.Pow(A, 2) + math.Pow(B, 2)))

			//fmt.Printf("A: %.0f, B: %.0f, C: %d \n", A, B, C)

			tX := TownX + i
			tY := TownY + j

			if C <= Distance {
				if tX > (Map.Width - 1) {
					tX = Map.Width - 1
				}
				if tY > (Map.Height - 1) {
					tY = Map.Height - 1
				}
				if tX < 0 {
					tX = 0
				}
				if tY < 0 {
					tY = 0
				}

				//fmt.Printf("tX: %d, tY: %d\n", tX, tY)
				(*BitMap)[tX][tY] = 1
			}
		}
	}

	return false
}

func GetFreeCellsOnMap(BitMap [][]byte) []FreeCell {

	var FreeCells []FreeCell

	for x, i := range BitMap {
		for y := range i {
			if BitMap[x][y] == 0 {
				FreeCells = append(FreeCells, FreeCell{X: x + 1, Y: y + 1})
			}
		}
	}
	return FreeCells
}

func MakeBitMap(Width int, Height int) [][]byte {

	bitMap := make([][]byte, Width)

	for i := range bitMap {
		bitMap[i] = make([]byte, Height)
	}

	return bitMap
}

func FindBestNextPoint(StartX int, StartY int, DestX int, DestY int) (X int, Y int) {

	X = 0
	Y = 0
	Cost := math.Inf(1)

	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {

			tX := StartX + i
			tY := StartY + j

			A := math.Abs(float64(DestX - tX))
			B := math.Abs(float64(DestY - tY))
			C := math.Sqrt(math.Pow(A, 2) + math.Pow(B, 2))

			if Cost != 0 {
				if C < Cost {
					Cost = C
					X = StartX + i
					Y = StartY + j
				}
			} else {
				Cost = 0
				X = DestX
				Y = DestY
			}
		}
	}
	return
}

func FindPath(StartX int, StartY int, DestX int, DestY int) {

	fmt.Printf("Start %d:%d\n", StartX, StartY)
	fmt.Printf("Destination %d:%d\n", DestX, DestY)

	bitMap := MakeBitMap(11, 11)

	type Node struct {
		X    int
		Y    int
		Cost float64
	}

	var (
		Path      []Node
		Last      Node
		TotalCost float64
	)

	Path = append(Path, Node{X: StartX, Y: StartY, Cost: 0.0})
	Last.X = StartX
	Last.Y = StartY
	Last.Cost = 0.0

	for {
		bitMap[Last.Y][Last.X] = 1
		Last.X, Last.Y = FindBestNextPoint(Last.X, Last.Y, DestX, DestY)
		TotalCost += Last.Cost
		Path = append(Path, Last)
		if Last.X == DestX && Last.Y == DestY {
			break
		}
	}

	fmt.Printf("Path: %v\nTotal Cost: %.4f", Path, TotalCost)

	for x, i := range bitMap {
		for y := range i {
			fmt.Printf("%v", bitMap[x][y])
		}
		fmt.Printf("\n")
	}

	fmt.Printf("%#v\n", bitMap)
	fmt.Printf("%d %d\n", len(bitMap), cap(bitMap))
	return
}

func refresh() {
	tick := time.NewTicker(refreshInterval)

	for {
		select {
		case <-tick.C:
			t := time.Now().Format("15:04:05")

			GlobalStep++

			app.QueueUpdateDraw(func() {
				textMap.SetText(PrintMap(GlobalMap, Towns, Caravan))
				fmt.Fprintf(textMap, "Global step: %d\n", GlobalStep)
				fmt.Fprintf(textMap, "Map ticker at %s\n", t)
				fmt.Fprintf(textMap, "Caravan destination: %s\n", Towns[Caravan.Target].Name)
				//				fmt.Fprintf(textLog, "Log ticker at %s\n", t)

				if Caravan.X == Towns[Caravan.Target].X && Caravan.Y == Towns[Caravan.Target].Y {

					fmt.Fprintf(textLog, "Arrived to destination \"%s\" at step %d\n", Towns[Caravan.Target].Name, GlobalStep)

					Caravan.PrevTarget = Caravan.Target

					for {
						Caravan.Target = RndRange(0, len(Towns)-1)
						if Caravan.Target != Caravan.PrevTarget {
							break
						}
					}

					//fmt.Fprintf(textLog, "destination: %s\n", Towns[Caravan.Target].Name)
				}

				tX, tY := FindBestNextPoint(Caravan.X, Caravan.Y, Towns[Caravan.Target].X, Towns[Caravan.Target].Y)

				Caravan.X = tX
				Caravan.Y = tY

			})
		}
	}
}

func init() {
	log.Println("Init")

	/*
	  Id int
	  Tier int
	  Name string
	  PriceMin int
	  PriceMax int
	  SellingUnit string
	  UnitVolume float64
	  UnitWeight float64
	*/

	/*
	  Расчет цены продажи/покупки
	  CurrentPrice := PriceMin + (PriceMax - PriceMix)*(1 - WareQuantity/WarehouseLimit)
	  С округлением вверх

	*/

	Goods = []TradingGood{
		{Id: 0, Tier: 0, Name: "Шаблон", PriceMin: 1, PriceMax: 1, SellingUnit: "ед.", UnitVolume: 1.0, UnitWeight: 1.0},
		{Id: 1, Tier: 1, Name: "Зерно", PriceMin: 2, PriceMax: 10, SellingUnit: "мешок", UnitVolume: 0.036, UnitWeight: 0.050},
		{Id: 2, Tier: 1, Name: "Дерево", PriceMin: 5, PriceMax: 20, SellingUnit: "кубометр", UnitVolume: 1.0, UnitWeight: 0.640},
		{Id: 3, Tier: 1, Name: "Камень", PriceMin: 4, PriceMax: 18, SellingUnit: "кубометр", UnitVolume: 1.0, UnitWeight: 1.7},
		{Id: 4, Tier: 1, Name: "Руда", PriceMin: 9, PriceMax: 30, SellingUnit: "тонна", UnitVolume: 0.5, UnitWeight: 1.0},
		
		{Id: 5, Tier: 2, Name: "Мука", PriceMin: 40, PriceMax: 75, SellingUnit: "мешок", UnitVolume: 0.036, UnitWeight: 0.050, 
			Resources: []Resources{
				Resources{Id: 1, RequiredPerUnit: 8},
			},
		},
		{Id: 6, Tier: 2, Name: "Доски", PriceMin: 50, PriceMax: 100, SellingUnit: "кубометр", UnitVolume: 0.0, UnitWeight: 0.0, 
			Resources: []Resources{
				Resources{Id: 2, RequiredPerUnit: 8},
			},
		},
		{Id: 7, Tier: 2, Name: "Каменная заготовка", PriceMin: 1, PriceMax: 100, SellingUnit: "партия", UnitVolume: 0.0, UnitWeight: 0.0, 
			Resources: []Resources{
				Resources{Id: 3, RequiredPerUnit: 8},
			},
		},
		{Id: 8, Tier: 2, Name: "Металлические инструменты", PriceMin: 1, PriceMax: 100, SellingUnit: "партия", UnitVolume: 0.0, UnitWeight: 0.0, 
			Resources: []Resources{
				Resources{Id: 4, RequiredPerUnit: 8},
			},
		},
		{Id: 9, Tier: 3, Name: "Деревянная мебель", PriceMin: 100, PriceMax: 200, SellingUnit: "", UnitVolume: 0.0, UnitWeight: 0.0, 
			Resources: []Resources{
				Resources{Id: 2, RequiredPerUnit: 4},
				Resources{Id: 6, RequiredPerUnit: 8},
			},
		},
	}
}

func main() {

	var (
		Size          int
		MinDistance   int
		MaxTownsCount int
	)

	for _, Good := range Goods {
		fmt.Printf("%s[%d] цена: %.1f/%.1f\n", Good.Name, Good.Tier, Good.PriceMin, Good.PriceMax)
		for _, Resource := range Good.Resources {
			fmt.Printf("  %s[%d]: %d\n", Goods[Resource.Id].Name, Goods[Resource.Id].Tier, Resource.RequiredPerUnit)
		}
	}

	os.Exit(0)

	app = tview.NewApplication()

	textMap = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(false).
		SetWordWrap(false).
		SetText("Loading...")

	textMap.
		SetBorder(true).
		SetTitleAlign(tview.AlignLeft).
		SetTitle("Map")

	textLog = tview.NewTextView().
		SetScrollable(true).
		SetWrap(true).
		SetWordWrap(true).
		SetMaxLines(100).
		SetText("Loading...")

	textLog.
		SetBorder(true).
		SetTitleAlign(tview.AlignLeft).
		SetTitle("Log")

	textTown = tview.NewTextView().
		SetScrollable(true).
		SetWrap(true).
		SetWordWrap(true).
		SetText("Loading...")

	textTown.
		SetBorder(true).
		SetTitleAlign(tview.AlignLeft).
		SetTitle("Towns")

	textCaravan = tview.NewTextView().
		SetScrollable(true).
		SetWrap(true).
		SetWordWrap(true).
		SetText("Loading...\n")

	textCaravan.
		SetBorder(true).
		SetTitleAlign(tview.AlignLeft).
		SetTitle("Caravan")

	grid := tview.NewGrid().
		SetRows(-1, -1).
		SetColumns(-2, -2, -3).
		SetMinSize(15, 20).
		SetBorders(false)

	grid.AddItem(textMap, 0, 0, 1, 2, 0, 0, false).
		AddItem(textLog, 0, 2, 2, 1, 0, 0, false).
		AddItem(textTown, 1, 0, 1, 1, 0, 0, false).
		AddItem(textCaravan, 1, 1, 1, 1, 0, 0, false)

	//os.Exit(0)

	//SetupCloseHandler()

	Size = 15
	MinDistance = 5
	MaxTownsCount = 26

	GlobalMap = MapDef{Width: Size, Height: Size * 4}

	bitMap := MakeBitMap(GlobalMap.Width, GlobalMap.Height)

	Towns = PutTownsOnMap(GlobalMap, &bitMap, MaxTownsCount, MinDistance)

	Caravan = CaravanDef{Name: "Caravan", Status: "At point", X: 1, Y: 1}

	Caravan.Target = RndRange(0, len(Towns)-1)
	Caravan.PrevTarget = -1

	go refresh()

	if err := app.SetRoot(grid, true).SetFocus(grid).Run(); err != nil {
		panic(err)
	}

	os.Exit(0)

	/*	for x, i := range bitMap {
		for y := range i {
			fmt.Printf("%v", bitMap[x][y])
		}
		fmt.Printf("\n")
	}*/

	//fmt.Printf("%v\n", bitMap)
	//tX, tY := FindBestNextPoint(1,1,32,11)

	//fmt.Printf("%d:%d\n" , tX, tY)

	//os.Exit(0)

	/*	step := 0
		prev := -1

		fmt.Sprintf("Start at %d:%d\n", Caravan.X, Caravan.Y)
		fmt.Printf("First destination \"%s\" at %d:%d\n", Towns[target].Name, Towns[target].X, Towns[target].Y)

		for {

			fmt.Print("\033[H\033[2J")

			pMap := PrintMap(GlobalMap, Towns, Caravan)
			fmt.Printf("%s", pMap)

			for _, town := range Towns {
				fmt.Printf("%+v\n", town)
				for _, ware := range town.Wares {
					Price := Goods[ware.Id].PriceMin + (Goods[ware.Id].PriceMax-Goods[ware.Id].PriceMin)*(1.0-ware.Quantity/town.WarehouseLimit)
					Price = math.Round(Price)
					fmt.Printf("  %s: %.0f/%.0f Цена: %.0f\n", Goods[ware.Id].Name, ware.Quantity, town.WarehouseLimit, Price)
				}
				fmt.Printf("\n")
			}
			fmt.Printf("Count: %d\n", len(Towns[Caravan.Target].Nameu)

			log.Printf("Destination \"%s\" at %d:%d\n", Towns[target].Name, Towns[target].X, Towns[target].Y)
			log.Printf("Step %d, pos %d:%d\n", step, Caravan.X, Caravan.Y)

			if Caravan.X == Towns[target].X && Caravan.Y == Towns[target].Y {

				log.Printf("Arrived at destination %d:%d\n", Caravan.X, Caravan.Y)

				prev = target

				for {
					target = RndRange(0, len(Towns)-1)
					if target != prev {
						break
					}
				}

			}
			tX, tY := FindBestNextPoint(Caravan.X, Caravan.Y, Towns[target].X, Towns[target].Y)
			Caravan.X = tX
			Caravan.Y = tY

			//MoveToPoint(&caravan, towns[target].X, towns[target].Y)

			step++

			time.Sleep(1000 * time.Millisecond)
		}*/

}

func SetupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		s := <-c
		log.Printf("Quit by \"%s\"", s)
		os.Exit(0)
	}()
}

/**  newPrimitive := func(text string) tview.Primitive {
	return tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetText(text)
} */

/*	viewMap := newPrimitive("Map")
	viewTowns := newPrimitive("Towns")
//	viewCaravans := newPrimitive("Caravans")
	viewActionsLog := newPrimitive("Actions log")*/

/*	boxMap := tview.NewBox().
	SetBorder(true).
	SetTitle("Map").
	SetTitleAlign(tview.AlignLeft)*/

/*
  boxTowns := tview.NewBox().
		SetBorder(true).
		SetTitle("Towns").
		SetTitleAlign(tview.AlignLeft)

	boxCaravans := tview.NewBox().
		SetBorder(true).
		SetTitle("Caravans").
		SetTitleAlign(tview.AlignLeft)

	textLog := tview.NewTextView()

  textMap := tview.NewTextView().
    SetWrap(false).
    SetText(PrintMap(GlobalMap, towns))

  grid := tview.NewGrid().
		SetRows(0, 0).
		SetColumns(-2, -1).
		SetMinSize(15, 20).
		SetBorders(true)

	grid.AddItem(textMap, 0, 0, 1, 1, 0, 0, false).
		AddItem(boxTowns, 0, 1, 1, 1, 0, 0, false).
		AddItem(textLog, 1, 0, 1, 1, 0, 0, false).
		AddItem(boxCaravans, 1, 1, 1, 1, 0, 0, false)

	if err := app.SetRoot(grid, true).SetFocus(grid).Run(); err != nil {
		panic(err)
	}*/

// Status:
//  At point
//  Moving
