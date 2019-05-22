package main

import (
	"github.com/wcharczuk/go-chart"
	"go.uber.org/zap"

	"bufio"
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
	re              = regexp.MustCompile(`\s*(\d*)\s*`)
	logger, _       = zap.NewDevelopment()
	maxVolume       = 250
	volumeThreshold = 3.
)

var (
	mux       sync.Mutex
	avgVolume = 0.
	peaks     = []Peak{}
	minutes   = []Minute{}
	message   = "We are good!"
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
	http.HandleFunc("/chart.png", drawChart)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("./html/index.html"))
		mux.Lock()
		defer mux.Unlock()
		data := TemplateData{
			message,
			maxVolume,
			avgVolume,
			volumeThreshold,
			peaks,
			minutes,
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, data)
	})
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./html/css/"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./html/js/"))))
	logger.Fatal("fail to serve", zap.Error(http.ListenAndServe("localhost:8090", nil)))
}

func drawChart(res http.ResponseWriter, req *http.Request) {
	/*
	   This is an example of using the `TimeSeries` to automatically coerce time.Time values into a continuous xrange.
	   Note: chart.TimeSeries implements `ValueFormatterProvider` and as a result gives the XAxis the appropriate formatter to use for the ticks.
	*/
	graph := chart.Chart{
		XAxis: chart.XAxis{
			Style: chart.StyleShow(),
		},
		Series: []chart.Series{
			chart.TimeSeries{
				XValues: []time.Time{
					time.Now().AddDate(0, 0, -10),
					time.Now().AddDate(0, 0, -9),
					time.Now().AddDate(0, 0, -8),
					time.Now().AddDate(0, 0, -7),
					time.Now().AddDate(0, 0, -6),
					time.Now().AddDate(0, 0, -5),
					time.Now().AddDate(0, 0, -4),
					time.Now().AddDate(0, 0, -3),
					time.Now().AddDate(0, 0, -2),
					time.Now().AddDate(0, 0, -1),
					time.Now(),
				},
				YValues: []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0, 11.0},
			},
		},
	}

	res.Header().Set("Content-Type", "image/png")
	graph.Render(chart.PNG, res)
}

func collectPeaks(dataCh chan data) {
	for {
		d := <-dataCh
		if float64(d.volume) > volumeThreshold*d.avg || d.volume > maxVolume {
			p := newPeak(d.time, d.volume, Black, d.avg, Black)
			logger.Info("peak collected", zap.Time("time", d.time), zap.Int("vol", d.volume), zap.Float64("avg", d.avg))
			mux.Lock()
			peaks = addPeak(peaks, p)
			message = updateMessage()
			avgVolume = d.avg
			mux.Unlock()
		}
	}
}

const (
	no   = "we are good"
	one  = "we can do better"
	more = "time to move elsewhere"
)

func updateMessage() string {
	m := no
	fiveAgo := time.Now().Add(-5 * time.Minute).Format("15:04:05")
	for _, p := range peaks {
		if p.Time > fiveAgo {
			switch m {
			case no:
				m = one
			case one:
				m = more
			default:
				return more
			}
		}
	}
	return m
}

func collectMinutes(dataCh chan data) {
	currentMinute := -1
	mSum := 0.
	mCount := 1.

	for {
		d := <-dataCh

		m := d.time.Minute()
		if currentMinute == -1 {
			currentMinute = m
		}
		if m != currentMinute {
			nm := newMinute(currentMinute, mSum/mCount, Black)

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
