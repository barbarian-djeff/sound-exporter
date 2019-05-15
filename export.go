package main

import (
	"bufio"
	"go.uber.org/zap"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	re        = regexp.MustCompile(`\s*(\d*)\s*`)
	logger, _ = zap.NewDevelopment()
)

func main() {

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

		if len(strings.TrimSpace(input)) > 0 {
			v, ok := readVolume(input)
			if !ok {
				break
			}
			logger.Info("record", zap.Int("volume", v))
		}
	}

	logger.Info("finish reading")
}

func readVolume(input string) (int, bool) {
	matches := re.FindStringSubmatch(input)
	if len(matches) < 2 {
		logger.Error("fail to find number", zap.String("input", input))
		return 0, false
	}
	number := matches[1]
	volume, err := strconv.Atoi(number)
	if err != nil {
		logger.Error("fail to parse number", zap.Error(err), zap.String("number", number))
		return 0, false
	}

	return volume, true
}
