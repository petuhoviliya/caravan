package main

import (
	"fmt"
	//"log"
	//"math"
	"math/rand"
	"os"
	//"os/signal"
	//"syscall"
	//"time"
	//"github.com/gdamore/tcell/v2"
	//"github.com/rivo/tview"
)

/*
\033[0m - reset
FG  BG
30   40   Black
31   41   Red
32   42   Green
33   43   Yellow
34   44   Blue
35   45   Magenta
36   46   Cyan
37   47   White
90   100   Bright Black (Gray)
91   101   Bright Red
92   102   Bright Green
93   103   Bright Yellow
94   104   Bright Blue
95   105   Bright Magenta
96   106   Bright Cyan
97   107   Bright White
*/

/*
"id": "desert",  \033
"id": "steppe",
"id": "plain",
"id": "forest",
"id": "mountain",
*/

type Relief struct {
	id    int8
	name  string
	color string
}

type MapTemplate struct {
	Rows int
	Cols int
	//  BitMap []byte
	ReliefMap []int8
}

type Square struct {
	X1 int
	Y1 int
	X2 int
	Y2 int
}

var (
	ReliefTypes []Relief
	BigSquare   []Square
	SmallSquare []Square
)

func init() {
	ReliefTypes = []Relief{
		Relief{
			id:    1,
			name:  "plain",
			color: "\033[48:5:40m\033[38:5:46m.",
			//color: "🌱",
		},
		Relief{
			id:    2,
			name:  "forest",
			color: "\033[48:5:34m\033[38:5:232m*",
			//color: "🌳",
			//color: "🌲",
		},
		Relief{
			id:    3,
			name:  "steppe",
			color: "\033[48:5:178m\033[38:5:232m*",
			//color: "🌾",
		},
		Relief{
			id:    4,
			name:  "desert",
			color: "\033[48:5:220m\033[38:5:232m*",
			//color: "🌵",
		},
		Relief{
			id:    5,
			name:  "mountain",
			color: "\033[48:5:240m\033[38:5:232m^",
			//color: "⛰",
		},
	}

}

func main() {
	cols := 60
	rows := 16

	s := Square{
		X1: 0,
		Y1: 0,
		X2: rows,
		Y2: cols,
	}
	BigSquare = SquareMap(s)
	SmallSquare = SquareMap(BigSquare[0])

	fmt.Printf("%v\n%v\n", BigSquare, SmallSquare)

	for _, v := range ReliefTypes {
		fmt.Printf("%s ", v.name)
		for i := 0; i < 10; i++ {
			fmt.Printf("%s ", v.color)
		}
		fmt.Printf("\033[0m\n")
	}

	GMap := NewMap(rows, cols)
	GMap.GenerateRelief()

	//  fmt.Printf("%+v",GMap.ReliefMap)

	for k, v := range GMap.ReliefMap {
		if k%GMap.Cols == 0 {
			fmt.Printf("\n")
		}
		//fmt.Printf("%d ", k)
		fmt.Printf("%s\033[0m", ReliefTypes[v].color)
	}

	fmt.Printf("\n")

	os.Exit(0)
}

func NewMap(rows int, cols int) *MapTemplate {

	Map := MapTemplate{Rows: rows, Cols: cols}
	Map.MakeReliefMap()

	return &Map
}

func RndRange(Min int, Max int) int {
	return rand.Intn(Max-Min+1) + Min
}

func Rnd(Max int) int {
	return RndRange(1, Max)
}

func SquareMap(s Square) []Square {
	var square []Square

	square = append(square,
		Square{ // 0
			X1: s.X1,
			Y1: s.Y1,
			X2: (s.X2 - s.X1) / 2,
			Y2: (s.Y2 - s.Y1) / 2,
		},
		Square{ // 1
			X1: s.X1,
			Y1: (s.Y2 - s.Y1) / 2,
			X2: (s.X2 - s.X1) / 2,
			Y2: s.Y2,
		},
		Square{ // 2
			X1: (s.X2 - s.X1) / 2,
			Y1: s.Y1,
			X2: s.X2,
			Y2: (s.Y2 - s.Y1) / 2,
		},
		Square{ // 3
			X1: (s.X2 - s.X1) / 2,
			Y1: (s.Y2 - s.Y1) / 2,
			X2: s.X2,
			Y2: s.Y2,
		},
	)

	return square
}

func (m *MapTemplate) GenerateRelief() {

	for k, _ := range m.ReliefMap {
		//i = int8(RndRange(0, len(ReliefTypes)-1))

		x, y := m.GetPosition(k)

		i := 0
		j := 3

		if (x >= BigSquare[i].X1 && x < BigSquare[i].X2) && (y >= BigSquare[i].Y1 && y < BigSquare[i].Y2) {
			m.ReliefMap[k] = 1
		} else if (x >= BigSquare[j].X1 && x < BigSquare[j].X2) && (y >= BigSquare[j].Y1 && y < BigSquare[j].Y2) {
			m.ReliefMap[k] = 4
		} else {
			m.ReliefMap[k] = 0
			//m.ReliefMap[k] = 0
		}

		if (x >= SmallSquare[i].X1 && x < SmallSquare[i].X2) && (y >= SmallSquare[i].Y1 && y < SmallSquare[i].Y2) {
			m.ReliefMap[k] = 3
		} else if (x >= SmallSquare[j].X1 && x < SmallSquare[j].X2) && (y >= SmallSquare[j].Y1 && y < SmallSquare[j].Y2) {
			m.ReliefMap[k] = 2
		}

	}
}

func (m *MapTemplate) Size() int {
	return m.Rows * m.Cols
}

func (m *MapTemplate) GetPosition(Index int) (int, int) {
	X := Index / m.Cols
	Y := Index % m.Cols
	return X, Y
}

func (m *MapTemplate) GetIndex(X int, Y int) int {
	return X*m.Cols + Y
}

/*func (m *MapTemplate) MakeBitmap() {
  m.BitMap = make([]byte, m.Size())
}*/

func (m *MapTemplate) MakeReliefMap() {
	m.ReliefMap = make([]int8, m.Size())
}

/*func (m *MapTemplate) GetFreeCells() []int {

  var free []int

  for index, value := range m.BitMap {
    if value == 0 {
      free = append(free, index)
    }
  }

  return free
}*/

/*func (m *MapTemplate) PlaceTown(X int, Y int, Radius int) bool{

  freeCells := m.GetFreeCells()

  if len(freeCells) == 0 {
    return false
  }

  for i := -Radius; i <= Radius; i++ {
    for j := -Radius; j <= Radius; j++ {


      A := math.Abs(float64(0 - i))
      B := math.Abs(float64(0 - j))
      C := int(math.Sqrt(math.Pow(A, 2) + math.Pow(B, 2)))

      tX := X + i
      tY := Y + j
      //fmt.Printf("-- %v, \n", PointInsideRadius(i, j, Radius))

      if C <= Radius {

        if tX > m.Width {
          tX = m.Width
        }
        if tY > m.Height {
          tY = m.Height
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
}*/
