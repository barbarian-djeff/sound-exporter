package main

import (
	"go.uber.org/zap"
	"time"

	"bufio"
	"html/template"
	"io"
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
	volumeThreshold  = 2.1

	peaks   = []peak{}
	minutes = []minute{}
)

type color string

const (
	blue  color = "color: #0000ff;"
	red   color = "color: #ff0000;"
	green color = "color: #00ff00;"
)

type peak struct {
	time    string
	current volume
	average volume
}

func newPeak(t time.Time, curVol int, curCol color, avgVol int, avgCol color) peak {
	return peak{
		t.Format("15:04:05"),
		volume{curVol, curCol},
		volume{avgVol, avgCol},
	}
}

type minute struct {
	time    string
	average volume
}

func newMinute(t time.Time, v int, c color) minute {
	return minute{
		t.Format("15:04"),
		volume{v, c},
	}
}

type volume struct {
	value int
	color color
}

type templateData struct {
	message          string
	maxAverageVolume int
	volumeThreshold  float64
	peaks            []peak
	minutes          []minute
}

// data is passed by the reader to the collectors (peak and minute)
type data struct {
	time   time.Time
	volume int
	avg    float64
}

func newData(v int, avg float64) data {
	return data{time.Now(), v, avg}
}

func main() {
	peakChannel := make(chan data)
	minuteChannel := make(chan data)
	go collectMinutes(minuteChannel)
	go collectPeaks(peakChannel)
	go readVolumes(peakChannel, minuteChannel)
	serveVolumes()
}

func serveVolumes() {
	tmpl := template.Must(template.ParseFiles("./html/index.html"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := templateData{
			"We are good!",
			maxAverageVolume,
			volumeThreshold,
			peaks,
			minutes,
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, data)
	})
	logger.Fatal("fail to serve", zap.Error(http.ListenAndServe("localhost:8090", nil)))
}

func collectPeaks(dataCh chan data) {
	for {
		d := <-dataCh
		logger.Info("collect peak", zap.Time("time", d.time), zap.Int("vol", d.volume), zap.Float64("avg", d.avg))
	}
}

func collectMinutes(dataCh chan data) {
	for {
		d := <-dataCh
		logger.Info("collect minutes", zap.Time("time", d.time), zap.Int("vol", d.volume), zap.Float64("avg", d.avg))
	}
}

func readVolumes(channels ...chan data) {
	logger.Info("read volumes")
	reader := bufio.NewReader(os.Stdin)

	var sum, count int
	for {
		input, err := reader.ReadString('\r')
		if err != nil {
			if err == io.EOF {
				break
			}
			logger.Fatal("stop", zap.Error(err))
		}
		if len(strings.TrimSpace(input)) > 0 {
			vol, ok := readVolume(input)
			if !ok {
				break
			}

			count++
			sum += vol
			avg := float64(sum) / float64(count)
			data := newData(vol, avg)
			for _, ch := range channels {
				ch <- data
			}
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
