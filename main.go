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

const refreshInterval = 750 * time.Millisecond

type PositionDef struct {
	X int
	Y int
}

type MapDef struct {
	Width  int
	Height int
}


type Cargo struct {
	TradingGoodId int
	Quantity float64
	BuyPrice float64
}

type TradeConfig struct {
	
	// Максимальная цена покупки товара
	// в процентах 
	// TradingGood.PriceMin - 0%  
	// TradingGood.PriceMax - 100%
	BuyMaxPrice int
	
	// Всегда покупать максимально возможное количество до полной емкости
	BuyFullCapacity bool

	// Максимальное кол-во покупки товара
	// в процентах от CaravanTemplate.CapacityMax
	// Значение игнорируется, если BuyFullCapacity == true
	BuyMaxAmount int


	// Всегда продавать с прибылью
	// Цена продажи не может быть ниже цены покупки
	SellWithProfit bool

	// Минимальная цена продажи товара
	// в процентах 
	// TradingGood.PriceMin - 0%  
	// TradingGood.PriceMax - 100%
	SellMinPrice int

}

type CaravanTemplate struct {
	Name       string
	Status     string
	X          int
	Y          int
	Target     int
	PrevTarget int
	Capacity float64
	CapacityMax float64
	Cargo []Cargo
	TradeConfig TradeConfig
}

type TownTemplate struct {
	Name           string
	Tier           int
	X              int
	Y              int
	WarehouseLimit float64
	Wares          []WareGood
	Visited int
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
	Towns      []TownTemplate
	Caravan    CaravanTemplate
	GlobalStep int
	TotalVisited int

	app *tview.Application

	textMap     *tview.TextView
	textLog     *tview.TextView
	textTown    *tview.TextView
	textCaravan *tview.TextView
)

func PrintMap(Map MapDef, Towns []TownTemplate, Caravan CaravanTemplate) string {
	// ╗ ╝ ╚ ╔ ╩ ╦ ╠ ═ ║ ╬ ╣ - borders
	// │ ┤ ┐ └ ┴ ┬ ├ ─ ┼ ┘ ┌ - roads
	// @ - caravan

	var PrintableMap string
	var ColorTag string

	for posX := 0; posX <= Map.Height + 1; posX++ {
		for posY := 0; posY <= Map.Width + 1; posY++ {

			//fmt.Printf("X:%d, Y:%d\n", posX, posY)

			if posY == 0 {
				if posX == 0 { 												// left upper corner
					PrintableMap = PrintableMap + "╔"
				} else if posX == (Map.Height + 1) { 	// left bottom corner
					PrintableMap = PrintableMap + "╚"
				} else {
					PrintableMap = PrintableMap + "║" 	// left border
				}
			
			} else if posY == (Map.Width + 1) {
				if posX == 0 { 												// right upper corner
					PrintableMap = PrintableMap + "╗\n"
				} else if posX == (Map.Height + 1) { 	// right  bottom corner
					PrintableMap = PrintableMap + "╝\n"
				} else {
					PrintableMap = PrintableMap + "║\n" // right border
				}
			
			} else {
				if posX == (Map.Height + 1) || posX == 0 { // up and bottom border
					PrintableMap = PrintableMap + "═"
				} else { // fill space inside map borders and place towns
					
					var mapObject string = " "
					for _, town := range Towns {
						if town.X == posX && town.Y == posY {
							switch Tier := town.Tier; Tier {
								case 1:
									ColorTag = "red"
								case 2:
									ColorTag = "orange"
								case 3:
									ColorTag = "green"
								default:
									ColorTag = "white"
							}
							mapObject = fmt.Sprintf("[%s]%s[%s]", ColorTag, town.Name[0:1], "white")
							//mapObject = fmt.Sprintf("%s", town.Name[0:1])
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

func MoveToPoint(Caravan *CaravanTemplate, DestX int, DestY int) {

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

func PutTownsOnMap(Map MapDef, BitMap *[][]byte, TownCount int, MinDistance int) []TownTemplate {

	var Towns []TownTemplate
	var Wares []WareGood
	var MaxTier2 int
	var MaxTier3 int
	var Tier2 int = 1
	var Tier3 int = 1

	if TownCount > 26 {
		TownCount = 26
	}

	MaxTier2 = RndRange(1,2)
	MaxTier3 = RndRange(1,1)

	alphabet := [26]string{"Alpha", "Bravo", "Charlie", "Delta", "Echo", "Foxtrot", "Golf", "Hotel", "India", "Juliet", "Kilo", "Lima", "Mike", "November", "Oscar", "Papa", "Quebec", "Romeo", "Sierra", "Tango", "Uniform", "Victor", "Whiskey", "X-ray", "Yankee", "Zulu"}

	for i := 0; i <= TownCount - 1; i++ {

		FreeCells := GetFreeCellsOnMap((*BitMap))

		if len(FreeCells) == 0 {
			break
		}

		//fmt.Printf("Free: %d\n", len(FreeCells))

		NextFreeCell := RndRange(0, len(FreeCells)-1)

		PutTownOnBitMap(Map, &(*BitMap), FreeCells[NextFreeCell].X, FreeCells[NextFreeCell].Y, MinDistance)

		Wares = nil

		for j, Good := range Goods {
			if j == 0 {
				continue
			}
			Wares = append(Wares, WareGood{Id: Good.Id, Quantity: float64(Rnd(500))})

		}

		Towns = append(Towns, TownTemplate{Name: alphabet[i], Tier: 1, X: FreeCells[NextFreeCell].X, Y: FreeCells[NextFreeCell].Y, WarehouseLimit: 500.0, Wares: Wares})

	}

	/*for _, t := range Towns {
		fmt.Printf("%s %d\n", t.Name, t.Tier)
	}*/

	for {
		if Tier2 > MaxTier2 { break }
		t := Rnd(len(Towns) - 1)
		if Towns[t].Tier == 1 {
			Tier2++
			Towns[t].Tier = 2
		}
	}

	for {
		if Tier3 > MaxTier3 { break }
		t := Rnd(len(Towns) - 1)
		if Towns[t].Tier == 1 {
			Tier3++
			Towns[t].Tier = 3
		}
	}


	return Towns
}

func PutTownOnBitMap(Map MapDef, BitMap *[][]byte, TownX int, TownY int, Distance int) {

	for i := -Distance; i <= Distance; i++ {
		for j := -Distance; j <= Distance; j++ {

			A := math.Abs(float64(0 - i))
			B := math.Abs(float64(0 - j))
			C := int(math.Sqrt(math.Pow(A, 2) + math.Pow(B, 2)))

			tX := TownX + i
			tY := TownY + j

			if C <= Distance {
				if tY > (Map.Width - 1) { tY = Map.Width - 1 }
				if tX > (Map.Height - 1) { tX = Map.Height - 1 }
				if tX < 0 { tX = 0 }
				if tY < 0 { tY = 0 }

				(*BitMap)[tX][tY] = 1
			}

		} // for j
	} // for i

}

func GetFreeCellsOnMap(BitMap [][]byte) []FreeCell {

	var FreeCells []FreeCell

/*for x, i := range BitMap {
		for y := range i {
			fmt.Printf("%v", BitMap[x][y])
		}
		fmt.Printf("\n")
	}*/


	for x, i := range BitMap {
		for y := range i {
			if BitMap[x][y] == 0 {
				FreeCells = append(FreeCells, FreeCell{X: x + 1, Y: y + 1})
			}
		} // for y
	} //for x

	//fmt.Printf("Free: %d\n\n", len(FreeCells))
	return FreeCells
}

func MakeBitMap(Width int, Height int) [][]byte {

	bitMap := make([][]byte, Height)

	for i := range bitMap {
		bitMap[i] = make([]byte, Width)
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



func CaravanSelectDestination(Caravan *CaravanTemplate) {}


func CaravanMoveToTown(Caravan *CaravanTemplate) {}

func TownGetWarePrice(Town TownTemplate, WareId int) float64 {
	Price := Goods[WareId].PriceMin + (Goods[WareId].PriceMax - Goods[WareId].PriceMin)*(1.0 - Town.Wares[WareId].Quantity/Town.WarehouseLimit)
	return Price
}

func TownGetWareWithLowestPrice(Town *TownTemplate) int {
	return 0
}


func SellForBestPrice(Caravan *CaravanTemplate) Cargo {
	var c Cargo
	return c
}


func BuyForBestPrice(Caravan *CaravanTemplate) Cargo {
	var c Cargo
	return c
}


func RedrawViewMap(){

	textMap.SetText(PrintMap(GlobalMap, Towns, Caravan))
	fmt.Fprintf(textMap, "Size %dx%d Global step: %d\n", GlobalMap.Width, GlobalMap.Height, GlobalStep)

}


func RedrawViewCaravan(){

	CaravanStatus := fmt.Sprintf("Destination: %s (%d, %d)\nPosition: %d:%d\n\nCargo (%.0f/%.0f):\n", 
		Towns[Caravan.Target].Name, 
		Towns[Caravan.Target].Y, 
		Towns[Caravan.Target].X, 
		Caravan.Y, 
		Caravan.X,
		Caravan.Capacity,
		Caravan.CapacityMax)

		/*
		TradingGoodId int
		Quantity float64
		BuyPrice float64
		*/
	if len(Caravan.Cargo) > 0 {
		for _, cargo := range Caravan.Cargo {
			CaravanStatus += fmt.Sprintf("  %s qnt: %.0f prc: %.0f\n",
			Goods[cargo.TradingGoodId].Name,
			cargo.Quantity,
			cargo.BuyPrice)	
		}
	} else {
		CaravanStatus += "  none"
	}
		
	textCaravan.SetText(CaravanStatus + "\n")
}

func RedrawViewTown(){
	
	textTown.SetText(Towns[Caravan.Target].Name + "\n")

	for _, ware := range Towns[Caravan.Target].Wares {
		//Price := Goods[ware.Id].PriceMin + (Goods[ware.Id].PriceMax-Goods[ware.Id].PriceMin)*(1.0-ware.Quantity/Towns[Caravan.Target].WarehouseLimit)
		Price := math.Round(TownGetWarePrice(Towns[Caravan.Target], ware.Id))
		fmt.Fprintf(textTown,"%s: %.0f/%.0f Цена: %.0f\n", Goods[ware.Id].Name, ware.Quantity, Towns[Caravan.Target].WarehouseLimit, Price)
	}

	fmt.Fprintf(textTown,"\n")
}

func RedrawViewLog(){}

func RedrawScreen() {
	RedrawViewMap()
	RedrawViewTown()
	RedrawViewCaravan()
	RedrawViewLog()
}

func PrintToGameLog(Text string) {

}

func GlobalActions() {
	/*
	
	Глобальные действия:
		Город
			цикл производства

		Караван 
			перемещение по карте
			продать товары
			купить товары
		
		Перерисовать интерфейс
	*/

	CaravanMoveToTown(&Caravan)

	SellForBestPrice(&Caravan)

	BuyForBestPrice(&Caravan)

	

	// Перерисовать интерфейс после всех действий
	RedrawScreen()
}


func GlobalTick() {
	tick := time.NewTicker(refreshInterval)

	for {
		select {
		case <-tick.C:
			//t := time.Now().Format("15:04:05")

			GlobalStep++

			app.QueueUpdateDraw(func() {

				// Выполнить все действия
				GlobalActions()
				
				if Caravan.X == Towns[Caravan.Target].X && Caravan.Y == Towns[Caravan.Target].Y {

					fmt.Fprintf(textLog, "Arrived to destination \"%s\" at step %d\n", Towns[Caravan.Target].Name, GlobalStep)
					
					Towns[Caravan.Target].Visited++
					TotalVisited++

					//textTown.SetText(Towns[Caravan.Target].Name + "\n\n")

					//fmt.Fprintf(textTown,"Visited: %d\n\n", Towns[Caravan.Target].Visited)
					//TownsList := ""
					/*TownsList += fmt.Sprintf("Count: %d\n", len(Towns))
					TownsList += fmt.Sprintf("Total visited: %d\n", TotalVisited)
					TownsList += fmt.Sprintf("Calculated: %.2f%%\n\n", 1.0/float64(len(Towns))*100.0)*/

					/*for _, t := range Towns {
						TownsList += fmt.Sprintf("%s: %d %.2f%%\n", t.Name[0:1], t.Visited,  float64(t.Visited)/float64(TotalVisited)*100.0)
					}*/

					TradeId := 0	
					BuyAmount := 0.0
					/*
					TradingGoodId int
					Quantity float64
					BuyPrice float64
					*/
					
					// Sell
					if len(Caravan.Cargo) > 0 {
						//fmt.Fprintf(textLog, "Dummy sell action\n")
					}

					// Buy
					if Caravan.Capacity <  Caravan.CapacityMax {


						TradeId = Rnd(len(Towns[Caravan.Target].Wares) - 1)

						if Towns[Caravan.Target].Wares[TradeId].Quantity > Caravan.Capacity {
							//BuyAmount = Caravan.CapacityMax - Caravan.Capacity
							BuyAmount = 25

							Price := Goods[TradeId].PriceMin + (Goods[TradeId].PriceMax-Goods[TradeId].PriceMin)*(1.0 - Towns[Caravan.Target].Wares[TradeId].Quantity / Towns[Caravan.Target].WarehouseLimit)
							Price = math.Round(Price)
							
							Towns[Caravan.Target].Wares[TradeId].Quantity -= BuyAmount
							Caravan.Cargo = append(Caravan.Cargo, Cargo{TradingGoodId: TradeId, Quantity: BuyAmount, BuyPrice: Price})
							Caravan.Capacity += BuyAmount
						
							fmt.Fprintf(textLog, " - Bought: %s, quantity: %.0f, price: %.0f\n", Goods[TradeId].Name, BuyAmount, Price)

						} else {


						}

					}


					Caravan.PrevTarget = Caravan.Target

					for {
						Caravan.Target = RndRange(0, len(Towns)-1)
						if Caravan.Target != Caravan.PrevTarget {
							break
						}
					}

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
		
		/*{Id: 5, Tier: 2, Name: "Мука", PriceMin: 40, PriceMax: 75, SellingUnit: "мешок", UnitVolume: 0.036, UnitWeight: 0.050, 
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
			Consumables: []Resources{
				Resources{Id: 8, RequiredPerUnit: 1},
			},
		},*/
	}
}

func main() {

	var (
		Size          int
		MinDistance   int
		MaxTownsCount int
	)

	/*for _, Good := range Goods {
		fmt.Printf("%s[%d] цена: %.1f/%.1f\n", Good.Name, Good.Tier, Good.PriceMin, Good.PriceMax)
		for _, Resource := range Good.Resources {
			fmt.Printf("  %s[%d]: %d\n", Goods[Resource.Id].Name, Goods[Resource.Id].Tier, Resource.RequiredPerUnit)
		}
		for _, Consumable := range Good.Consumables {
			fmt.Printf("  * %s[%d]: %d\n", Goods[Consumable.Id].Name, Goods[Consumable.Id].Tier, Consumable.RequiredPerUnit)
		}
	}*/

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
		SetText("Loading...\n")

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
		SetText("Loading...")

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

	Size = 10
	MinDistance = 5
	MaxTownsCount = 26

	log.Println("Create global map")
	GlobalMap = MapDef{Width: Size * 2, Height: Size}

	log.Println("Create bitmap")
	bitMap := MakeBitMap(GlobalMap.Width, GlobalMap.Height)
	
	log.Println("Put towns on map")
	Towns = PutTownsOnMap(GlobalMap, &bitMap, MaxTownsCount, MinDistance)

	Caravan = CaravanTemplate{
		Name: "Caravan",
		Status: "At point",
		X: 1, Y: 1,
		CapacityMax: 100.0,
		TradeConfig: TradeConfig{
			BuyMaxPrice: 25,
			BuyFullCapacity: true,
			BuyMaxAmount: 50,
			SellWithProfit: true,
			SellMinPrice: 50,
		},
	}

	Caravan.Target = RndRange(0, len(Towns)-1)
	Caravan.PrevTarget = -1

	//fmt.Println(PrintMap(GlobalMap, Towns, Caravan))
	//fmt.Printf("%+v\n",GlobalMap)

//	os.Exit(0)

	go GlobalTick()

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
