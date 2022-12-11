package main

import (
  "fmt"
  "log"
  "os"
  "time"

  "github.com/gdamore/tcell/v2"
  "github.com/petuhoviliya/caravan/internal/caravan"
  "github.com/petuhoviliya/caravan/internal/common"
  "github.com/petuhoviliya/caravan/internal/game"
  "github.com/petuhoviliya/caravan/internal/logger"
  "github.com/petuhoviliya/caravan/internal/town"
  _ "github.com/petuhoviliya/caravan/internal/world"
  "github.com/rivo/tview"
)

const (
  TickerInterval = 1000 * time.Millisecond
)

/*
Слои карты:
  0 - рельеф: равнина, пустыня, степь, лес, горы
  1 - водоемы: реки, озера, моря
  2 - дороги: центральные, обычные, мосты
  3 - населенные пункты
  4 - точки интереса
  5 - существа: нейтральные и враждебные

*/

/*
type Resources struct {
  Id              int
  RequiredPerUnit int
}

type TradingGood struct {
  Id          int
  Tier        int
  Name        string
  PriceMin    int64
  PriceMax    int64
  Unit        string
  UnitVolume  float64
  UnitWeight  float64
  Resources   []Resources
  Consumables []Resources
}
*/

var (
  Game game.Template

  TotalVisited int

  Towns     map[int]town.Template
  Goods     map[int]common.TradingGood
  Caravan   caravan.Template
  Status    map[string]uint8
  StatusNum map[uint8]string

  Tui         *tview.Application
  textMap     *tview.TextView
  textLog     *tview.TextView
  textTown    *tview.TextView
  textCaravan *tview.TextView
  textStatus  *tview.TextView

  //AlphabetRU []string
)

func RedrawViewMap() {
  textMap.SetText(Game.PrintableMap())
  fmt.Fprintf(textMap, "Размер %dx%d Глобальный шаг: %d\n", Game.World.Width, Game.World.Height, Game.Step)
}

func RedrawViewCaravan() {

  CaravanStatus := fmt.Sprintf("Назначение: %s (%d, %d)\nПозиция: %d:%d\nДеньги: %d\n\nГруз (%d/%d):\n",
    Game.Towns[Game.Caravan.Target].Name,
    Game.Towns[Game.Caravan.Target].X+1,
    Game.Towns[Game.Caravan.Target].Y+1,
    Game.Caravan.X+1,
    Game.Caravan.Y+1,
    Game.Caravan.Money,
    Game.Caravan.CargoCapacity(),
    Game.Caravan.CapacityMax)

  /*
     TradingGoodId int
     Quantity float64
     BuyPrice float64
  */

  if len(Game.Caravan.Cargo) > 0 {
    for _, cargo := range Game.Caravan.Cargo {
      CaravanStatus += fmt.Sprintf("  %s кол: %d, цена: %.2f, куплено в: %s\n",
        Goods[cargo.GoodId].Name,
        cargo.Quantity,
        cargo.BuyPrice,
        Game.Towns[cargo.TownId].Name)
    }
  } else {
    CaravanStatus += "  нет"
  }

  textCaravan.SetText(CaravanStatus + "\n")
}

func RedrawViewTown() {

  var (
    txt      string
    nextTown town.Template
    //prevTown town.Template
  )

  /*if Game.TotalVisited > 0 {
      for k := 0; k <= len(Game.Towns)-1; k++ {
        txt += fmt.Sprintf("%s: %d (%.2f%%)\n", Game.Towns[k].Name, Game.Towns[k].Visited, float64(Game.Towns[k].Visited)/float64(Game.TotalVisited)*100)
      }
    }
    textTown.SetText(txt)*/

/*
↻ ↺ ↻ ⇅ ⇵ ↑ ↓ ← →
⇅ 0 - продается и покупается
↑ 1 - только продается
↓ 2 - только покупается
↻ 3 - потребляется как ресурс для производства, только покупается
↺ 4 - производимый товар, только продается
*/
	//arrows :=  []string{"⇅","↑","↓","↻","↺"}

  nextTown = Game.Towns[Game.Caravan.Target]
  /*if Game.Caravan.PrevTarget != -1 {
    prevTown = Game.Towns[Game.Caravan.PrevTarget]
  }*/


  // Текущий пункт назначения
  txt += fmt.Sprintf("Куда идем: %s\n", nextTown.Name)

  for key := 1; key <= len(nextTown.Goods); key++ {
    Price := nextTown.GetGoodPrice(key)
    txt += fmt.Sprintf("[%s] %s: %d/%d Цена: %f\n",town.GoodStatus[nextTown.Goods[key].Status], Game.Goods[key].Name, nextTown.Goods[key].Quantity, nextTown.WarehouseLimit, Price)
  }


  /*txt += fmt.Sprintf("\n")
    // Предыдущий пункт назначения
    if prevTown.Name != "" {
      txt += fmt.Sprintf("Где был: %s\n", prevTown.Name)
      for key := 1; key <= len(prevTown.Goods); key++ {
        Price := prevTown.GetGoodPrice(key)
        txt += fmt.Sprintf("%s: %d/%d Цена: %.2f\n", Game.Goods[key].Name, prevTown.Goods[key].Quantity, prevTown.WarehouseLimit, Price)
      }
      txt += fmt.Sprintf("\n\n")
    }*/

  textTown.SetText(txt)
}

func RedrawViewLog() {}

func RedrawViewStatus() {}

func RedrawScreen() {
  RedrawViewMap()
  RedrawViewTown()
  RedrawViewCaravan()
  //RedrawViewLog()
  //RedrawViewStatus()
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
	Game.Caravan.PayTaxes(Game.Step)
  Game.CaravanMoveToTown()
	

  //  SellForBestPrice(&Caravan)

  //  BuyForBestPrice(&Caravan)

  // Перерисовать интерфейс после всех действий
  RedrawScreen()
}

func GlobalTick() {

  for {
    select {
    case <-Game.Ticker.C:

      Game.Step++

      Tui.QueueUpdateDraw(func() {

        // Выполнить все действия
        GlobalActions()

      })
    }
  }
}

func SetGameSpeed(TimeFactor time.Duration) {
  Game.TimeFactor = TimeFactor
  Game.Ticker.Reset(TickerInterval / Game.TimeFactor)
  SpeedStatus := fmt.Sprintf("Сжатие времени: [green]x%d[white]", Game.TimeFactor)
  textStatus.SetText(SpeedStatus)
}

func ToggleGamePause() {
  if !Game.Pause {
    Game.Pause = true
    Game.Ticker.Stop()
    textMap.SetTitle("Карта - ПАУЗА")
  } else {
    Game.Pause = false
    Game.Ticker.Reset(TickerInterval / Game.TimeFactor)
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
  Game.Ticker = time.NewTicker(TickerInterval / Game.TimeFactor)

  ToggleGamePause()
  RedrawScreen()
}

func InitInterface() {
  /*

   */
}

func init() {

  log.Println("Init")
  //rSeed := time.Now().UnixNano()
  var rSeed int64 = time.Now().UnixNano()

  //rand.Seed(1676424407175440563)
  //rand.Seed(rSeed)
  log.Printf("Seed: %d\n", rSeed)

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

  /*const CaravanStatusStarting uint8 = 255
    const CaravanStatusArrived uint8 = 1
    const CaravanStatusRestock uint8 = 2
    const CaravanStatusSelling uint8 = 3
    const CaravanStatusBuying uint8 = 4
    const CaravanStatusInTown uint8 = 5
    const CaravanStatusDeparture uint8 = 6
    const CaravanStatusMoving uint8 = 7 */

  Status = map[string]uint8{
    "Starting":  255,
    "Arrived":   1,
    "Restock":   2,
    "Selling":   3,
    "Buying":    4,
    "InTown":    5,
    "Departure": 6,
    "Moving":    7,
  }
  StatusNum = make(map[uint8]string)

  for k, v := range Status {
    StatusNum[v] = k
  }

  // Конфигурация города в зависимости от уровня (TownTemplate.Tier)
  /*  TownConfig = map[int]TownConfigTemplate{
      //  WarehouseLimit, ColorTag
      1: {500.0, "[red]"},
      2: {1000.0, "[orange]"},
      3: {2000.0, "[green]"},
    }*/

  /*
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
  */
  Goods = map[int]common.TradingGood{
    // Id   Tier  Name      PriceMin  PriceMax  Unit        Volume  Weight  Resources  Consumables
    1:  {1, 1, "Зерно", 2, 5, "мешок", 0.036, 0.050, nil, nil},
    2:  {2, 1, "Дерево", 3, 5, "кубометр", 1.0, 0.640, nil, nil},
    3:  {3, 1, "Камень", 1, 2, "кубометр", 1.0, 1.7, nil, nil},
    4:  {4, 1, "Руда", 4, 9, "тонна", 0.5, 1.0, nil, nil},
    5:  {5, 2, "Мука", 12, 18, "тонна", 0.5, 1.0, nil, nil},
    6:  {6, 2, "Доски", 10, 15, "тонна", 0.5, 1.0, nil, nil},
    7:  {7, 2, "Металл", 10, 14, "тонна", 0.5, 1.0, nil, nil},
    8:  {8, 3, "Инструменты", 21, 30, "тонна", 0.5, 1.0, nil, nil},
    9:  {9, 3, "Мебель", 22, 26, "тонна", 0.5, 1.0, nil, nil},
    10: {10, 3, "Хлеб", 15, 20, "тонна", 0.5, 1.0, nil, nil},
  }

}

func main() {
  if err := logger.Init("application.log"); err != nil {
    panic(err)
  }

  defer logger.Close()

  logger.Info("Приложение запущено")

  Game = game.Template{
    Pause:      false,
    Step:       0,
    TimeFactor: 8, // 1, 2, 4, 8
    Goods:      Goods,
  }

  Game.NewMap(84, 12)

  Game.GenerateTowns()

  startX, startY := Game.GetRandomTownPosition()

  Game.Caravan = caravan.Template{
    Name:        "Караван",
    Status:      Status["Starting"],
    X:           startX,
    Y:           startY,
    Money:       1000,
    CapacityMax: 500,
    //Target: common.RndRange(0, len(Game.Towns)-1),
    //PrevTarget : -1,
    Config: caravan.Config{
			MinBalance:      100,
      BuyMaxPrice:     0.25, // Покупать если удовлетворено условие:  Цена <= BuyMaxPrice * (PriceMin + (PriceMax - PriceMin))
      BuyFullCapacity: true, // Стараться купить Кол-во равное CapacityMax, если получится, то покупается несколько видов товаров
      BuyMaxAmount:    0.50, // Если BuyFullCapacity == false, то Кол-во покупаемого товара не более чем BuyMaxAmount * CapacityMax
      BuyMinAmount:    0.10, // Минимальное кол-во для покупки BuyMinAmount * CapacityMax
      SellWithProfit:  true, // Всегда продавать по цене большей чем цена покупки
      SellMinPrice:    0.50, // Если SellWithProfit == false, то продавать если Цена >= SellMinPrice * (PriceMin + (PriceMax - PriceMin))
    },
  }

  Game.CaravanSelectDestination()

  Tui = tview.NewApplication()

  textStatus = tview.NewTextView().
    SetDynamicColors(true).
    SetText("Сжатие времени: [green]х1[white]")

  textStatus.
    SetBorder(true).
    SetTitleAlign(tview.AlignLeft).
    SetTitle("Статус")

  Tui.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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
    case 81, 113:
      // qQ - использовать для выхода с сохранением
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

  logger.Info("Создаем караван")

  //Caravan.Target = common.RndRange(1, len(Towns))
  //Caravan.PrevTarget = -1

  //fmt.Println(PrintMap(GlobalMap, Towns, Caravan))
  //fmt.Printf("%+v\n",GlobalMap)

  //  os.Exit(0)

  InitGame()

  go GlobalTick()

  if err := Tui.SetRoot(grid, true).SetFocus(grid).Run(); err != nil {
    panic(err)
  }

  fmt.Printf("%+v\n", Game)
  fmt.Println(Game.TotalVisited)
	//fmt.Printf("%#v\n", town.GoodStatus)
  os.Exit(0)
}
