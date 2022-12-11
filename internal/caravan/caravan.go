package caravan

import (
  
	_ "fmt"
  "math"
  "sort"
	"os"

  _ "github.com/petuhoviliya/caravan/internal/common"
  "github.com/petuhoviliya/caravan/internal/town"
  "github.com/petuhoviliya/caravan/internal/logger"
)


type Template struct {
  Name        string
  Status      uint8
  Money       int64
  X           int
  Y           int
  Target      int
  PrevTarget  int
  CapacityMax int64
  Cargo       []Cargo
  Config Config
}

type Cargo struct {
  GoodId   int
  TownId   int
  Quantity int64
  BuyPrice float64
	BuyDate  int64
}

type Config struct {

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

	// Продавать в первую очередь товар, который давно куплен
	CargoSoldTooOld bool

	// Возраст товара для первоочередной продажи, если CargoSoldTooOld == true
	CargoOldAge int

	// Возраст товара от которого надо просто избавится по любой цене, если CargoSoldTooOld == true
	CargoSuperOldAge int
	
	// Сливать одинаковые товары в одну позицию
	CargoMergeSame bool
	
	// Метод сливания цены покупки
  // 1 - среднее
  // 2 - по низшей
  // 3 - по высшей
	CargoMergePrice uint8
}


type SortedWare struct {
  Key   int
  Value town.Goods
}

type ShoppingListItemTemplate struct {
  Quantity int64
  Price    float64
}

type ShoppingListTemplate struct {
  Items map[int]ShoppingListItemTemplate
}

func (s *ShoppingListTemplate) GetTotalQuantity() int64 {
  var totalQuantity int64

  for _, v := range s.Items {
    totalQuantity += v.Quantity
  }

  return totalQuantity
}

func (s *ShoppingListTemplate) GetTotalPrice() float64 {
  var totalPrice float64

  for _, v := range s.Items {
    totalPrice += float64(v.Quantity) * v.Price
  }

  return totalPrice
}

func (s *ShoppingListTemplate) AddItem(Id int, Quantity int64, Price float64) {
  s.Items[Id] = ShoppingListItemTemplate{
    Quantity: Quantity,
    Price:    Price,
  }
}


var (
  ShoppingList ShoppingListTemplate
)

func (t *Template) Move(X, Y int) {}

func (t *Template) MoveBest(X, Y int) {}

func (t *Template) ChooseDestination() {}

func (t *Template) PayTaxes(Step int) {
	var taxAmount int = 10
	if Step % 7 == 0 {
		t.RemoveMoney(int64(taxAmount))
		logger.Debugf("PAY TAXES: Step:%d, Amount:%d\n", Step, taxAmount)
	}

}

func (t *Template) CargoCapacity() int64 {
  var Capacity int64

  if len(t.Cargo) == 0 {
    return 0
  }

  for _, cargo := range t.Cargo {
    Capacity += cargo.Quantity
  }
  return Capacity
}

func (t *Template) Sell(TownId, CargoId int) {}

func (t *Template) Buy(TownId, CargoId int) {}


func SortWaresByQuantityAsc(wares map[int]town.Goods) []SortedWare {

  var sorted []SortedWare

  for k, v := range wares {
    sorted = append(sorted, SortedWare{k, v})
  }

  sort.Slice(sorted, func(i, j int) bool {
    return sorted[i].Value.Quantity < sorted[j].Value.Quantity
  })

  return sorted
}

func SortWaresByQuantityDesc(wares map[int]town.Goods) []SortedWare {

  var sorted []SortedWare

  for k, v := range wares {
    sorted = append(sorted, SortedWare{k, v})
  }

  sort.Slice(sorted, func(i, j int) bool {
    return sorted[i].Value.Quantity > sorted[j].Value.Quantity
  })

  return sorted
}

func (t *Template) DoTrade(Town *town.Template) {

  t.SellForBestPrice(Town)
  t.BuyForBestPrice(Town)
}

func (t *Template) AddMoney(Quantity int64) bool {
  t.Money += Quantity
  return true
}

func (t *Template) RemoveMoney(Quantity int64) bool {
  if t.Money >= Quantity {
    t.Money -= Quantity
    return true
  }
  return false
}

func (t *Template) AddToCargo(Cargo Cargo) {
	
	logger.Debugf("ADD: %+v\n", Cargo)
	
	if t.CargoCapacity() == 0 {
		t.Cargo = append(t.Cargo, Cargo)
		return
	}

	for k, v := range t.Cargo {
		if v.GoodId == Cargo.GoodId && v.BuyPrice == Cargo.BuyPrice {
			t.Cargo[k].Quantity += Cargo.Quantity
			break
		}else{
  		t.Cargo = append(t.Cargo, Cargo)
			break
		}
	}

	t.OrganizeCargo()
}

func (t *Template) RemoveFromCargo(Id int, Quantity int64) {

	logger.Debugf("REMOVE: Id:%d, Quantity: %d\n",Id, Quantity)

  for k, v := range t.Cargo {
    if v.GoodId == Id {
    	t.Cargo[k].Quantity -= Quantity
			break
    }
  }

	t.CleanUpCargo()
  
}

func (t *Template) CleanUpCargo() {
  var NewCargo []Cargo
  
	for _, v := range t.Cargo {
    if v.Quantity != 0 {
      NewCargo = append(NewCargo, v)
    }
  }

  t.Cargo = NewCargo
}

func (t *Template) OrganizeCargo() {
/*
  GoodId   int
  TownId   int
  Quantity int64
  BuyPrice float64
	BuyDate  int64
*/
	type aggregatedData struct {
		totalCount   int64
		totalPrice   float64
		countPrices  int
		firstCity    int
	}

	var NewCargo []Cargo

	aggregated := make(map[int]*aggregatedData)

	for _, v := range t.Cargo {
		if data, exists := aggregated[v.GoodId]; exists {
			// Если ID уже есть, суммируем количество и цену
			data.totalCount += v.Quantity
			data.totalPrice += v.BuyPrice
			data.countPrices++
		} else {
			// Если ID новый, создаем запись с первым городом
			aggregated[v.GoodId] = &aggregatedData{
				totalCount:   v.Quantity,
				totalPrice:   v.BuyPrice,
				countPrices:  1,
				firstCity:    v.TownId,
			}
		}
	}

	for k, data := range aggregated {
		avgPrice := data.totalPrice/float64(data.countPrices)
		NewCargo = append(NewCargo, Cargo{
  		GoodId: k,
		  TownId: data.firstCity,
		  Quantity: data.totalCount,
		  BuyPrice: avgPrice,
		})
	}
	t.Cargo = NewCargo
}


func (t *Template) DeleteFromCargo(Id int) {
  var NewCargo []Cargo

  for _, v := range t.Cargo {
    if v.GoodId != Id {
      NewCargo = append(NewCargo, v)
    }
  }

  t.Cargo = NewCargo
}

func (t *Template) SellToTown(Town *town.Template, List ShoppingListTemplate) {

	logger.Debugf("SELL: %+v\n", List)

  for k, v := range List.Items {
    t.RemoveFromCargo(k, v.Quantity)
    t.AddMoney(int64(float64(v.Quantity) * v.Price))
		Town.AddGoods(k, v.Quantity)
  }

}

func (t *Template) BuyFromTown(Town *town.Template, List ShoppingListTemplate) {
	
	logger.Debugf("BUY: %+v\n", List)

  for k, v := range List.Items {
    t.AddToCargo(Cargo{
      GoodId:   k,
      TownId:   Town.Id,
      Quantity: v.Quantity,
      BuyPrice: v.Price,
    })
    t.RemoveMoney(int64(float64(v.Quantity) * v.Price))
		Town.RemoveGoods(k, v.Quantity)
  }

}

func (t *Template) SellForBestPrice(Town *town.Template) {
  //if c.CargoCapacity() == int64(0) {
  //  return
  //}

  var (
    Price          float64
    SellPrice      float64
    SellAmount     int64
    sorted         []SortedWare
  )

  sorted = SortWaresByQuantityAsc(Town.Goods)

  //fmt.Printf("\n\n%+v\n\n", sorted)

  ShoppingList.Items = make(map[int]ShoppingListItemTemplate)

	//logger.Debugf("SORTED: %+v\n", sorted)

  for _, v := range sorted {

    Q := float64(1.0 - float64(v.Value.Quantity)/float64(Town.WarehouseLimit))      // коэффицент заполненности склада
    D := float64(Town.Goods[v.Value.Id].PriceMax - Town.Goods[v.Value.Id].PriceMin) // абсолютная разница цены на товар
    Qb := t.Config.SellMinPrice * D                                 // коэфициент для закупки

    Price = float64(Town.Goods[v.Value.Id].PriceMin) + D*Q
    SellPrice = float64(Town.Goods[v.Value.Id].PriceMin) + Qb

    TownFreeCapacity := Town.WarehouseLimit - v.Value.Quantity
		
		//logger.Debugf("TownFreeCapacity: %+v\n", TownFreeCapacity)

    SellAmount = 0

    for _, value := range t.Cargo {
      if value.GoodId == v.Value.Id {

        if TownFreeCapacity >= value.Quantity {
          SellAmount = value.Quantity
        } else {
          SellAmount = TownFreeCapacity
        }

        break
      }
    }

    if Price > SellPrice && SellAmount != 0 {
      ShoppingList.AddItem(v.Value.Id, SellAmount, Price)
    }
  }

  t.SellToTown(Town, ShoppingList)
  //fmt.Printf("\n %+v\n%d единиц товара за %.2f \n", ShoppingList, ShoppingList.GetTotalQuantity(), ShoppingList.GetTotalPrice())
}

func (t *Template) BuyForBestPrice(Town *town.Template) {

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
    sorted        []SortedWare
  )

  sorted = SortWaresByQuantityDesc(Town.Goods)
  //fmt.Printf("\n\n%+v\n\n", sorted)

  ShoppingList.Items = make(map[int]ShoppingListItemTemplate)
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


  for _, v := range sorted {

    //Q := float64(1.0 - float64(v.Value.Quantity)/float64(Town.WarehouseLimit))      // коэффицент заполненности склада
    D := float64(Town.Goods[v.Value.Id].PriceMax - Town.Goods[v.Value.Id].PriceMin) // абсолютная разница цены на товар
    Qb := t.Config.BuyMaxPrice * D                                  // коэфициент для закупки

    Price = Town.GetGoodPrice(v.Value.Id)
		//float64(Town.Goods[v.Value.Id].PriceMin) + D*Q
    BuyPrice = float64(Town.Goods[v.Value.Id].PriceMin) + Qb


    FreeCapacity := t.CapacityMax - t.CargoCapacity() - ShoppingList.GetTotalQuantity()
		
		/*logger.Debugf("SHOPPING LIST: %+v\n", ShoppingList)

		logger.Debugf("CAPACITY: Free:%d Max:%d Cargo:%d List:%d\n",
			FreeCapacity,
			t.CapacityMax,
			t.CargoCapacity(),
			ShoppingList.GetTotalQuantity(),
		)*/

    if FreeCapacity >= v.Value.Quantity {
      BuyAmountMax = v.Value.Quantity
    } else {
      BuyAmountMax = FreeCapacity
    }

    TotalPriceMax := float64(t.Money) - t.Config.MinBalance - ShoppingList.GetTotalPrice()
		if TotalPriceMax < 0 {
			TotalPriceMax = 0.0
		}
    BuyAmount = int64(math.Floor(TotalPriceMax / Price))
		
		if BuyAmount <0 {
			logger.Debugf("AMOUNT: Money:%d MinBalance:%.2f ShoppingList.GetTotalPrice:%.4f TotalPriceMax:%.4f Price:%.4f Amount: %d\n",
				t.Money,
				t.Config.MinBalance,
				ShoppingList.GetTotalPrice(),
				TotalPriceMax,
				Price,
				BuyAmount,
			)
			os.Exit(1)

		}
		
    if BuyAmount > BuyAmountMax {
      BuyAmount = BuyAmountMax
    }

    //if ShoppingList.GetTotalPrice() > 0 {
    //  BuyAmount = math.Floor(ShoppingList.GetTotalPrice())
    //}


    //fmt.Println(ShoppingList.GetTotalPrice(),TotalPrice)


    if Price < BuyPrice && BuyAmount != 0 {

      ShoppingList.AddItem(v.Value.Id, BuyAmount, Price)
    }


  }

  t.BuyFromTown(Town, ShoppingList)

  //fmt.Printf("\n %+v\n%d единиц товара за %.2f \n", ShoppingList, ShoppingList.GetTotalQuantity(), ShoppingList.GetTotalPrice())

  //fmt.Printf("\n%+v\n",c)

}
