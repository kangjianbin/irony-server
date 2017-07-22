package main

import (
	"log"
	"os"
)

type fileLogger struct {
	logger *log.Logger
	file   *os.File
	debug  bool
}

var mlogger fileLogger

func initLogger() {
	mlogger.file = os.Stderr
	mlogger.logger = log.New(mlogger.file, "", log.Ltime|log.Lmicroseconds)
}

func setupLogger(fileName string) {
	releaseLogger()
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return
	}
	logger := log.New(f, "", log.LstdFlags)
	mlogger.logger = logger
	mlogger.file = f
}

func releaseLogger() {
	if mlogger.file == os.Stderr || mlogger.file == nil {
		return
	}
	mlogger.file.Close()
}

func setDebug(isOn bool) {
	mlogger.debug = isOn
}

func logDebug(format string, a ...interface{}) {
	if !mlogger.debug {
		return
	}
	mlogger.logger.Printf(format, a...)
}

func logInfo(format string, a ...interface{}) {
	if mlogger.logger == nil {
		return
	}
	mlogger.logger.Printf(format, a...)
}
