package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

const OFF = 0
const ERROR_ONLY = 1
const FULL = 2

var generalLogger *log.Logger
var generalFileLogger *log.Logger
var errorLogger *log.Logger
var errorFileLogger *log.Logger
var telegramLogger func(string) error

var TelegramLogLevel uint = 0
var FileLogLevel uint = 1
var StdLogLevel uint = 1

// Заменить много логгров на много *os.File, чтобы логгер правильно показывал файл, в котором возникла ошибка

func init() {
	generalLogger = log.New(os.Stdout, "General:\t", log.Ldate|log.Ltime)
	errorLogger = log.New(os.Stderr, "Error:\t", log.Ldate|log.Ltime|log.Llongfile)

	absPath, err := filepath.Abs("./logs")
	if err != nil {
		errorLogger.Printf("Error reading given path: %s\n", err.Error())
	}

	currentTime := time.Now()
	logFileName := fmt.Sprintf("%s/%s.log", absPath, currentTime.Format("2006-01-02_15-04-05"))
	generalLog, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		errorLogger.Printf("Error opening file: %s\n", err.Error())
	}

	generalFileLogger = log.New(generalLog, "General:\t", log.Ldate|log.Ltime)
	errorFileLogger = log.New(generalLog, "Error:\t", log.Ldate|log.Ltime|log.Llongfile)
}

func SetupTelegramLogger(output func(string) error) {
	telegramLogger = output
}


func Println(line string) {
	if StdLogLevel == FULL {
		generalLogger.Println(line)
	}
	if FileLogLevel == FULL {
		generalFileLogger.Println(line)
	}
	if TelegramLogLevel == FULL {
		telegramLogger(line)
	}
}

func Error(line string, err error) {
	line = line + err.Error()
	if StdLogLevel >= ERROR_ONLY {
		errorLogger.Println(line)
	}
	if FileLogLevel >= ERROR_ONLY {
		errorFileLogger.Println(line)
	}
	if TelegramLogLevel >= ERROR_ONLY {
		telegramLogger(line)
	}
}
