package main

import (
  "fmt"
/*  "log"
  "os"
  "time"*/

  "github.com/petuhoviliya/caravan/internal/common"
)


type Soldier struct {
	Grade int
	HP int
	Armor int
	Attack int
}


type Party struct {
	Soldiers map[int]Soldier
}

func (p *Party) NewParty(Grade, Count int){
	p.Soldiers = make(map[int]Soldier)
	for i := 1; i <= Count; i++ {
		p.Soldiers[i] = Soldier{
			Grade: 1,
			HP: common.RndRange(50,60),
			Armor: common.RndRange(5,10),
			Attack: common.RndRange(20,30),
		}
	}
}


func init() {

}

func main() {
	var (
		One Party
		Two Party
	)

	One.NewParty(1,10)
	Two.NewParty(1,10)

	DoBattle(&One, &Two)

	fmt.Printf("One: %+v\n", One)
	fmt.Printf("Two: %+v\n", Two)
}

func DoBattle(Attacker *Party, Defender *Party) {
	var (
		newAttacker Party
		newDefender Party	
	)
	for k, soldier := range Attacker.Soldiers {
		fmt.Printf("%d: %+v\n",k, soldier)
		
		
	}
	Attacker = &newAttacker
	Defender = &newDefender
}

