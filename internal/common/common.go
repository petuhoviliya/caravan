package common

import (
  "math/rand"
  "math"
  )

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

type Manufacturer struct {
	Id int
	Tier int
	Name string
}

var AlphabetRU = []string{
    "Амурск", "Биробиджан", "Владивосток", "Грозный",
    "Дубна", "Ейск", "Жуковский", "Зеленоград",
    "Иркутск", "Казань", "Липецк", "Мурманск",
    "Ноглики", "Омск", "Партизанск", "Рязань",
    "Смоленск", "Томск", "Уссурийск", "Феодосия",
    "Хабаровск", "Цимлянск", "Чита", "Шатура",
    "Щелково", "Элиста", "Южно-Сахалинск", "Якутск",
  }

var Goods = map[int]TradingGood{
    // Id    Tier  Name       PriceMin  PriceMax  Unit        Volume  Weight  Resources  Consumables
    1: {1, 1, "Зерно", 2, 10, "мешок", 0.036, 0.050, nil, nil},
    2: {2, 1, "Дерево", 5, 20, "кубометр", 1.0, 0.640, nil, nil},
    3: {3, 1, "Камень", 4, 18, "кубометр", 1.0, 1.7, nil, nil},
    4: {4, 1, "Руда", 9, 30, "тонна", 0.5, 1.0, nil, nil},
  }

func RndRange(min, max int) int {
  return rand.Intn(max-min+1) + min
}

func Rnd(max int) int {
  return RndRange(1, max)
}

func Rnd100() int {
  return RndRange(1, 100)
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

