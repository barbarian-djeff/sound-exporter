package main

import (
	"math"
	"time"
)

func NewSoundAggregator() SoundAggregator {
	return SoundAggregator{
		start: time.Now(),
		count: .0,
		sum:   .0,
	}
}

type SoundAggregator struct {
	start time.Time
	count int64
	sum   float64
}

// Rms calculates the rms :)
func (s *SoundAggregator) Rms(data []int32) (int64, bool) {
	rms := rms(data)
	if time.Since(s.start) > 250000000 { // 1/4s
		avg := int64(rms / float64(s.count))
		s.start = time.Now()
		s.sum = rms
		s.count = 1
		return avg, true
	}
	s.sum += rms
	s.count++
	return 0, false
}

// Rms calculates the rms :)
func rms(data []int32) float64 {
	l := len(data)
	var sum float64
	for _, v := range data {
		f := float64(v)
		sum += f * f
	}
	// fmt.Println("data", sum)
	return math.Sqrt(sum / float64(l))
}
