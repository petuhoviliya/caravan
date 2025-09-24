package main

import (
	"fmt"
	"math/rand"
	"time"
)

// Soldier представляет солдата
type Soldier struct {
	Health int
	Damage int
	Armor  int
}

// Army представляет армию
type Army struct {
	Name     string
	Soldiers []Soldier
}

// NewSoldier создает нового солдата с случайными параметрами
func NewSoldier(grade int) Soldier {
	return Soldier{
		Health: 50 + rand.Intn(grade), // от 50 до 100
		Damage: 5  + rand.Intn(grade),
		Armor:  1  + rand.Intn(grade),  // от 5 до 15
	}
}

// NewArmy создает новую армию с заданным количеством солдат
func NewArmy(name string, grade, count int) Army {
	soldiers := make([]Soldier, count)
	for i := 0; i < count; i++ {
		soldiers[i] = NewSoldier(grade)
	}
	return Army{
		Name:     name,
		Soldiers: soldiers,
	}
}

// Attack атакует другого солдата
// Возвращает нанесенный урон (атака минус броня, но не менее 1)
func (s *Soldier) Attack(target *Soldier) {
	damage := s.Damage + rand.Intn(5) - target.Armor
	if damage < 1 {
		damage = 1 // минимальный урон
	}
	target.Health -= damage
	if target.Health < 0 {
		target.Health = 0
	}
}

// IsAlive проверяет, жив ли солдат
func (s *Soldier) IsAlive() bool {
	return s.Health > 0
}

// Battle проводит битву между двумя армиями
func Battle(army1, army2 *Army) int{

	round := 1

	for len(army1.AliveSoldiers()) > 0 && len(army2.AliveSoldiers()) > 0 {
		//fmt.Printf("--- Раунд %d ---\n", round)
		
		// Получаем живых солдат
		alive1 := army1.AliveSoldiers()
		alive2 := army2.AliveSoldiers()
		
		// Случайный солдат из первой армии атакует случайного из второй
		if len(alive1) > 0 && len(alive2) > 0 {
			soldier1 := &army1.Soldiers[alive1[rand.Intn(len(alive1))]]
			soldier2 := &army2.Soldiers[alive2[rand.Intn(len(alive2))]]
			
			if soldier1.IsAlive() && soldier2.IsAlive() {
				soldier1.Health = soldier1.Health
				soldier1.Damage = soldier1.Damage
				soldier1.Armor = soldier1.Armor
				
				soldier2.Health = soldier2.Health
				soldier2.Damage = soldier2.Damage
				soldier2.Armor = soldier2.Armor

				soldier2.Attack(soldier1)
				soldier1.Attack(soldier2)
			}
		}
		
		round++
	}
	// Определяем победителя
	if len(army1.AliveSoldiers()) > 0 {
		return 1
	} else if len(army2.AliveSoldiers()) > 0 {
		return 2
	} else {
		return 0
	}
}

// AliveSoldiers возвращает индексы живых солдат
func (a *Army) AliveSoldiers() []int {
	var alive []int
	for i, soldier := range a.Soldiers {
		if soldier.IsAlive() {
			alive = append(alive, i)
		}
	}
	return alive
}

func main() {
	var (
		winFirst int
		winSecond int
		Draw int
		totalBattles int = 1000
		army1 Army
		army2 Army
	)

	rand.Seed(time.Now().UnixNano())
	
	// Проводим битву
	for i := 1; i <= totalBattles; i++ {
	
		army1 = NewArmy("Красные", 2, 84)
		army2 = NewArmy("Синие", 1, 100)

		result := Battle(&army1, &army2)

		if result == 1 {
			winFirst++
		}else if result == 2 {
			winSecond++
		}else{
			Draw++
		}
	}
	
	fmt.Printf("Итог:\n%s: %d\n%s: %d\nНичьи: %d\n", 
		army1.Name, winFirst,
		army2.Name, winSecond,
		Draw,
		)
}
