package log

import (
	"log"
	"os"
	"fmt"
)

type Logger struct {
	InfoLogger *log.Logger
	ErrLogger  *log.Logger
	infoFile   *os.File
	errFile    *os.File
}

// var AppLogger *Logger

func NewLog(logDir, logFile string) (*Logger, error) {
	appLogger := new(Logger)
	var infoFile = os.Stdout
	var errFile = os.Stdout
	var err error
	if logFile != "stdout" {
		if err = os.Mkdir(logDir, 0775); err != nil {
			return nil, fmt.Errorf("error while creating log dir: %s", err.Error())
		}
		infoFile, err = os.OpenFile(fmt.Sprintf("%s/%s.info", logDir, logFile), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("error while creating log file: %s", err.Error())
		}
		errFile, err = os.OpenFile(fmt.Sprintf("%s/%s.err", logDir, logFile), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("error while creating log file: %s", err.Error())
		}
	}
	
	infoLogger := log.New(infoFile, "INFO\t", log.LstdFlags)
	appLogger.InfoLogger = infoLogger
	errLogger := log.New(errFile, "ERROR\t", log.LstdFlags)
	appLogger.ErrLogger = errLogger
	
	return appLogger, nil
}

// Printf sends msg to logger with INFO prefix
func (l *Logger) Printf(msg string, v ...any) {
	l.InfoLogger.Printf(msg, v...)
}

// Errorf sends msg to logger with ERROR prefix
func (l *Logger) Errorf(msg string, v ...any) {
	l.ErrLogger.Printf(msg, v...)
}

// todo make proper close for concurrent usage
func (l *Logger) Close() {
	l.infoFile.Close()
	l.errFile.Close()
}
