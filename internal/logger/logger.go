package logger

import (
	"TeleBotNotifications/internal/config"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

const OFF = 0
const ERROR_ONLY = 1
const FULL = 2

var General *log.Logger
var Error *log.Logger

var logFile *os.File

func init() {
	General = log.New(os.Stdout, "\x1b[37m", log.Ldate|log.Ltime)
	Error = log.New(os.Stderr, "\x1b[31mError:\t", log.Ldate|log.Ltime|log.Llongfile)
}

func Setup(conf *config.LoggerConfig, telegramBot io.Writer) error {
	var generalWrites []io.Writer
	var errorWriters []io.Writer

	if conf.StdLogLevel >= ERROR_ONLY {
		errorWriters = append(errorWriters, os.Stderr)
	}
	if conf.StdLogLevel == FULL {
		generalWrites = append(generalWrites, os.Stdout)
	}

	if conf.FileLogLevel >= ERROR_ONLY {
		err := createLogFile(conf.Path)
		if err != nil {
			return err
		}
		errorWriters = append(errorWriters, logFile)
	}
	if conf.FileLogLevel == FULL {
		generalWrites = append(generalWrites, logFile)
	}

	if conf.TelegramLogLevel >= ERROR_ONLY {
		errorWriters = append(errorWriters, telegramBot)
	}
	if conf.TelegramLogLevel == FULL {
		generalWrites = append(generalWrites, telegramBot)
	}

	generalWriter := io.MultiWriter(generalWrites...)
	errorWriter := io.MultiWriter(errorWriters...)

	General = log.New(generalWriter, "\x1b[37m", log.Ldate|log.Ltime)
	Error = log.New(errorWriter, "\x1b[31mError:\t", log.Ldate|log.Ltime|log.Llongfile)

	return nil
}

func createLogFile(folderPath string) (err error) {
	_, err = os.Stat(folderPath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(folderPath, 0766)
		if err != nil {
			return fmt.Errorf("error creating a folder: %s", err.Error())
		}
		err = os.Chmod(folderPath, 0766)
		if err != nil {
			return fmt.Errorf("error giving folder permissions: %s", err.Error())
		}
	} else if err != nil {
		return fmt.Errorf("error accessing a folder: %s", err.Error())
	}

	currentTime := time.Now()
	logFileName := fmt.Sprintf("%s/%s.log", folderPath, currentTime.Format("2006-01-02_15-04-05"))

	logFile, err = os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("error creating and opening file: %s", err.Error())
	}
	return nil
}
