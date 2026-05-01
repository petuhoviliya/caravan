package main

import (
  "errors"
  "fmt"
  "hash/fnv"
  "math/rand"
)

// validateSeed проверяет соответствие строки требованиям:
// длина от 1 до 10, только латинские буквы (A-Z, a-z) и цифры (0-9)
func validateSeed(s string) error {
  if len(s) < 1 || len(s) > 10 {
    return errors.New("длина строки должна быть от 1 до 10 символов")
  }
  for _, r := range s {
    if !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')) {
      return errors.New("строка должна содержать только латинские буквы (A-Z, a-z) и цифры (0-9)")
    }
  }
  return nil
}

// stringToSeed детерминированно преобразует строку в int64
func stringToSeed(s string) int64 {
  // FNV-1a 64-bit: быстрый, детерминированный, входит в stdlib
  h := fnv.New64a()
  h.Write([]byte(s))
  // uint64 -> int64. math/rand корректно обрабатывает любые биты зерна.
  return int64(h.Sum64())
}

func main() {
  // Пример валидных и невалидных строк
  testStrings := []string{"Ab3", "GoLang2026", "short", "with spaces!", ""}

  for _, s := range testStrings {
    fmt.Printf("🔍 Проверка %q: ", s)
    if err := validateSeed(s); err != nil {
      fmt.Printf("❌ %v\n", err)
      continue
    }

    // 1. Преобразуем строку в числовое зерно
    seed := stringToSeed(s)

    // 2. Создаём изолированный генератор с этим зерном
    rng := rand.New(rand.NewSource(seed))

    // 3. Генерация случайных данных
    fmt.Printf("✅ seed=%d | Int(0-99)=%d | Float32=%.3f\n",
      seed,
      rng.Intn(100),
      rng.Float32(),
    )
  }
}
