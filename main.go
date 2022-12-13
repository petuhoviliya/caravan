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
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const refreshInterval = 1000 * time.Millisecond

// Status:
//
//	At point
//	Moving
const CaravanStatusMoving uint8 = 1
const CaravanStatusInTown uint8 = 2
const CaravanStatusStarting uint8 = 255

const TownBaseWarehouseLimit = 500.0

const MapSize = 15
const MapMinDistance = 5
const MapMaxTownsCount = 26

type PositionDef struct {
	X int
	Y int
}

type MapDef struct {
	Width  int
	Height int
}

type Cargo struct {
	WareId   int
	TownId   int
	Quantity float64
	BuyPrice float64
}

type TradeConfig struct {

	// Максимальная цена покупки товара
	// в процентах
	// TradingGood.PriceMin - 0%
	// TradingGood.PriceMax - 100%
	BuyMaxPrice float64

	// Всегда покупать максимально возможное количество до полной емкости
	BuyFullCapacity bool

	// Максимальное кол-во покупки товара
	// в процентах от CaravanTemplate.CapacityMax
	// Значение игнорируется, если BuyFullCapacity == true
	BuyMaxAmount float64

	// Минимальное кол-во к покупке
	// в процентах от CaravanTemplate.CapacityMax
	BuyMinAmount float64

	// Всегда продавать с прибылью
	// Цена продажи не может быть ниже цены покупки
	SellWithProfit bool

	// Минимальная цена продажи товара
	// в процентах
	// TradingGood.PriceMin - 0%
	// TradingGood.PriceMax - 100%
	SellMinPrice float64
}

type CaravanTemplate struct {
	Name        string
	Status      uint8
	Money       float64
	X           int
	Y           int
	Target      int
	PrevTarget  int
	CapacityMax float64
	Cargo       []Cargo
	TradeConfig TradeConfig
}

type TownConfigTemplate struct {
	WarehouseLimit float64
	ColorTag string
}


type TownTemplate struct {
	Id             int
	Name           string
	Tier           int
	X              int
	Y              int
	WarehouseLimit float64
	Wares          map[int]WareGood
	Visited        int
}

type FreeCell struct {
	X int
	Y int
}

type Resources struct {
	Id              int
	RequiredPerUnit int
}

type TradingGood struct {
	Id          int
	Tier        int
	Name        string
	PriceMin    float64
	PriceMax    float64
	Unit        string
	UnitVolume  float64
	UnitWeight  float64
	Resources   []Resources
	Consumables []Resources
}

type WareGood struct {
	Id       int
	Quantity float64
}

var (
	GlobalMap    MapDef
	GlobalPause  bool
	GlobalTicker *time.Ticker
	GlobalStep   int
	TotalVisited int

	Goods   map[int]TradingGood
	Towns   map[int]TownTemplate
	TownConfig map[int]TownConfigTemplate
	Caravan CaravanTemplate

	app         *tview.Application
	textMap     *tview.TextView
	textLog     *tview.TextView
	textTown    *tview.TextView
	textCaravan *tview.TextView
)

func PrintMap(Map MapDef, Towns map[int]TownTemplate, Caravan CaravanTemplate) string {
	// ╗ ╝ ╚ ╔ ╩ ╦ ╠ ═ ║ ╬ ╣ - borders
	// │ ┤ ┐ └ ┴ ┬ ├ ─ ┼ ┘ ┌ - roads
	// @ - caravan

	var PrintableMap string
	var ColorTag string

	for posX := 0; posX <= Map.Height+1; posX++ {
		for posY := 0; posY <= Map.Width+1; posY++ {

			//fmt.Printf("X:%d, Y:%d\n", posX, posY)

			if posY == 0 {
				if posX == 0 { // left upper corner
					PrintableMap = PrintableMap + "╔"
				} else if posX == (Map.Height + 1) { // left bottom corner
					PrintableMap = PrintableMap + "╚"
				} else {
					PrintableMap = PrintableMap + "║" // left border
				}

			} else if posY == (Map.Width + 1) {
				if posX == 0 { // right upper corner
					PrintableMap = PrintableMap + "╗\n"
				} else if posX == (Map.Height + 1) { // right  bottom corner
					PrintableMap = PrintableMap + "╝\n"
				} else {
					PrintableMap = PrintableMap + "║\n" // right border
				}

			} else {
				if posX == (Map.Height+1) || posX == 0 { // up and bottom border
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

func PutTownsOnMap(Map MapDef, BitMap *[][]byte, TownCount int, MinDistance int) map[int]TownTemplate {

	var Towns map[int]TownTemplate
	var Wares map[int]WareGood
	var MaxTier2 int
	var MaxTier3 int
	var Tier2 int = 1
	var Tier3 int = 1

	if TownCount > 26 {
		TownCount = 26
	}

	MaxTier2 = RndRange(2, 3)
	MaxTier3 = RndRange(1, 2)
	
	Wares = make(map[int]WareGood)
	Towns = make(map[int]TownTemplate)

	alphabet := [26]string{
		"Alpha", "Bravo", "Charlie", "Delta",
		"Echo", "Foxtrot", "Golf", "Hotel",
		"India", "Juliet", "Kilo", "Lima",
		"Mike", "November", "Oscar", "Papa",
		"Quebec", "Romeo", "Sierra", "Tango",
		"Uniform", "Victor", "Whiskey", "X-ray",
		"Yankee", "Zulu",
	}

	for i := 1; i <= TownCount; i++ {

		FreeCells := GetFreeCellsOnMap((*BitMap))

		if len(FreeCells) == 0 {
			break
		}

		//fmt.Printf("Free: %d\n", len(FreeCells))

		NextFreeCell := RndRange(0, len(FreeCells)-1)
		PutTownOnBitMap(Map, &(*BitMap), FreeCells[NextFreeCell].X, FreeCells[NextFreeCell].Y, MinDistance)

		Wares = make(map[int]WareGood)

		for key := range Goods {
			Wares[key] = WareGood{key, float64(Rnd(500))}
		}
		// Id, Name, Tier, X, Y, WarehouseLimit, Wares, Visited
		Towns[i] = TownTemplate{i, alphabet[i-1], 1, FreeCells[NextFreeCell].X, FreeCells[NextFreeCell].Y, TownBaseWarehouseLimit, Wares, 0}
	}


	for {
		//break
		if Tier2 > MaxTier2 {
			break
		}
		t := Rnd(len(Towns))
		if Towns[t].Tier == 1 {
			Tier2++
			t1 := Towns[t]
			t1.Tier = 2
			Towns[t] = t1
		}
	}

	for {
		//break
		if Tier3 > MaxTier3 {
			break
		}
		t := Rnd(len(Towns))
		if Towns[t].Tier == 1 {
			Tier3++
			t1 := Towns[t]
			t1.Tier = 3
			Towns[t] = t1
		}
	}

	for _, t := range Towns {
		fmt.Printf("%+v\n", t)
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
				if tY > (Map.Width - 1) {
					tY = Map.Width - 1
				}
				if tX > (Map.Height - 1) {
					tX = Map.Height - 1
				}
				if tX < 0 {
					tX = 0
				}
				if tY < 0 {
					tY = 0
				}

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

func CaravanCalculateCargoCapacity(Caravan *CaravanTemplate) float64 {

	var Capacity float64

	if len(Caravan.Cargo) == 0 {
		return 0
	}

	for _, c := range Caravan.Cargo {
		Capacity += c.Quantity
	}
	//Caravan.Capacity = Capacity
	return Capacity
}

func CaravanSelectDestination(Caravan *CaravanTemplate) {

	Caravan.PrevTarget = Caravan.Target

	for {
		Caravan.Target = RndRange(1, len(Towns))
		if Caravan.Target != Caravan.PrevTarget {
			break
		}
	}
}

func CaravanMoveToTown(Caravan *CaravanTemplate) {

	Caravan.X, Caravan.Y = FindBestNextPoint(Caravan.X, Caravan.Y, Towns[Caravan.Target].X, Towns[Caravan.Target].Y)

	if Caravan.X == Towns[Caravan.Target].X && Caravan.Y == Towns[Caravan.Target].Y {

		TextLog := fmt.Sprintf("[%d]: Прибыл в \"%s\"\n", GlobalStep, Towns[Caravan.Target].Name)
		PrintToGameLog(TextLog)

		Caravan.Status = CaravanStatusInTown
		CaravanSelectDestination(Caravan)

	} else {
		Caravan.Status = CaravanStatusMoving
	}
}

func TownGetWarePrice(Town TownTemplate, WareId int) float64 {
	Price := Goods[WareId].PriceMin + (Goods[WareId].PriceMax-Goods[WareId].PriceMin)*(1.0-Town.Wares[WareId].Quantity/Town.WarehouseLimit)
	return Price
}

func TownGetWareWithLowestPrice(Town TownTemplate) int {

	var LowestPrice = math.Inf(1)
	var Price float64
	var Id int

	for _, Ware := range Town.Wares {
		Price = TownGetWarePrice(Town, Ware.Id)
		if Price < LowestPrice {
			LowestPrice = Price
			Id = Ware.Id
		}
	}

	//PrintToGameLog(fmt.Sprintf("Ware: %d\n%+v\n\n", Id, Town))

	return Id
}

func SellForBestPrice(Caravan *CaravanTemplate) {
	if Caravan.Status != CaravanStatusInTown {
		return
	}
	return
}

func BuyForBestPrice(Caravan *CaravanTemplate) {

	var (
		TradeId      int
		Price        float64
		BuyAmount    float64
		MaxBuyAmount float64
		TextLog      string
	)

	if Caravan.Status != CaravanStatusInTown {
		return
	}

	/*
		BuyMaxPrice:     0.25,
		BuyFullCapacity: true,
		BuyMaxAmount:    0.50,
		BuyMinAmount:    0.10,
		SellWithProfit:  true,
		SellMinPrice:    0.50,
	*/

	MaxBuyAmount = Caravan.CapacityMax - CaravanCalculateCargoCapacity(Caravan)
	//fmt.Fprintf(textLog, "- Можем купить: %.0f (%0.f - %0.f)\n", MaxBuyAmount, Caravan.CapacityMax, CaravanCalculateCargoCapacity(Caravan))

	TradeId = TownGetWareWithLowestPrice(Towns[Caravan.PrevTarget])
	//fmt.Fprintf(textLog, "- К покупке: id: %d, %s\n", TradeId, Goods[TradeId].Name)

	if Towns[Caravan.PrevTarget].Wares[TradeId].Quantity > MaxBuyAmount {
		BuyAmount = MaxBuyAmount
	} else {
		BuyAmount = Towns[Caravan.PrevTarget].Wares[TradeId].Quantity
	}

	if BuyAmount == 0 {
		return
	}

	Price = TownGetWarePrice(Towns[Caravan.PrevTarget], TradeId)

	if BuyAmount > math.Floor(Caravan.Money/Price) {
		BuyAmount = math.Floor(Caravan.Money / Price)
	}

	//fmt.Fprintf(textLog, "--- %+v\n", Towns[Caravan.PrevTarget].Wares[TradeId])

	Towns[Caravan.PrevTarget].Wares[TradeId] = WareGood{TradeId, Towns[Caravan.PrevTarget].Wares[TradeId].Quantity - BuyAmount}

	Caravan.Cargo = append(Caravan.Cargo, Cargo{WareId: TradeId, TownId: Caravan.PrevTarget, Quantity: BuyAmount, BuyPrice: Price})

	Caravan.Money -= BuyAmount * Price

	TextLog = fmt.Sprintf("  Куплено: %s, кол-во: %.1f, цена: %.2f\n", Goods[TradeId].Name, BuyAmount, Price)

	PrintToGameLog(TextLog)

	return
}

func RedrawViewMap() {
	textMap.SetText(PrintMap(GlobalMap, Towns, Caravan))
	fmt.Fprintf(textMap, "Размер %dx%d Глобальный шаг: %d\n", GlobalMap.Width, GlobalMap.Height, GlobalStep)
}

func RedrawViewCaravan() {

	CaravanStatus := fmt.Sprintf("Назначение: %s (%d, %d)\nПозиция: %d:%d\nДеньги: %.2f\n\nГруз (%.0f/%.0f):\n",
		Towns[Caravan.Target].Name,
		Towns[Caravan.Target].Y,
		Towns[Caravan.Target].X,
		Caravan.Y,
		Caravan.X,
		Caravan.Money,
		CaravanCalculateCargoCapacity(&Caravan),
		Caravan.CapacityMax)

	/*
		TradingGoodId int
		Quantity float64
		BuyPrice float64
	*/
	if len(Caravan.Cargo) > 0 {
		for _, cargo := range Caravan.Cargo {
			CaravanStatus += fmt.Sprintf("  %s кол: %.0f, цена: %.2f, куплено в: %s\n",
				Goods[cargo.WareId].Name,
				cargo.Quantity,
				cargo.BuyPrice,
				Towns[cargo.TownId].Name)
		}
	} else {
		CaravanStatus += "  нет"
	}

	textCaravan.SetText(CaravanStatus + "\n")
}

func RedrawViewTown() {

	textTown.SetText("")

	// Текущий пункт назначения
	fmt.Fprintf(textTown, "Куда идем: %s\n", Towns[Caravan.Target].Name)

	for key := 1; key <= len(Towns[Caravan.Target].Wares); key++ {
		Price := TownGetWarePrice(Towns[Caravan.Target], key)

		fmt.Fprintf(textTown, "%s: %.0f/%.0f Цена: %.2f\n", Goods[key].Name, Towns[Caravan.Target].Wares[key].Quantity, Towns[Caravan.Target].WarehouseLimit, Price)
	}

	fmt.Fprintf(textTown, "\n")

	// Предыдущий пункт назначения
	if Caravan.PrevTarget != -1 {

		fmt.Fprintf(textTown, "Где был: %s\n", Towns[Caravan.PrevTarget].Name)

		for key := 1; key <= len(Towns[Caravan.PrevTarget].Wares); key++ {
			Price := TownGetWarePrice(Towns[Caravan.PrevTarget], key)
			fmt.Fprintf(textTown, "%s: %.0f/%.0f Цена: %.2f\n", Goods[key].Name, Towns[Caravan.PrevTarget].Wares[key].Quantity, Towns[Caravan.PrevTarget].WarehouseLimit, Price)
		}

		fmt.Fprintf(textTown, "\n\n")
	}
}

func RedrawViewLog() {}

func RedrawScreen() {
	RedrawViewMap()
	RedrawViewTown()
	RedrawViewCaravan()
	RedrawViewLog()
}

func PrintToGameLog(Text string) {
	fmt.Fprintf(textLog, "%s", Text)
}

func GlobalActions() {
	/*

		Глобальные действия:
			Город
				цикл производства

			Караван
				продать товары
				купить товары
				перемещение по карте

			Перерисовать интерфейс
	*/

	CaravanMoveToTown(&Caravan)

	SellForBestPrice(&Caravan)

	BuyForBestPrice(&Caravan)

	// Перерисовать интерфейс после всех действий
	RedrawScreen()
}

func GlobalTick() {
	GlobalTicker = time.NewTicker(refreshInterval)

	for {
		select {
		case <-GlobalTicker.C:
			//t := time.Now().Format("15:04:05")

			GlobalStep++

			app.QueueUpdateDraw(func() {

				// Выполнить все действия
				GlobalActions()

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

	
	// Конфигурация города в зависимости от уровня (TownTemplate.Tier)
	TownConfig = map[int]TownConfigTemplate{
		//						WarehouseLimit, ColorTag
		1: TownConfigTemplate{500.0, "[red]"},
		2: TownConfigTemplate{1000.0, "[orange]"},
		3: TownConfigTemplate{2000.0, "[green]"},
	}

	/*
		Id          int
		Tier        int
		Name        string
		PriceMin    float64
		PriceMax    float64
		SellingUnit string
		UnitVolume  float64
		UnitWeight  float64
		Resources   []Resources
		Consumables []Resources
	*/

	Goods = map[int]TradingGood{
		//						Id, Tier, Name, PriceMin, PriceMax, Unit
		1: TradingGood{1, 1, "Зерно", 2, 10, "мешок", 0.036, 0.050, nil, nil},
		2: TradingGood{2, 1, "Дерево", 5, 20, "кубометр", 1.0, 0.640, nil, nil},
		3: TradingGood{3, 1, "Камень", 4, 18, "кубометр", 1.0, 1.7, nil, nil},
		4: TradingGood{4, 1, "Руда", 9, 30, "тонна", 0.5, 1.0, nil, nil},
	}
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

func main() {

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

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case 256:
			if GlobalPause == false {
				GlobalPause = true
				GlobalTicker.Stop()
				textMap.SetTitle("Map - ПАУЗА")
			} else {
				GlobalPause = false
				GlobalTicker.Reset(refreshInterval)
				textMap.SetTitle("Map")
			}
		}
		return event
	})

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
		SetColumns(-2, -2, -2).
		SetMinSize(15, 20).
		SetBorders(false)

	grid.AddItem(textMap, 0, 0, 1, 2, 0, 0, false).
		AddItem(textLog, 0, 2, 2, 1, 0, 0, false).
		AddItem(textTown, 1, 0, 1, 1, 0, 0, false).
		AddItem(textCaravan, 1, 1, 1, 1, 0, 0, false)

	//os.Exit(0)

	//SetupCloseHandler()

/*	Size = 15
	MinDistance = 5
	MaxTownsCount = 26*/

	log.Println("Create global map")
	GlobalMap = MapDef{Width: MapSize * 2, Height: MapSize}

	log.Println("Create bitmap")
	bitMap := MakeBitMap(GlobalMap.Width, GlobalMap.Height)

	log.Println("Put towns on map")
	Towns = PutTownsOnMap(GlobalMap, &bitMap, MapMaxTownsCount, MapMinDistance)

	log.Println("Generate Caravan")
	Caravan = CaravanTemplate{
		Name:        "Caravan",
		Status:      CaravanStatusStarting,
		X:           1,
		Y:           1,
		Money:       1000,
		CapacityMax: 100.0,
		TradeConfig: TradeConfig{
			BuyMaxPrice:     0.25,
			BuyFullCapacity: true,
			BuyMaxAmount:    0.50,
			BuyMinAmount:    0.10,
			SellWithProfit:  true,
			SellMinPrice:    0.50,
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
