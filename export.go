package main

import (
	"go.uber.org/zap"

	"bufio"
	"html/template"
	"io"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	re        = regexp.MustCompile(`\s*(\d*)\s*`)
	logger, _ = zap.NewDevelopment()

	maxAverageVolume = 200
	volumeThreshold  = 2.0
)

type data struct {
}

func main() {
	go readVolumes()
	serveVolumes()
}

func serveVolumes() {
	tmpl := template.Must(template.ParseFiles("./html/index.html"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := data{}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, data)
	})
	logger.Fatal("fail to serve", zap.Error(http.ListenAndServe("localhost:8090", nil)))
}

func readVolumes() {
	logger.Info("read volumes")
	reader := bufio.NewReader(os.Stdin)

	var max, sum, count int
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
