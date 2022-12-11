package main

import (
	"fmt"
)

func main() {
	const rows, cols = 60, 20

	// Пример: найти координаты для числа 500
	numberToFind := 91
	row, col := findCoordinates(rows, cols, numberToFind)

	if row != -1 && col != -1 {
		fmt.Printf("Число %d находится в строке %d и столбце %d\n", numberToFind, row, col)
	} else {
		fmt.Printf("Число %d не найдено в массиве\n", numberToFind)
	}
}

// Функция для нахождения координат числа в двумерном массиве без использования циклов
func findCoordinates(rows, cols, number int) (int, int) {
	if number < 0 || number >= rows*cols {
		return -1, -1 // Число вне диапазона
	}

	row := number / cols
	col := number % cols

	return row, col
}
