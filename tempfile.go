package main

import (
	"io/ioutil"
	"os"
)

const (
	IronyTempPrefix = "irony-temp"
)

var tempFile *os.File

func getTempFilePath() string {
	if tempFile == nil {
		var err error
		tempFile, err = ioutil.TempFile("", IronyTempPrefix)
		if err != nil {
			exitError("Failed to create temp file")
		}
	}
	return tempFile.Name()
}

func closeTempFile() {
	if tempFile != nil {
		tempFile.Close()
		tempFile = nil
	}
}
