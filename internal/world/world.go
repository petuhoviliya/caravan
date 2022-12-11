package world

import (
  "math"

  "github.com/petuhoviliya/caravan/internal/common"
)

type Template struct {
  Width  int
  Height int
  BitMap []byte
}

func (t *Template) Size() int {
  return t.Width * t.Height
}

func (t *Template) Position(Index int) (int, int) {
  X := Index / t.Height
  Y := Index % t.Height
  return X, Y
}

func (t *Template) Index(X, Y int) int {
  return X*t.Height + Y
}

func (t *Template) MakeBitmap() {
  t.BitMap = make([]byte, t.Size())
}

func (t *Template) FreeCell() (int, int) {
  free := t.GetFreeCells()
  index := common.RndRange(0, len(free)-1)
  return t.Position(free[index])

}

func (t *Template) GetFreeCells() []int {

  var free []int

  for index, value := range t.BitMap {
    if value == 0 {
      free = append(free, index)
    }
  }

  return free
}

func (t *Template) PlaceTown(X, Y, Radius int) bool {

  for i := -Radius; i <= Radius; i++ {
    for j := -Radius; j <= Radius; j++ {

      C := int(math.Hypot(float64(i), float64(j)))

      tX := X + i
      tY := Y + j

      if C <= Radius {

        if tX > (t.Width - 1) {
          tX = (t.Width - 1)
        }
        if tY > (t.Height - 1) {
          tY = (t.Height - 1)
        }
        if tX < 0 {
          tX = 0
        }
        if tY < 0 {
          tY = 0
        }

        t.BitMap[t.Index(tX, tY)] = 1
      }
    } // for j
  } // for i

  return true
}

