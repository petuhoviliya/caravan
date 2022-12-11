package game

import (
  "fmt"
  "math"
  "time"
	_ "maps"
	_ "slices"

  "github.com/petuhoviliya/caravan/internal/common"
  "github.com/petuhoviliya/caravan/internal/caravan"
  "github.com/petuhoviliya/caravan/internal/world"
  "github.com/petuhoviliya/caravan/internal/town"
  "github.com/petuhoviliya/caravan/internal/logger"
)

const (
  TownPlaceRadius int = 5
)

type Template struct {
  Pause        bool
  Step         int
  Ticker       *time.Ticker
  TimeFactor   time.Duration
  TotalVisited int

  World   world.Template
  Towns   map[int]town.Template
	Goods   map[int]common.TradingGood
  Caravan caravan.Template
}

func (t *Template) RandomGoods() map[int]town.Goods {
	var NewGoods  map[int]town.Goods

	NewGoods = make(map[int]town.Goods)

	for k,v := range t.Goods {
		NewGoods[k] =	town.Goods{
			Id: v.Id,
			Quantity: int64(common.RndRange(0, int(town.TownWarehouseLimit))),
			PriceMin: v.PriceMin,
			PriceMax: v.PriceMax,
			Status: uint8(common.RndRange(0, len(town.GoodStatus)-1)),
		}
	}

	return NewGoods
}


func (t *Template) GenerateTowns() {

  t.Towns = make(map[int]town.Template)

  for id, name := range common.AlphabetRU {

    if len(t.World.GetFreeCells()) == 0 {
      break
    }

    t.Towns[id] = t.NewTown(id, name)
  }
}

func (t *Template) NewTown(Id int, Name string) town.Template {

  X, Y := t.World.FreeCell()
  t.World.PlaceTown(X, Y, TownPlaceRadius)

  return town.Template{Id, Name, 1, X, Y, town.TownWarehouseLimit, t.RandomGoods(), 0}
}

func (t *Template) NewMap(W, H int) {

  t.World = world.Template{Width: W, Height: H}
  t.World.MakeBitmap()

}

func (t *Template) PrintableMap() string {
  /*

      ║ ╣ ╠ ╬ ╗ ╔ ╝ ╚ ╩ ╦ ═  - borders

      │ ┤ ├ ┼ ┐ ┌ ┘ └ ┴ ┬ ─  - roads

      ┄ ┆ ┈ ┊                - bridges

      ┅ ┇ ┉ ┋                - bridges

      ┃ ┫ ┣ ╋ ┓ ┏ ┛ ┗ ┻ ┳ ━  - main roads

      ╤ ╧ ╢ ╟ - roads out to borders

      @ - caravan

     U+250x   ─   ━   │   ┃   ┄   ┅   ┆   ┇   ┈   ┉   ┊   ┋   ┌   ┍   ┎   ┏
     U+251x   ┐   ┑   ┒   ┓   └   ┕   ┖   ┗   ┘   ┙   ┚   ┛   ├   ┝   ┞   ┟
     U+252x   ┠   ┡   ┢   ┣   ┤   ┥   ┦   ┧   ┨   ┩   ┪   ┫   ┬   ┭   ┮   ┯
     U+253x   ┰   ┱   ┲   ┳   ┴   ┵   ┶   ┷   ┸   ┹   ┺   ┻   ┼   ┽   ┾   ┿
     U+254x   ╀   ╁   ╂   ╃   ╄   ╅   ╆   ╇   ╈   ╉   ╊   ╋   ╌   ╍   ╎   ╏
     U+255x   ═   ║   ╒   ╓   ╔   ╕   ╖   ╗   ╘   ╙   ╚   ╛   ╜   ╝   ╞   ╟
     U+256x   ╠   ╡   ╢   ╣   ╤   ╥   ╦   ╧   ╨   ╩   ╪   ╫   ╬   ╭   ╮   ╯
     U+257x   ╰   ╱   ╲   ╳   ╴   ╵   ╶   ╷   ╸   ╹   ╺   ╻   ╼   ╽   ╾   ╿
  */

  var PrintableMap string
  var ColorTag string

  for posY := -1; posY <= t.World.Height; posY++ {
    for posX := -1; posX <= t.World.Width; posX++ {

      //fmt.Printf("X:%d, Y:%d\n", posX, posY)

      if posY == -1 {
        if posX == -1 { // левый верхний угол
          PrintableMap = PrintableMap + "╔"
        } else if posX == (t.World.Width) { // правый верхний угол
          PrintableMap = PrintableMap + "╗\n"
        } else {
          PrintableMap = PrintableMap + "═" // верхний край
        }
      } else if posY == (t.World.Height) {
        if posX == -1 { // левый нижний угол
          PrintableMap = PrintableMap + "╚"
        } else if posX == (t.World.Width) { // правый нижний угол
          PrintableMap = PrintableMap + "╝\n"
        } else {
          PrintableMap = PrintableMap + "═"
        }
      } else {
        if posX == (t.World.Width) {
          PrintableMap = PrintableMap + "║\n"
        } else if posX == -1 {
          PrintableMap = PrintableMap + "║"
        } else {

          var mapObject string = " "
          for _, town := range t.Towns {
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
              //mapObject = fmt.Sprintf("%s", town.Name[0:2])
            }
          }

          if t.Caravan.X == posX && t.Caravan.Y == posY {
            mapObject = "@"
          }

          PrintableMap = PrintableMap + mapObject
        }
      }
    }
  }
  return PrintableMap
}

func (t *Template) CaravanMoveToTown() {

	var TownId town.Template
    

  t.Caravan.X, t.Caravan.Y = FindBestNextPoint(t.Caravan.X, t.Caravan.Y, t.Towns[t.Caravan.Target].X, t.Towns[t.Caravan.Target].Y)

  if t.Caravan.X == t.Towns[t.Caravan.Target].X && t.Caravan.Y == t.Towns[t.Caravan.Target].Y {

    //TextLog := fmt.Sprintf("[%d]: Прибыл в \"%s\"\n", t.Step, t.Towns[t.Caravan.Target].Name)

    //PrintToGameLog(TextLog)

    t.Caravan.Status = 0
		
		TownId = t.Towns[t.Caravan.Target]

		logger.Debugf("****************\n")

    logger.Debugf("TOWN BEFORE TRADE: %s %+v\n",TownId.Name,TownId.GetGoods())

		t.Caravan.SellForBestPrice(&TownId)
		t.CommitChangesToTown(TownId)

    logger.Debugf("TOWN AFTER SELL: %s: %+v\n", TownId.Name,TownId.GetGoods())
		
		TownId = t.Towns[t.Caravan.Target]

		t.Caravan.BuyForBestPrice(&TownId)
		t.CommitChangesToTown(TownId)

    logger.Debugf("TOWN AFTER BUY: %s: %+v\n",TownId.Name,TownId.GetGoods())


    //logger.Debugf("CARGO: %+v\n", t.Caravan.Cargo)
    
    v := t.Towns[t.Caravan.Target]

    v.Visited++

    t.Towns[t.Caravan.Target] = v
    t.TotalVisited++

    t.CaravanSelectDestination()

  } else {
    t.Caravan.Status = 0
  }
}

func (t *Template) CommitChangesToTown(Town town.Template) {
	var NewTowns map[int]town.Template

	NewTowns = make(map[int]town.Template)

	for k, v := range t.Towns {
		if v.Id == Town.Id {
			NewTowns[k] = Town
		}else{
			NewTowns[k] = v
		}
	}

	t.Towns = NewTowns
}

func (t *Template) CaravanSelectDestination() {

  t.Caravan.PrevTarget = t.Caravan.Target

  for {
    t.Caravan.Target = common.RndRange(0, len(t.Towns)-1)
    if t.Caravan.Target != t.Caravan.PrevTarget {
      break
    }
  }
}

func (t *Template) GetRandomTownId() (id int) {
  return common.RndRange(0, len(t.Towns)-1)
}

func (t *Template) GetTownPositionById(id int) (X, Y int) {
  return t.Towns[id].X, t.Towns[id].Y
}

func (t *Template) GetRandomTownPosition() (X, Y int) {
  id := t.GetRandomTownId()
  return t.GetTownPositionById(id)
}

func PointInsideRadius(X, Y, Radius int) bool {

  A := math.Abs(float64(0 - X))
  B := math.Abs(float64(0 - Y))
  C := int(math.Sqrt(math.Pow(A, 2) + math.Pow(B, 2)))

  return C <= Radius
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
