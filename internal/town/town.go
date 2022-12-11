package town

// Village → Town → City 

import (
	"os"
  _ "github.com/petuhoviliya/caravan/internal/common"
  "github.com/petuhoviliya/caravan/internal/logger"
)

const TownWarehouseLimit int64 = 1000

//var GoodStatus []string

/*
↻ ↺ ↻ ⇅ ⇵ ↑ ↓ ← →
⇅ 0 - продается и покупается
↑ 1 - только продается
↓ 2 - только покупается
↻ 3 - потребляется как ресурс для производства, только покупается
↺ 4 - производимый товар, только продается
*/

var GoodStatus = []string{0:"⇅",1:"↑",2:"↓",3:"↻",4:"↺"}

type Goods struct {
  Id       int
  Quantity int64
	PriceMin float64
	PriceMax float64
	Status   uint8
}

type Template struct {
  Id             int
  Name           string
  Tier           int
  X              int
  Y              int
  WarehouseLimit int64
  Goods          map[int]Goods
  Visited        int
}


func (t *Template) CheckUpgrade() bool {
	return false
}

func (t *Template) DoUpgrade() {}


func (t *Template) FillWares() {
  //t.Wares = make(map[int]Wares)
}

func (t *Template) GetGoods() map[int]Goods {
	return t.Goods
}

func (t *Template) AddGoods(Id int, Quantity int64) {
	var NewGoods map[int]Goods

	NewGoods=make(map[int]Goods)
	
	for k,v	:= range t.Goods {
		if k == Id {
			NewGoods[k] = Goods {
  			Id: k,
			  Quantity: v.Quantity + Quantity,
				PriceMin: v.PriceMin,
				PriceMax: v.PriceMax,
			}
		}else{
			NewGoods[k] = v
		}
	}
	t.Goods = NewGoods
}

func (t *Template) RemoveGoods(Id int, Quantity int64) {
	var NewGoods map[int]Goods
	
	NewGoods=make(map[int]Goods)

	for k,v	:= range t.Goods {
		if k == Id {
			NewGoods[k] = Goods {
  			Id: k,
			  Quantity: v.Quantity - Quantity,
				PriceMin: v.PriceMin,
				PriceMax: v.PriceMax,
			}
		}else{
			NewGoods[k] = v
		}
	}
	t.Goods = NewGoods
}


func (t *Template) GetGoodQuantity(Id int) int64 {
	return t.Goods[Id].Quantity
}

func (t *Template) GetGoodPrice(Id int) float64 {

	if t.GetGoodQuantity(Id) == t.WarehouseLimit {
		return t.Goods[Id].PriceMin
	}
	if t.GetGoodQuantity(Id) == 0 {
		return t.Goods[Id].PriceMax
	}

	Q := float64(1.0 - float64(t.GetGoodQuantity(Id))/float64(t.WarehouseLimit)) // коэффицент заполненности склада
	D := float64(t.Goods[Id].PriceMax - t.Goods[Id].PriceMin)                    // абсолютная разница цены на товар
	Price := float64(t.Goods[Id].PriceMin) + D * Q

	if Price < 0 {
		logger.Debugf("TOWN: %+v\n", t )
		logger.Debugf("GOOD PRICE: Qnty:%d/%d Q:%f D:%f Price:%f\n",
			t.GetGoodQuantity(Id),
			t.WarehouseLimit,
			Q,
			D,
			Price,
		)
		os.Exit(1)
	}

	return Price
}

