package main

import (
	"bufio"
	"go.uber.org/zap"
	"html/template"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	re               = regexp.MustCompile(`\s*(\d*)\s*`)
	logger, _        = zap.NewDevelopment()
	maxAverageVolume = 200.0
	volumeThreshold  = 2.1
)

var (
	mux     sync.Mutex
	peaks   = []Peak{}
	minutes = []Minute{}
)

// data is passed by the reader to the collectors (peak and minute) for a specific time
type data struct {
	time   time.Time
	volume int     // current volume at 'time'
	avg    float64 // global average at 'time'
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
		mux.Lock()
		defer mux.Unlock()
		data := TemplateData{
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
		if float64(d.volume) > volumeThreshold*d.avg || d.avg > maxAverageVolume {
			p := newPeak(d.time, d.volume, Black, d.avg, Black)
			logger.Info("peak collected", zap.Time("time", d.time), zap.Int("vol", d.volume), zap.Float64("avg", d.avg))
			mux.Lock()
			peaks = addPeak(peaks, p)
			mux.Unlock()
		}
	}
}

var (
	currentMinute = -1
	mSum          = 0.
	mCount        = 1.
)

func collectMinutes(dataCh chan data) {
	for {
		d := <-dataCh

		m := d.time.Minute()
		if m != currentMinute {
			nm := newMinute(d.time, mSum/mCount, Black)

			// store info
			mux.Lock()
			minutes = addMinute(minutes, nm)
			mux.Unlock()
			logger.Info("minute collected", zap.String("time", nm.Time), zap.Float64("avg", nm.Average.Value))

			// next minute: reset counters
			currentMinute = m
			mSum = float64(d.volume)
			mCount = 1
		} else {
			mSum += float64(d.volume)
			mCount++
		}
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
