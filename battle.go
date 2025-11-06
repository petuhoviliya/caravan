package main

import (
  "fmt"
/*  "log"
  "os"
  "time"*/

  "github.com/petuhoviliya/caravan/internal/common"
)

const MaxHitChance int = 95
const MinDamage int = 1


type Soldier struct {
	Grade int
	Health int
	Armor int
	Attack int
	Accuracy int
	Evade int
}


type Party struct {
	Soldiers map[int]Soldier
}


func (s *Soldier) IsAlive() bool {
	return s.Health > 0
}

func (s *Soldier) IsHit() bool {
	return s.Accuracy >= common.Rnd(100)
}

func (s *Soldier) IsEvade() bool {
	return s.Evade >= common.Rnd(100)
}

func (s *Soldier) DoAttack(target *Soldier) int {

	if !s.IsHit() { return -1 } 			// промахнулся
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
			Grade: 1,
			Health: common.RndRange(50,60),
			Armor: common.RndRange(5,10),
			Attack: common.RndRange(20,30),
			Accuracy: 75,
			Evade: 5,
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

	One.NewParty(1,1)
	Two.NewParty(1,1)

	DoBattle(&One, &Two)

	//fmt.Printf("One: %+v\n", One)
	//fmt.Printf("Two: %+v\n", Two)
}

func DoBattle(A *Party, B *Party) {
	var (
		SoldierA Soldier
		SoldierB Soldier
		ResultA, ResultB int
//		newAttacker Party
//		newDefender Party	
	)

	for len(A.GetAlive()) > 0 && len(B.GetAlive()) > 0 {
		//fmt.Printf("--- Раунд %d ---\n", round)
		
		// Получаем индексы живых солдат
		Attackers := A.GetAlive()
		Defenders := B.GetAlive()
		
		// Случайный солдат из первой армии атакует случайного из второй
		if len(Attackers) > 0 && len(Defenders) > 0 {
			fmt.Printf("--- Новый бой\n")

			RndA := Attackers[common.Rnd(len(Attackers))-1]
			RndB := Attackers[common.Rnd(len(Defenders))-1]

			SoldierA = A.Soldiers[RndA]
			SoldierB = B.Soldiers[RndB]
			
			if SoldierA.IsAlive() && SoldierB.IsAlive() {
				ResultA = SoldierA.DoAttack(&SoldierB)
				ResultB = SoldierB.DoAttack(&SoldierA)
			}
			if ResultA == -1 {
				fmt.Printf(">> Солдат А промахивается по Солдату Б\n")
			}else if ResultA == -2 {
				fmt.Printf(">> Солдат Б увернулся от атаки Солдата А\n")
			}else {
				fmt.Printf(">> Солдат А попадает по Солдату Б с уроном: %d\n", ResultA)
			}

			if ResultB == -1 {
				fmt.Printf("<< Солдат Б промахивается по Солдату А\n")
			}else if ResultB == -2 {
				fmt.Printf("<< Солдат А увернулся от атаки Солдата Б\n")
			}else {
				fmt.Printf("<< Солдат Б попадает по Солдату А с уроном: %d\n", ResultB)
			}

			fmt.Printf("%#v\n", SoldierA)
			fmt.Printf("%#v\n", SoldierB)

			A.Soldiers[RndA] = SoldierA
			B.Soldiers[RndB] = SoldierB
		}
	}
/*	for k, soldier := range A.Soldiers {
		fmt.Printf("%d: %+v\n",k, soldier)
	}*/

	//A = &newAttacker
	//B = &newDefender
}

