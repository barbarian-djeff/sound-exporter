package main

import (
	"github.com/wcharczuk/go-chart"
	"go.uber.org/zap"
	"regexp"
	"sync"
	"time"
)

var (
	re              = regexp.MustCompile(`\s*(\d*)\s*`)
	logger, _       = zap.NewDevelopment()
	maxVolume       = 4000.
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
		// receive 4 rms per sec but merge them into one sec for peak detection
		ok, v, t := ts.collect(d)
		if ok {
			ts = newTSData()
			mux.Lock()
			// add to the serie for display in the chart
			addVolume(t, v)
			mux.Unlock()

			if v > volumeThreshold*globalAvg || v > maxVolume {
				// peak detectec
				p := newPeak(t, v, Black, globalAvg, Black)
				logger.Info("peak collected", zap.Time("time", t), zap.Float64("vol", v), zap.Float64("avg", globalAvg))
				mux.Lock()
				peaks = addPeak(peaks, p)
				message = updateMessage()
				avgVolume = globalAvg
				mux.Unlock()
			}
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
	var mSum int64
	var mCount int64
	mCount = 1

	for {
		d := <-dataCh

		m := d.time.Minute()
		if currentMinute == -1 {
			currentMinute = m
		}
		if m != currentMinute {
			nm := newMinute(currentMinute, float64(mSum)/float64(mCount), Black)
			logger.Info("count per minute", zap.Int64("count", mCount))
			// store info
			mux.Lock()
			minutes = addMinute(minutes, nm)
			mux.Unlock()
			logger.Info("minute collected", zap.String("time", nm.Time), zap.Float64("avg", nm.Average.Value))

			// next minute: reset counters
			currentMinute = m
			mSum = d.volume
			mCount = 1
		} else {
			mSum += d.volume
			mCount++
		}
	}
}
