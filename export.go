package main

import (
	"bufio"
	"go.uber.org/zap"
	"io"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	re        = regexp.MustCompile(`\s*(\d*)\s*`)
	logger, _ = zap.NewDevelopment()
)

var (
	max, sum, count int
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

			count++
			sum += v
			avg := float64(sum) / float64(count)
			dev := math.Abs(float64(v) - avg)
			if v > max {
				max = v
			}
			logger.Info("record", zap.Int("volume", v), zap.Int("max", max), zap.Float64("average", avg), zap.Float64("dev", dev))
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
