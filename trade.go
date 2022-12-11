package main

import (
  "fmt"
  "log"
  "maps"
  "math"
  "math/rand"
  "os"
  "slices"
  "sort"
  "text/tabwriter"
  "time"
)

const TickerInterval = 1000 * time.Millisecond
const TownWarehouseLimit = 100

type GameTemplate struct {
  Pause        bool
  Step         int
  Ticker       *time.Ticker
  TimeFactor   time.Duration
  TotalVisited int
  Goods        map[int]TradingGood
  Caravan      CaravanTemplate
  Town         TownTemplate
}

type CaravanTemplate struct {
  Name        string
  Status      uint8
  Money       int64
  X           int
  Y           int
  Target      int
  PrevTarget  int
  CapacityMax int64
  Cargo       []Cargo
  TradeConfig TradeConfig
}

type TownTemplate struct {
  Id             int
  Name           string
  Tier           int
  X              int
  Y              int
  WarehouseLimit int64
  Wares          map[int]WareGood
  Visited        int
}

type Cargo struct {
  WareId   int
  TownId   int
  Quantity int64
  BuyPrice float64
}

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

type WareGood struct {
  Id       int
  Quantity int64
}

type SortedWare struct {
  Key   int
  Value WareGood
}

type ShoppingListItemTemplate struct {
  Quantity int64
  Price    float64
}

type ShoppingListTemplate struct {
  Items map[int]ShoppingListItemTemplate
}

func (b *ShoppingListTemplate) GetTotalQuantity() int64 {
  var totalQuantity int64

  for _, v := range b.Items {
    totalQuantity += v.Quantity
  }

  return totalQuantity
}

func (b *ShoppingListTemplate) GetTotalPrice() float64 {
  var totalPrice float64

  for _, v := range b.Items {
    totalPrice += float64(v.Quantity) * v.Price
  }

  return totalPrice
}

func (b *ShoppingListTemplate) AddItem(Id int, Quantity int64, Price float64) {
  b.Items[Id] = ShoppingListItemTemplate{
    Quantity: Quantity,
    Price:    Price,
  }
}

type TradeConfig struct {

  // Минимальный остаток денег
  MinBalance float64

  // Минимальный остаток денег в процентах
  MinBalancePercent float64

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

  // Максимальный стоимость одной сделки по покупке
  // в процентах от баланса каравана на момент перед сделкой
  BuyOrderTotalPrice float64

  // Всегда продавать с прибылью
  // Цена продажи не может быть ниже цены покупки
  SellWithProfit bool

  // Минимальный процент прибыли, если SellWithProfit == true
  // Может быть 0
  SellWithProfitProc float64

  // Минимальный процент прибыли с которой продавать
  // в процентах
  //SellMinProfit float64

  // Минимальная цена продажи товара
  // в процентах
  // TradingGood.PriceMin - 0%
  // TradingGood.PriceMax - 100%
  SellMinPrice float64

  // Всегда оставлять указанный процент грузоподъемности свободным
  // в процентах
  CargoReservedSpace float64

  // Резервировать указанное значение грузоподъемности под еду
  CargoReserverdForFood int
}

var (
  Game         GameTemplate
  Caravan      CaravanTemplate
  ShoppingList ShoppingListTemplate

  BuyerTown  TownTemplate
  SellerTown TownTemplate
)

func init() {

  log.Println("Init")
  var rSeed int64 = time.Now().UnixNano()

  // rSeed = 1757754895702173829
  rand.New(rand.NewSource(rSeed))
  log.Printf("Seed: %d\n", rSeed)

}

func main() {

  InitGame()
  //fmt.Printf("%+v\n", Game.Town)
  Game.Caravan.DoTrade()
  os.Exit(0)

  GlobalTick()

  fmt.Println(Game.Step)
  os.Exit(0)
}

func GlobalTick() {

  for {
    select {
    case <-Game.Ticker.C:

      Game.Step++
      //Game.Town = RandomTown()
      log.Println("Ход", Game.Step)
      //fmt.Printf("%+v\n", Game.Town)
      Game.Caravan.DoTrade()
      //os.Exit(0)
    }
  }
}

func InitGame() {

  Game = GameTemplate{
    Pause:      false,
    Step:       0,
    TimeFactor: 5, // 1, 2, 4, 8
  }

  Game.Ticker = time.NewTicker(TickerInterval * Game.TimeFactor)
  Game.Ticker.Reset(TickerInterval * Game.TimeFactor)

  Game.Goods = map[int]TradingGood{
    // Id   Tier  Name      PriceMin  PriceMax  Unit        Volume  Weight  Resources  Consumables
    1:  {1, 1, "Зерно", 2, 10, "мешок", 0.036, 0.050, nil, nil},
    2:  {2, 1, "Дерево", 5, 20, "кубометр", 1.0, 0.640, nil, nil},
    3:  {3, 1, "Камень", 4, 18, "кубометр", 1.0, 1.7, nil, nil},
    4:  {4, 1, "Руда", 9, 30, "тонна", 0.5, 1.0, nil, nil},
    5:  {5, 2, "Мука", 22, 110, "тонна", 0.5, 1.0, nil, nil},
    6:  {6, 2, "Доски", 25, 120, "тонна", 0.5, 1.0, nil, nil},
    7:  {7, 2, "Металл", 29, 130, "тонна", 0.5, 1.0, nil, nil},
    8:  {8, 3, "Инструменты", 129, 230, "тонна", 0.5, 1.0, nil, nil},
    9:  {9, 3, "Мебель", 125, 220, "тонна", 0.5, 1.0, nil, nil},
    10: {10, 3, "Хлеб", 122, 210, "тонна", 0.5, 1.0, nil, nil},
  }
  //fmt.Printf("\n\n%+v\n\n",Game.Goods)

  /*  Game.Town = TownTemplate {
      Id:   1,
      Name:  "Энск",
      Tier:  1,
      X:    0,
      Y:    0,
      WarehouseLimit: 100.0,
      Wares: map[int]WareGood{
        1:{
          Id: 4,
          Quantity: 75,
        },
      },
    }*/

  Game.Caravan = CaravanTemplate{
    Name:        "Караван",
    Money:       1000,
    CapacityMax: 100,
    TradeConfig: TradeConfig{
      MinBalance:         50.0, // После любых торговых операций должна остаться эта сумма
      BuyMaxPrice:        0.25, // Покупать если удовлетворено условие:  Цена <= BuyMaxPrice * (PriceMin + (PriceMax - PriceMin))
      BuyFullCapacity:    true, // Стараться купить Кол-во равное CapacityMax, если получится, то покупается несколько видов товаров
      BuyMaxAmount:       0.50, // Если BuyFullCapacity == false, то Кол-во покупаемого товара за раз не более чем BuyMaxAmount * CapacityMax
      BuyMinAmount:       0.10, // Минимальное кол-во для покупки BuyMinAmount * CapacityMax
      BuyOrderTotalPrice: 0.90, // Общая сумма одной сделки по покупке не может превышать BuyOrderTotalPrice * Баланс, т.е. не более 90% от текущего баланса
      SellWithProfit:     true, // Всегда продавать по цене большей чем цена покупки
      SellWithProfitProc: 0.05, // Минимальный процент прибыли при продаже, если SellWithProfit == true
      SellMinPrice:       0.50, // Если SellWithProfit == false, то продавать если Цена >= SellMinPrice * (PriceMin + (PriceMax - PriceMin))
    },
  }

  log.Println("Game.Goods:", len(Game.Goods))

  SellerTown = TownTemplate{
    Id:             1,
    Name:           "Продаванск",
    Tier:           1,
    X:              0,
    Y:              0,
    WarehouseLimit: TownWarehouseLimit,
    Wares:          RandomWares(5),
  }

  BuyerTown = TownTemplate{
    Id:             2,
    Name:           "Покупанск",
    Tier:           2,
    X:              0,
    Y:              0,
    WarehouseLimit: TownWarehouseLimit * 2,
    Wares:          RandomWares(5),
  }

}

/*
Common functions
*/
func RndRange(min, max int) int {
  return rand.Intn(max-min+1) + min
}

func Rnd(max int) int {
  return RndRange(1, max)
}

func SortWaresByQuantityAsc(wares map[int]WareGood) []SortedWare {

  var sorted []SortedWare

  for k, v := range wares {
    sorted = append(sorted, SortedWare{k, v})
  }

  sort.Slice(sorted, func(i, j int) bool {
    return sorted[i].Value.Quantity < sorted[j].Value.Quantity
  })

  return sorted
}

func SortWaresByQuantityDesc(wares map[int]WareGood) []SortedWare {

  var sorted []SortedWare

  for k, v := range wares {
    sorted = append(sorted, SortedWare{k, v})
  }

  sort.Slice(sorted, func(i, j int) bool {
    return sorted[i].Value.Quantity > sorted[j].Value.Quantity
  })

  return sorted
}

func RandomWares(Minimal int) map[int]WareGood {
  var (
    wares    map[int]WareGood
    goods    map[int]TradingGood
    keys     []int
    rKey     int
    rId      int
    count    int
    quantity int64
  )

  wares = make(map[int]WareGood)

  goods = maps.Clone(Game.Goods)        // локальная копия списка товаров
  count = RndRange(Minimal, len(goods)) // сколько товаров добавляем в город

  //fmt.Printf("%+v\n\n",goods)

  for i := 1; i <= count; i++ {
    keys = slices.Collect(maps.Keys(goods)) // список ключей
    slices.Sort(keys)
    //fmt.Printf("%+v\n",keys)
    rKey = Rnd(len(keys))
    //fmt.Printf("rKey: %+v\n",rKey)
    rId = keys[rKey-1]
    //fmt.Printf("rId: %+v\n",rId)
    quantity = int64(RndRange(1, TownWarehouseLimit))
    wares[rId] = WareGood{Id: goods[rId].Id, Quantity: quantity}
    delete(goods, rId)
  }
  clear(goods)

  //fmt.Printf("%+v\n\n",wares)

  return wares
}

func DoActions() {}

/*
  CaravanTemplate
*/

func (c *CaravanTemplate) CargoCapacity() int64 {
  var Capacity int64

  if len(c.Cargo) == 0 {
    return 0
  }

  for _, cargo := range c.Cargo {
    Capacity += cargo.Quantity
  }
  return Capacity
}

func (c *CaravanTemplate) DoTrade() {

  w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)

  c.BuyForBestPrice(&SellerTown)

  fmt.Println("")
  fmt.Println("***")
  fmt.Println("")

  fmt.Printf("Груз: \033[1m%s\033[0m\n", "")
  fmt.Fprintf(w, "\033[1mТовар\tКол-во\tЦена покупки\tГород покупки\t\n")

  for _, v := range c.Cargo {
    fmt.Fprintf(w, "\033[0m%s\t%d\t%.2f\t%d\t\n",
      Game.Goods[v.WareId].Name,
      v.Quantity,
      v.BuyPrice,
      v.TownId,
    )
  }

  w.Flush()

  fmt.Println("")
  fmt.Println("***")
  fmt.Println("")

  c.SellForBestPrice(&BuyerTown)

  fmt.Println("")
  fmt.Println("***")
  fmt.Println("")

  fmt.Printf("%+v\n", c)

  fmt.Printf("Груз: \033[1m%s\033[0m\n", "")
  fmt.Fprintf(w, "\033[1mТовар\tКол-во\tЦена покупки\tГород покупки\t\n")

  for _, v := range c.Cargo {
    fmt.Fprintf(w, "\033[0m%s\t%d\t%.2f\t%d\t\n",
      Game.Goods[v.WareId].Name,
      v.Quantity,
      v.BuyPrice,
      v.TownId,
    )
  }
  w.Flush()
}

func (c *CaravanTemplate) AddMoney(Quantity int64) bool {
  c.Money += Quantity
  return true
}

func (c *CaravanTemplate) RemoveMoney(Quantity int64) bool {
  if c.Money >= Quantity {
    c.Money -= Quantity
    return true
  }
  return false
}

func (c *CaravanTemplate) AddToCargo(Cargo Cargo) {
  c.Cargo = append(c.Cargo, Cargo)
}

func (c *CaravanTemplate) RemoveFromCargo(Id int, Quantity int64) {
  for k, v := range c.Cargo {
    if v.WareId == Id {
      if v.Quantity == Quantity {
        c.DeleteFromCargo(Id)
      } else {
        c.Cargo[k].Quantity -= Quantity
      }
    }
  }
}

func (c *CaravanTemplate) DeleteFromCargo(Id int) {
  var NewCargo []Cargo

  for _, v := range c.Cargo {
    if v.WareId != Id {
      NewCargo = append(NewCargo, v)
    }
  }

  c.Cargo = NewCargo
}

func (c *CaravanTemplate) SellToTown(Town *TownTemplate, List ShoppingListTemplate) {

  for k, v := range List.Items {
    c.RemoveFromCargo(k, v.Quantity)
    c.AddMoney(int64(float64(v.Quantity) * v.Price))
  }

}

func (c *CaravanTemplate) BuyFromTown(Town *TownTemplate, List ShoppingListTemplate) {
  /*
     WareId   int
     TownId   int
     Quantity int64
     BuyPrice float64
  */
  for k, v := range List.Items {
    c.AddToCargo(Cargo{
      WareId:   k,
      TownId:   Town.Id,
      Quantity: v.Quantity,
      BuyPrice: v.Price,
    })
    c.RemoveMoney(int64(float64(v.Quantity) * v.Price))
  }

}

func (c *CaravanTemplate) SellForBestPrice(Town *TownTemplate) {
  //if c.CargoCapacity() == int64(0) {
  //  return
  //}

  var (
    Price          float64
    SellPrice      float64
    SellAmount     int64
    SellAmountMax  int64
    SellIntent     string
    TotalPrice     float64
    RelativePrice  float64
    BuyingPrice    float64
    ProfitIntent   string
    Profit         float64
    ProfitRelative float64
    sorted         []SortedWare
  )

  sorted = SortWaresByQuantityAsc(Town.Wares)

  //fmt.Printf("\n\n%+v\n\n", sorted)

  ShoppingList.Items = make(map[int]ShoppingListItemTemplate)
  w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
  //tabwriter.AlignRight|tabwriter.Debug

  fmt.Printf("Смотрим что можем продать в: \033[1m%s\033[0m\n", Town.Name)
  fmt.Fprintf(w, "\033[1mТовар\tКол-во\tЦена (%%)\tМин/Макс\tМин цена продажи\tК продаже\tПрибыль (%%)\t\n")

  for _, v := range sorted {

    Q := float64(1.0 - float64(v.Value.Quantity)/float64(Town.WarehouseLimit))      // коэффицент заполненности склада
    D := float64(Game.Goods[v.Value.Id].PriceMax - Game.Goods[v.Value.Id].PriceMin) // абсолютная разница цены на товар
    Qb := Game.Caravan.TradeConfig.SellMinPrice * D                                 // коэфициент для закупки

    Price = float64(Game.Goods[v.Value.Id].PriceMin) + D*Q
    SellPrice = float64(Game.Goods[v.Value.Id].PriceMin) + Qb
    RelativePrice = (Price - float64(Game.Goods[v.Value.Id].PriceMin)) / D * 100

    TownFreeCapacity := Town.WarehouseLimit - v.Value.Quantity

    SellIntent = "-"
    ProfitIntent = "-"
    SellAmount = 0
    BuyingPrice = 0

    for _, value := range c.Cargo {
      if value.WareId == v.Value.Id {
        BuyingPrice = value.BuyPrice

        if TownFreeCapacity >= value.Quantity {
          SellAmountMax = value.Quantity
        } else {
          SellAmountMax = TownFreeCapacity
        }

        if SellAmount > SellAmountMax {
          SellAmount = SellAmountMax
        } else {
          SellAmount = value.Quantity
        }

        break
      }
    }

    TotalPrice = float64(SellAmount) * Price

    if Price > SellPrice && SellAmount != 0 {

      Profit = Price - BuyingPrice
      ProfitRelative = BuyingPrice / Price * 100

      ShoppingList.AddItem(v.Value.Id, SellAmount, Price)
      SellIntent = fmt.Sprintf("%d за %.2f",
        SellAmount,
        TotalPrice)
      ProfitIntent = fmt.Sprintf("%.2f (%.2f%%)",
        Profit,
        ProfitRelative,
      )
    }
    fmt.Fprintf(w, "\033[0m%s (%d)\t%d/%d\t%.2f (%0.f%%)\t%.2f / %.2f\t%.2f\t%s\t%s\t\n",
      Game.Goods[v.Value.Id].Name,
      v.Value.Id,
      v.Value.Quantity,
      Town.WarehouseLimit,
      Price,
      RelativePrice,
      float64(Game.Goods[v.Value.Id].PriceMin),
      float64(Game.Goods[v.Value.Id].PriceMax),
      SellPrice,
      SellIntent,
      ProfitIntent,
    )

  }
  w.Flush()

  c.SellToTown(Town, ShoppingList)
  fmt.Printf("\n %+v\n%d единиц товара за %.2f \n", ShoppingList, ShoppingList.GetTotalQuantity(), ShoppingList.GetTotalPrice())
}

func (c *CaravanTemplate) BuyForBestPrice(Town *TownTemplate) {

  /*
     Порядок проверок закупа
     Цена
     Остаточная емкость каравана
       Сколько покупаем
     Хватит ли денег
  */

  var (
    Price         float64
    BuyPrice      float64
    BuyAmount     int64
    BuyAmountMax  int64
    BuyIntent     string
    TotalPrice    float64
    RelativePrice float64
    sorted        []SortedWare
  )

  sorted = SortWaresByQuantityDesc(Town.Wares)
  //fmt.Printf("\n\n%+v\n\n", sorted)

  ShoppingList.Items = make(map[int]ShoppingListItemTemplate)
  w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
  //tabwriter.AlignRight|tabwriter.Debug

  /*
     MinBalance:         50.0, // После любых торговых операций должна остаться эта сумма
     BuyMaxPrice:        0.25, // Покупать если удовлетворено условие:  Цена <= BuyMaxPrice * (PriceMin + (PriceMax - PriceMin))
     BuyFullCapacity:    true, // Стараться купить Кол-во равное CapacityMax, если получится, то покупается несколько видов товаров
     BuyMaxAmount:       0.50, // Если BuyFullCapacity == false, то Кол-во покупаемого товара за раз не более чем BuyMaxAmount * CapacityMax
     BuyMinAmount:       0.10, // Минимальное кол-во для покупки BuyMinAmount * CapacityMax
     BuyOrderTotalPrice: 0.90, // Общая сумма одной сделки по покупке не может превышать BuyOrderTotalPrice * Баланс, т.е. не более 90% от текущего баланса
     SellWithProfit:     true, // Всегда продавать по цене большей чем цена покупки
     SellWithProfitProc: 0.05, // Минимальный процент прибыли при продаже, если SellWithProfit == true
     SellMinPrice:       0.50, // Если SellWithProfit == false, то продавать если Цена >= SellMinPrice * (PriceMin + (PriceMax - PriceMin))
  */

  //fmt.Printf("%+v\n\n", Game.Town.Wares)

  fmt.Printf("Смотрим что можем купить в: \033[1m%s\033[0m\n", Town.Name)
  fmt.Fprintf(w, "\033[1mТовар\tКол-во\tЦена (%%)\tМин/Макс\tМакс закупочная цена\tК закупке\t\n")

  for _, v := range sorted {

    Q := float64(1.0 - float64(v.Value.Quantity)/float64(Town.WarehouseLimit))      // коэффицент заполненности склада
    D := float64(Game.Goods[v.Value.Id].PriceMax - Game.Goods[v.Value.Id].PriceMin) // абсолютная разница цены на товар
    Qb := Game.Caravan.TradeConfig.BuyMaxPrice * D                                  // коэфициент для закупки

    Price = float64(Game.Goods[v.Value.Id].PriceMin) + D*Q
    BuyPrice = float64(Game.Goods[v.Value.Id].PriceMin) + Qb
    RelativePrice = (Price - float64(Game.Goods[v.Value.Id].PriceMin)) / D * 100

    FreeCapacity := Game.Caravan.CapacityMax - Game.Caravan.CargoCapacity() - ShoppingList.GetTotalQuantity()

    if FreeCapacity >= v.Value.Quantity {
      BuyAmountMax = v.Value.Quantity
    } else {
      BuyAmountMax = FreeCapacity
    }

    TotalPriceMax := float64(Game.Caravan.Money) - Game.Caravan.TradeConfig.MinBalance - ShoppingList.GetTotalPrice()
    BuyAmount = int64(math.Floor(TotalPriceMax / Price))
    ShoppingList.GetTotalQuantity()

    if BuyAmount > BuyAmountMax {
      BuyAmount = BuyAmountMax
    }

    //if ShoppingList.GetTotalPrice() > 0 {
    //  BuyAmount = math.Floor(ShoppingList.GetTotalPrice())
    //}

    TotalPrice = float64(BuyAmount) * Price

    //fmt.Println(ShoppingList.GetTotalPrice(),TotalPrice)

    BuyIntent = "-"

    if Price < BuyPrice && BuyAmount != 0 {

      ShoppingList.AddItem(v.Value.Id, BuyAmount, Price)
      BuyIntent = fmt.Sprintf("%d за %.2f\t", BuyAmount, TotalPrice)
    }

    fmt.Fprintf(w, "\033[0m%s (%d)\t%d/%d\t%.2f (%0.f%%)\t%.2f / %.2f\t%.2f\t%s\033[0m\t\n",
      Game.Goods[v.Value.Id].Name,
      v.Value.Id,
      v.Value.Quantity,
      TownWarehouseLimit,
      Price,
      RelativePrice,
      float64(Game.Goods[v.Value.Id].PriceMin),
      float64(Game.Goods[v.Value.Id].PriceMax),
      BuyPrice,
      BuyIntent,
    )

  }
  w.Flush()

  c.BuyFromTown(Town, ShoppingList)

  fmt.Printf("\n %+v\n%d единиц товара за %.2f \n", ShoppingList, ShoppingList.GetTotalQuantity(), ShoppingList.GetTotalPrice())

  //fmt.Printf("\n%+v\n",c)

}
