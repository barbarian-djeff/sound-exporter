package main

import (
	"fmt"
	"github.com/gordonklaus/portaudio"
	"github.com/wcharczuk/go-chart"
	"go.uber.org/zap"
	"regexp"
	"sync"
	"time"
)

var (
	re              = regexp.MustCompile(`\s*(\d*)\s*`)
	logger, _       = zap.NewDevelopment()
	maxVolume       = 100000.
	volumeThreshold = 3.
)

var (
	mux       sync.Mutex
	avgVolume = 0.
	peaks     = []Peak{}
	minutes   = []Minute{}
	serie     = chart.TimeSeries{
		XValues: []time.Time{},
		YValues: []float64{},
	}
	message = "We are good!"
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

func readFromPortAudio(stop chan int, stopped chan int, channels ...chan data) {
	logger.Info("read from portaudio")

	portaudio.Initialize()
	defer portaudio.Terminate()
	in := make([]int32, 64)
	stream, err := portaudio.OpenDefaultStream(1, 0, 44100, len(in), in)
	chk(err)
	defer stream.Close()

	chk(stream.Start())

	sa := NewSoundAggregator()
	count := 0
	var sum int
	for {
		err := stream.Read()
		if err != nil {
			fmt.Println(err)
		} else {
			rms, ok := sa.Rms(in)
			if ok && rms >= 0 {
				vol := int(rms)
				fmt.Printf("%d\n", rms)
				count++
				sum += vol
				avg := float64(sum) / float64(count)
				data := newData(vol, avg)
				for _, ch := range channels {
					ch <- data
				}
			}
		}

		select {
		case <-stop:
			fmt.Println("stop portaudio streaming")
			chk(stream.Stop())
			stopped <- 1
			return
		default:
		}
	}
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}

func addVolume(t time.Time, v float64) {
	serie.XValues = append(serie.XValues, t)
	serie.YValues = append(serie.YValues, v)
	if len(serie.YValues) > 30 {
		serie.XValues = serie.XValues[1:]
		serie.YValues = serie.YValues[1:]
	}
}

func collectPeaks(dataCh chan data) {
	ts := newTSData()
	for {
		d := <-dataCh
		globalAvg := d.avg
		ok, v, t := ts.collect(d)
		if ok {
			mux.Lock()
			addVolume(t, v)
			mux.Unlock()
			if v > volumeThreshold*globalAvg || v > maxVolume {
				p := newPeak(t, v, Black, globalAvg, Black)
				logger.Info("peak collected", zap.Time("time", t), zap.Float64("vol", v), zap.Float64("avg", globalAvg))
				mux.Lock()
				peaks = addPeak(peaks, p)
				message = updateMessage()
				avgVolume = globalAvg
				mux.Unlock()
			}
			ts = newTSData()
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
