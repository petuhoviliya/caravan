package main

import (
  "fmt"
/*  "log"
  "os"
  "time"*/

  "github.com/petuhoviliya/caravan/internal/common"
)

const (
  MaxHitChance int = 95
  MinDamage    int = 1

  BaseHealth      int = 50
  BaseAttack      int = 20
  BaseArmor       int = 10
  BaseHitChance   int = 60
  BaseEvadeChance int = 5

  VarHealth int = 10
  VarAttack int = 10
  VarArmor  int = 5
)


type Soldier struct {
  Grade    int
  Health   int
  Armor    int
  Attack   int
  Accuracy int
  Evade    int
}

type Party struct {
  Soldiers map[int]Soldier
}


func (s *Soldier) IsAlive() bool {
  return s.Health > 0
}

func (s *Soldier) IsHit() bool {
  if s.Accuracy > MaxHitChance {
    s.Accuracy = MaxHitChance
  }
  return s.Accuracy >= common.Rnd100()
}

func (s *Soldier) IsEvade() bool {
  return s.Evade >= common.Rnd100()
}

func (s *Soldier) DoAttack(target *Soldier) int {

  if !s.IsHit() { return -1 }       // промахнулся
  if target.IsEvade() { return -2 } // цель увернулась

  damage := s.Attack - target.Armor
  if damage < MinDamage {
    damage = MinDamage // минимальный урон
  }
  target.Health -= damage
  if target.Health < 0 {
    target.Health = 0
  }

  return damage
}

func (p *Party) NewParty(Grade, Count int){
  p.Soldiers = make(map[int]Soldier)
  for i := 1; i <= Count; i++ {
    p.Soldiers[i] = Soldier{
      Grade:    Grade,
      Health:   common.RndRange(BaseHealth, BaseHealth + VarHealth) + Grade*10,
      Armor:    common.RndRange(BaseArmor,  BaseArmor  + VarArmor)  + Grade*2,
      Attack:   common.RndRange(BaseAttack, BaseAttack + VarAttack) + Grade,
      Accuracy: BaseHitChance   + Grade*5,
      Evade:    BaseEvadeChance + Grade,
    }
  }
}

func (p *Party) GetAlive() []int {
  var alive []int
  for i, soldier := range p.Soldiers {
    if soldier.IsAlive() {
      alive = append(alive, i)
    }
  }
  return alive
}

func init() {

}

func main() {
  var (
    One Party
    Two Party
  )

  One.NewParty(7,5)
  Two.NewParty(5,10)

  DoBattle(&One, &Two)

  //fmt.Printf("One: %+v\n", One)
  //fmt.Printf("Two: %+v\n", Two)
}

func DoBattle(A *Party, B *Party) {
  var (
    SoldierA, SoldierB Soldier
    ResultA, ResultB, round int

//    newAttacker Party
//    newDefender Party
  )

  fmt.Printf("Битва начинается\n")

  for len(A.GetAlive()) > 0 && len(B.GetAlive()) > 0 {
    round++
    fmt.Printf("- Раунд %d -\n", round)

    // Получаем индексы живых солдат
    Attackers := A.GetAlive()
    Defenders := B.GetAlive()

    // Случайный солдат из первой армии атакует случайного из второй
    if len(Attackers) > 0 && len(Defenders) > 0 {

      RndA := Attackers[common.Rnd(len(Attackers))-1]
      RndB := Defenders[common.Rnd(len(Defenders))-1]

      SoldierA = A.Soldiers[RndA]
      SoldierB = B.Soldiers[RndB]

      if SoldierA.IsAlive() && SoldierB.IsAlive() {
        ResultA = SoldierA.DoAttack(&SoldierB)
        ResultB = SoldierB.DoAttack(&SoldierA)
      }
      if ResultA == -1 {
        fmt.Printf("  >> Солдат А промахивается по Солдату Б\n")
      }else if ResultA == -2 {
        fmt.Printf("  >> Солдат Б увернулся от атаки Солдата А\n")
      }else {
        fmt.Printf("  >> Солдат А попадает по Солдату Б с уроном: %d\n", ResultA)
      }

      if ResultB == -1 {
        fmt.Printf("  << Солдат Б промахивается по Солдату А\n")
      }else if ResultB == -2 {
        fmt.Printf("  << Солдат А увернулся от атаки Солдата Б\n")
      }else {
        fmt.Printf("  << Солдат Б попадает по Солдату А с уроном: %d\n", ResultB)
      }

      //fmt.Printf("%#v\n", SoldierA)
      //fmt.Printf("%#v\n", SoldierB)

      A.Soldiers[RndA] = SoldierA
      B.Soldiers[RndB] = SoldierB
    }
  }
  fmt.Println("Битва окончена")
  if len(A.GetAlive()) > len(B.GetAlive()) {
    fmt.Printf("Победила Армия А, в живых осталось: %d\n\n", len(A.GetAlive()))
  }else if len(A.GetAlive()) < len(B.GetAlive()) {
    fmt.Printf("Победила Армия Б, в живых осталось: %d\n\n", len(B.GetAlive()))
  }else{
    fmt.Printf("Все умерли\n\n")
  }

  fmt.Printf("%+v\n", A)
  fmt.Printf("%+v\n", B)

/*  for k, soldier := range A.Soldiers {
    fmt.Printf("%d: %+v\n",k, soldier)
  }*/

  //A = &newAttacker
  //B = &newDefender
}

