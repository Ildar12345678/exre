package log

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type Logger struct {
	infoLogger *log.Logger
	errLogger  *log.Logger
	infoFile   *os.File
	errFile    *os.File
}

func NewLog(logDir, logFile string) (*Logger, error) {
	appLogger := new(Logger)
	var infoFile = os.Stdout
	var errFile = os.Stdout
	var err error
	if logFile != "stdout" {
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			if err = os.Mkdir(logDir, 0775); err != nil {
				return nil, fmt.Errorf("error while creating log dir: %s", err.Error())
			}
		}
		
		infoFile, err = os.OpenFile(filepath.Join(logDir, logFile+".info"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("error while creating log file: %s", err.Error())
		}
		errFile, err = os.OpenFile(filepath.Join(logDir, logFile+".err"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("error while creating log file: %s", err.Error())
		}
	}
	
	infoLogger := log.New(infoFile, "INFO ", log.LstdFlags)
	appLogger.infoLogger = infoLogger
	errLogger := log.New(errFile, "ERROR ", log.LstdFlags)
	appLogger.errLogger = errLogger
	
	return appLogger, nil
}

// Printf sends msg to logger with INFO prefix
func (l *Logger) Printf(msg string, v ...any) {
	l.infoLogger.Printf(msg, v...)
}

// Errorf sends msg to logger with ERROR prefix
func (l *Logger) Errorf(msg string, v ...any) {
	l.errLogger.Printf(msg, v...)
}

func (l *Logger) Close() {
	l.infoFile.Close()
	l.errFile.Close()
}
