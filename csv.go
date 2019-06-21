package main

import (
	"go.uber.org/zap"
	"os"
)

func openFiles(fileName string) []*os.File {
	logger.Info("write into file", zap.String("name", fileName))
	if fileName == "" {
		return []*os.File{os.Stdout}
	}

	os.Remove(fileName)
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0655)
	if err != nil {
		logger.Fatal("fail to open file", zap.Error(err))
	}
	return []*os.File{os.Stdout, f}
}

func closeFiles(files []*os.File) {
	for _, f := range files {
		if err := f.Close(); err != nil {
			logger.Fatal("fail to close file", zap.Error(err))
		}
	}
}
