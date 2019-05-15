package main

import (
	"bufio"
	"go.uber.org/zap"
	"io"
	"os"
)

func main() {
	logger, _ := zap.NewDevelopment()
	logger.Info("start reading")

	reader := bufio.NewReader(os.Stdin)
	for {
		input, err := reader.ReadString('\r')
		if err != nil {
			if err == io.EOF {
				break
			}
			logger.Fatal("stop", zap.Error(err))
		}

		logger.Info("read", zap.String("input", input))
	}

	logger.Info("finish reading")
}
