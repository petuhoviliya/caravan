package logger

import (
  "fmt"
  "log"
  "os"
  "sync"
  //"time"
)

var (
  loggerInstance *log.Logger
  once           sync.Once
  logFile        *os.File
)

// Init инициализирует логгер с указанным файлом
func Init(filename string) error {
  var err error

  once.Do(func() {
    // Открываем файл для логирования (добавляем к существующему содержимому)
    logFile, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
      fmt.Printf("Ошибка открытия файла лога: %v\n", err)
      return
    }

    // Создаем логгер
    loggerInstance = log.New(logFile, "", log.Ldate|log.Ltime|log.Lshortfile)
  })

  return err
}

// Close закрывает файл лога
func Close() {
  if logFile != nil {
    logFile.Close()
  }
}

// Info записывает информационное сообщение
func Info(v ...interface{}) {
  if loggerInstance != nil {
    loggerInstance.SetPrefix("INFO: ")
    loggerInstance.Output(2, fmt.Sprint(v...))
  }
}

// Infof записывает форматированное информационное сообщение
func Infof(format string, v ...interface{}) {
  if loggerInstance != nil {
    loggerInstance.SetPrefix("INFO: ")
    loggerInstance.Output(2, fmt.Sprintf(format, v...))
  }
}

// Error записывает сообщение об ошибке
func Error(v ...interface{}) {
  if loggerInstance != nil {
    loggerInstance.SetPrefix("ERROR: ")
    loggerInstance.Output(2, fmt.Sprint(v...))
  }
}

// Errorf записывает форматированное сообщение об ошибке
func Errorf(format string, v ...interface{}) {
  if loggerInstance != nil {
    loggerInstance.SetPrefix("ERROR: ")
    loggerInstance.Output(2, fmt.Sprintf(format, v...))
  }
}

// Warning записывает предупреждение
func Warning(v ...interface{}) {
  if loggerInstance != nil {
    loggerInstance.SetPrefix("WARNING: ")
    loggerInstance.Output(2, fmt.Sprint(v...))
  }
}

// Warningf записывает форматированное предупреждение
func Warningf(format string, v ...interface{}) {
  if loggerInstance != nil {
    loggerInstance.SetPrefix("WARNING: ")
    loggerInstance.Output(2, fmt.Sprintf(format, v...))
  }
}

// Debug записывает отладочное сообщение
func Debug(v ...interface{}) {
  if loggerInstance != nil {
    loggerInstance.SetPrefix("DEBUG: ")
    loggerInstance.Output(2, fmt.Sprint(v...))
  }
}

// Debugf записывает форматированное отладочное сообщение
func Debugf(format string, v ...interface{}) {
  if loggerInstance != nil {
    loggerInstance.SetPrefix("DEBUG: ")
    loggerInstance.Output(2, fmt.Sprintf(format, v...))
  }
}
