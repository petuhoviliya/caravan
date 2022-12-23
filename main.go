package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"time"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const TickerInterval = 1000 * time.Millisecond

const CaravanStatusMoving uint8 = 1
const CaravanStatusInTown uint8 = 2
const CaravanStatusStarting uint8 = 255

type money int64


type GameTemplate struct {
	Tui tview.Application
	Map MapTemplate
	Towns []TownTemplate
	Caravan CaravanTemplate
}

type MapTemplate struct {
	Width  int
	Height int
	BitMap []byte
}

func NewMap(w int, h int) *MapTemplate {
	
	Map := MapTemplate{Width: w, Height: h}
	Map.MakeBitmap()

	return &Map
}

func (m *MapTemplate) Size() int {
	return m.Width * m.Height
}

func (m *MapTemplate) GetPosition(Index int) (int, int) {
	X := Index % m.Width
	Y := (Index - X) / m.Width
	return X, Y
}

func (m *MapTemplate) GetIndex(X int, Y int) int {
	return Y * m.Width + X
}

func (m *MapTemplate) MakeBitmap() {
	m.BitMap = make([]byte, m.Size())
}

func (m *MapTemplate) GetFreeCells() []int {

	var free []int

	for index, value := range m.BitMap {
		if value == 0 {
			free = append(free, index)
		}
	}

	return free
}

func (m *MapTemplate) PlaceTown(X int, Y int, Radius int) bool{

	for i := -Radius; i <= Radius; i++ {
		for j := -Radius; j <= Radius; j++ {

			A := float64(i)
			B := float64(j)
			C := int(math.Sqrt(math.Pow(A, 2) + math.Pow(B, 2)))

			tX := X + i
			tY := Y + j

			if C <= Radius {

				if tX > (m.Width - 1) {
					tX = (m.Width - 1)
				}
				if tY > (m.Height - 1) {
					tY = (m.Height - 1)
				}
				if tX < 0 {
					tX = 0
				}
				if tY < 0 {
					tY = 0
				}
				
				m.BitMap[m.GetIndex(tX, tY)] = 1
			}
		} // for j
	} // for i

	return true
}

func (m *MapTemplate) GetPrintableMap() string { return "" }


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

func NewCaravan() {}

func (c *CaravanTemplate) Move(X int, Y int) {}

func (c *CaravanTemplate) MoveBest(X int, Y int) {}

func (c *CaravanTemplate) ChooseDestination() {}

func (c *CaravanTemplate) CargoCapacity() {}

func (c *CaravanTemplate) Sell(TownId int, CargoId int) {}

func (c *CaravanTemplate) Buy(TownId int, CargoId int) {}



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

func NewTown(Id int, Name string, Tier int, X int, Y int) *TownTemplate {
	Town := TownTemplate{
		Id: Id,
		Name: Name,
		Tier: Tier,
		X: X,
		Y: Y,
	}
	return &Town
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
	GlobalMap    *MapTemplate
	GlobalPause  bool
	GlobalTicker *time.Ticker
	GlobalSpeedFactor time.Duration
	GlobalStep   int
	TotalVisited int

	Goods   map[int]TradingGood
	Towns   map[int]TownTemplate
	TownConfig map[int]TownConfigTemplate
	Caravan CaravanTemplate

	Tui *tview.Application
	app         *tview.Application
	textMap     *tview.TextView
	textLog     *tview.TextView
	textTown    *tview.TextView
	textCaravan *tview.TextView
	textStatus  *tview.TextView

	AlphabetRU []string
)

func RndRange(Min int, Max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(Max-Min+1)+Min
}

func Rnd(Max int) int {
	return RndRange(1, Max)
}

func PrintMap(Map MapTemplate, Towns map[int]TownTemplate, Caravan CaravanTemplate) string {
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
							mapObject = fmt.Sprintf("[%s]%s[%s]", ColorTag, town.Name[0:2], "white")
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


func PutTownsOnMap(Map MapTemplate, BitMap *[][]byte, TownCount int, MinDistance int) map[int]TownTemplate {

	var Towns map[int]TownTemplate
	var Wares map[int]WareGood
	var MaxTier2 int
	var MaxTier3 int
	var Tier2 int = 1
	var Tier3 int = 1
	
	Wares = make(map[int]WareGood)
	Towns = make(map[int]TownTemplate)


	if TownCount > len(AlphabetRU) {
		TownCount = len(AlphabetRU)
	}

	MaxTier2 = RndRange(2, 3)
	MaxTier3 = RndRange(1, 2)

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
			Wares[key] = WareGood{key, float64(Rnd(50))}
		}
		// Id, Name, Tier, X, Y, WarehouseLimit, Wares, Visited
		Towns[i] = TownTemplate{i, AlphabetRU[i-1], 1, FreeCells[NextFreeCell].X, FreeCells[NextFreeCell].Y, 500, Wares, 0}
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

func PointInsideRadius(X int, Y int, Radius int) bool {

	A := math.Abs(float64(0 - X))
	B := math.Abs(float64(0 - Y))
	C := int(math.Sqrt(math.Pow(A, 2) + math.Pow(B, 2)))


	//if int(math.Hypot(float64(X),float64(Y))) <= Radius {
	if C <= Radius {
		return true
	}
	return false
}


func PutTownOnBitMap(Map MapTemplate, BitMap *[][]byte, TownX int, TownY int, Distance int) {

	for i := -Distance; i <= Distance; i++ {
		for j := -Distance; j <= Distance; j++ {

			tX := TownX + i
			tY := TownY + j

			if PointInsideRadius(i, j, Distance) {
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

func RedrawViewStatus() {}

func RedrawScreen() {
	RedrawViewMap()
	RedrawViewTown()
	RedrawViewCaravan()
	RedrawViewLog()
	RedrawViewStatus()
}

func PrintToGameLog(Text string) {
	fmt.Fprintf(textLog, "%s", Text)
}

func PrintToStatusBar(Text string) {}

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
	GlobalTicker = time.NewTicker(TickerInterval/GlobalSpeedFactor)

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

func SetGameSpeed(SpeedFactor time.Duration) {
	GlobalSpeedFactor = SpeedFactor
	GlobalTicker.Reset(TickerInterval/GlobalSpeedFactor)
	SpeedStatus := fmt.Sprintf("Сжатие времени: [green]x%d[white]", GlobalSpeedFactor)
	textStatus.SetText(SpeedStatus)
}

func ToggleGamePause() {
	if GlobalPause == false {
		GlobalPause = true
		GlobalTicker.Stop()
		textMap.SetTitle("Карта - ПАУЗА")
	} else {
		GlobalPause = false
		GlobalTicker.Reset(TickerInterval/GlobalSpeedFactor)
		textMap.SetTitle("Карта")
	}
}

func InitGame() {
	/*
	Порядок действий

	0. Генерируем карту
	1. Генерируем города
		1.1 Распологаем города на карте
	2. Генерируем караван
	3. Запускаем гланый цикл

	*/

}

func InitInterface() {
	/*

	*/
}



func init() {

	log.Println("Init")

	// 1, 2, 4, 8
	GlobalSpeedFactor = 1


	AlphabetRU = []string{
		"Амурск", "Биробиджан", "Владивосток", "Грозный",
		"Дубна", "Ейск", "Жуковский",	"Зеленоград",
		"Иркутск", "Казань", "Липецк", "Мурманск",
		"Ноглики", "Омск", "Партизанск", "Рязань",
		"Смоленск", "Томск", "Уссурийск", "Феодосия",
		"Хабаровск", "Цимлянск", "Чита", "Шатура",
		"Щелково", "Элиста", "Южно-Сахалинск", "Якутск",
	}

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
	
	// Конфигурация города в зависимости от уровня (TownTemplate.Tier)
	TownConfig = map[int]TownConfigTemplate{
		//	WarehouseLimit, ColorTag
		1: {500.0, 	"[red]"},
		2: {1000.0, "[orange]"},
		3: {2000.0, "[green]"},
	}

	/*
		Id          int
		Tier        int
		Name        string
		PriceMin    float64
		PriceMax    float64
		Unit 				string
		UnitVolume  float64
		UnitWeight  float64
		Resources   []Resources
		Consumables []Resources
	*/

	Goods = map[int]TradingGood{
		// Id		Tier	Name 			PriceMin	PriceMax	Unit				Volume	Weight	Resources	Consumables
		1: {1, 	1, 		"Зерно", 	2, 				10, 			"мешок", 		0.036, 	0.050, 	nil, 			nil},
		2: {2, 	1, 		"Дерево",	5, 				20, 			"кубометр",	1.0, 		0.640, 	nil, 			nil},
		3: {3, 	1, 		"Камень", 4, 				18, 			"кубометр",	1.0, 		1.7, 		nil, 			nil},
		4: {4, 	1, 		"Руда", 	9, 				30, 			"тонна", 		0.5, 		1.0, 		nil, 			nil},
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
	{Id: 8, Tier: 2, Name: "Металлический слиток", PriceMin: 1, PriceMax: 100, SellingUnit: "партия", UnitVolume: 0.0, UnitWeight: 0.0,
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

	app = tview.NewApplication()

	textStatus = tview.NewTextView().
		SetDynamicColors(true).
		SetText(fmt.Sprintf("Сжатие времени: [green]х1[white] х2 х4 х8"))

	textStatus.
		SetBorder(true).
		SetTitleAlign(tview.AlignLeft).
		SetTitle("Статус")


	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		//textStatus.SetText(fmt.Sprintf("%+v", event.Rune()))
		switch event.Rune() {
		case 32:
			// spacebar
			ToggleGamePause()
		case 49:
			// 1
			SetGameSpeed(1)
		case 50:
			// 2
			SetGameSpeed(2)
		case 51:
			// 3
			SetGameSpeed(4)
		case 52:
			// 4
			SetGameSpeed(8)
		}
		return event
	})

	textMap = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(false).
		SetWordWrap(false).
		SetText("Загружается...")

	textMap.
		SetBorder(true).
		SetTitleAlign(tview.AlignLeft).
		SetTitle("Карта")

	textLog = tview.NewTextView().
		SetScrollable(true).
		SetWrap(true).
		SetWordWrap(true).
		SetMaxLines(100).
		SetText("Загружается...\n")

	textLog.
		SetBorder(true).
		SetTitleAlign(tview.AlignLeft).
		SetTitle("Журнал")

	textTown = tview.NewTextView().
		SetScrollable(true).
		SetWrap(true).
		SetWordWrap(true).
		SetText("Загружается...")

	textTown.
		SetBorder(true).
		SetTitleAlign(tview.AlignLeft).
		SetTitle("Города")

	textCaravan = tview.NewTextView().
		SetScrollable(true).
		SetWrap(true).
		SetWordWrap(true).
		SetText("Загружается...")

	textCaravan.
		SetBorder(true).
		SetTitleAlign(tview.AlignLeft).
		SetTitle("Караван")


	grid := tview.NewGrid().
		SetRows(-15, -15, -2).
		SetColumns(-2, -2, -2).
		SetMinSize(15, 20).
		SetBorders(false)

	grid.AddItem(textMap, 0, 0, 1, 2, 0, 0, false).
		AddItem(textLog, 0, 2, 3, 1, 0, 0, false).
		AddItem(textTown, 1, 0, 1, 1, 0, 0, false).
		AddItem(textCaravan, 1, 1, 1, 1, 0, 0, false).
		AddItem(textStatus, 2, 0, 1, 2, 0, 0, false)


	log.Println("Create global map")
	GlobalMap = NewMap(60,15)

	log.Println("Put towns on map")
	Towns = PutTownsOnMap(GlobalMap, &bitMap, 28, 5)

	log.Println("Generate Caravan")
	Caravan = CaravanTemplate{
		Name:        "Караван",
		Status:      CaravanStatusStarting,
		X:           1,
		Y:           1,
		Money:       1000,
		CapacityMax: 100.0,
		TradeConfig: TradeConfig{
			BuyMaxPrice:     0.25, // Покупать если удовлетворено условие:  Цена <= BuyMaxPrice * (PriceMin + (PriceMax - PriceMin))
			BuyFullCapacity: true, // Стараться купить Кол-во равное CapacityMax, если получится, то покупается несколько видов товаров
			BuyMaxAmount:    0.50, // Если BuyFullCapacity == false, то Кол-во покупаемого товара не более чем BuyMaxAmount * CapacityMax
			BuyMinAmount:    0.10, // Минимальное кол-во для покупки BuyMinAmount * CapacityMax
			SellWithProfit:  true, // Всегда продавать по цене большей чем цена покупки
			SellMinPrice:    0.50, // Если SellWithProfit == false, то продавать если Цена >= SellMinPrice * (PriceMin + (PriceMax - PriceMin))
		},
	}

	Caravan.Target = RndRange(1, len(Towns))
	Caravan.PrevTarget = -1

	//fmt.Println(PrintMap(GlobalMap, Towns, Caravan))
	//fmt.Printf("%+v\n",GlobalMap)

	//	os.Exit(0)

	InitGame()

	go GlobalTick()

	if err := app.SetRoot(grid, true).SetFocus(grid).Run(); err != nil {
		panic(err)
	}

	os.Exit(0)


}


	/*ttMap := NewMap(60, 15)

	/*for i := 0; i < ttMap.Size(); i++ {
		ttX, ttY := ttMap.GetPosition(i)
		ttI := ttMap.GetIndex(ttX, ttY)
		fmt.Printf("%d - %d:%d (%d)\n ", i, ttX, ttY, ttI)
	}

	//os.Exit(0)

	ttTowns := []TownTemplate{}

/*	ttMap.PlaceTown(1, 1, 5)
//	ttMap.PlaceTown(15, 1, 5)
//	ttMap.PlaceTown(1, 10, 5)
//	ttMap.PlaceTown(15, 10, 5)

	for k, v := range ttMap.BitMap{
		if k % ttMap.Width == 0 {
			fmt.Printf("\n")
		}
		fmt.Printf("%+v ",v)
	}
	fmt.Printf("\n")

	fmt.Printf("FreeCelss: %+v\n", ttMap.GetFreeCells())

	fmt.Printf("Free cells count: %v\n", len(ttMap.GetFreeCells()))

	index := RndRange(0, len(ttMap.GetFreeCells())-1)

	fmt.Printf("Index: %d, value: %v\n", index, ttMap.GetFreeCells()[index])

	nextFreeCell :=  ttMap.GetFreeCells()[index]

	ttX, ttY := ttMap.GetPosition(nextFreeCell)
	
	fmt.Printf("Pos: %d:%d\n", ttX, ttY)

	
	// ----
	ttMap.PlaceTown(ttX, ttY, 5)

	for k, v := range ttMap.BitMap{
		if k % ttMap.Width == 0 {
			fmt.Printf("\n")
		}
		fmt.Printf("%+v ",v)
	}
	fmt.Printf("\n")

	fmt.Printf("FreeCelss: %+v\n", ttMap.GetFreeCells())

	fmt.Printf("Free cells count: %v\n", len(ttMap.GetFreeCells()))
	
	//os.Exit(0)

	for i:=1; i <= len(AlphabetRU); i++ {

		freeCells := ttMap.GetFreeCells()

		if len(freeCells) == 0 {
			break
		}
		/*for k, v := range ttMap.BitMap{
			if k % ttMap.Width == 0 {
				fmt.Printf("\n")
			}
			fmt.Printf("%d",v)
		}
		fmt.Printf("\n")

		fmt.Printf("%v\n", ttMap.BitMap)

		fmt.Printf("Free: %d\n",len(freeCells))

		index := RndRange(0, len(freeCells)-1)

		ttX, ttY := ttMap.GetPosition(freeCells[index])

		//fmt.Printf("%d - %d:%d\n" ,index, ttX, ttY)
		//fmt.Printf("%v\n\n", ttMap.GetFreeCells())

		if ttMap.PlaceTown(ttX, ttY, 5) {
			ttTowns = append(ttTowns, TownTemplate{Id: i, Name: AlphabetRU[i-1] ,X: ttX,Y: ttY,},)
		}
	}


	for k, v := range ttMap.BitMap{
		if k % ttMap.Width == 0 {
			fmt.Printf("\n")
		}
		fmt.Printf("%d",v)
	}
	fmt.Printf("\n")


	fmt.Println(len(ttMap.GetFreeCells()))
	fmt.Printf("%+v\n",ttTowns)
	fmt.Println(len(ttTowns))

	os.Exit(0)
	*/


/*	Alphabet = []string{
		"Alpha", "Bravo", "Charlie", "Delta",
		"Echo", "Foxtrot", "Golf", "Hotel",
		"India", "Juliet", "Kilo", "Lima",
		"Mike", "November", "Oscar", "Papa",
		"Quebec", "Romeo", "Sierra", "Tango",
		"Uniform", "Victor", "Whiskey", "X-ray",
		"Yankee", "Zulu",
	}*/

/*	AlphabetRU = []string{
		"Анна", "Борис", "Василий", "Григорий",
		"Дмитрий", "Елена", "Ёлка", "Женя",
		"Зинаида", "Иван", "Константин", "Леонид",
		"Михаил", "Николай", "Ольга", "Павел",
		"Роман", "Семен", "Татьяна", "Ульяна",
		"Федор", "Харитон", "Цапля", "Человек",
		"Шура", "Щука", "Эхо", "Юрий", "Яков",
	}*/

