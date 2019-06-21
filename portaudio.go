package main

import (
	"fmt"
	"github.com/gordonklaus/portaudio"
	"time"
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
