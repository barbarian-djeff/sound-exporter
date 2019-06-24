package main

import (
	"math"
	"time"
)

func NewSoundAggregator() SoundAggregator {
	return SoundAggregator{
		start: time.Now(),
		sum:   .0,
	}
}

type SoundAggregator struct {
	start time.Time
	sum   float64
}

// Rms calculates the rms :)
func (s *SoundAggregator) Rms(data []int32) (int64, bool) {
	rms := rms(data)
	if time.Since(s.start) > 250000000 { // 1/4s
		res := s.sum
		s.start = time.Now()
		s.sum = rms
		return int64(res / 1000000.), true
	}
	s.sum += rms
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
