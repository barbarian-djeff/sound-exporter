package main

import (
	"fmt"
	"time"
)

const tsCollectingTimeSec = 1

type timeSerieData struct {
	start time.Time
	total int
	count int
}

func newTSData() timeSerieData {
	return timeSerieData{time.Now(), 0, 0}
}

func (d *timeSerieData) collect(data data) (ready bool, volume float64, end time.Time) {
	elapsed := data.time.Sub(d.start)
	if elapsed.Seconds() > tsCollectingTimeSec {
		avg := float64(d.total) / float64(d.count)
		logger.Info("elapsed " + fmt.Sprintf("%f", avg))
		return true, avg, data.time
	}
	d.count++
	d.total += data.volume
	return false, 0, data.time
}
