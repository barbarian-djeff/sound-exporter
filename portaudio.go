package main

import (
	"fmt"
	"github.com/gordonklaus/portaudio"
	"os"
	"os/signal"
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

func readFromPortAudio(sync portaudioSync, channels ...chan data) {
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

		if sync.stopStream(stream) {
			return
		}
	}
}

type portaudioSync struct {
	sig     chan os.Signal
	stop    chan int
	stopped chan int
}

func newSync() portaudioSync {
	s := portaudioSync{
		sig:     make(chan os.Signal, 1),
		stop:    make(chan int, 1),
		stopped: make(chan int, 1),
	}
	signal.Notify(s.sig, os.Interrupt, os.Kill)
	return s
}

func (s portaudioSync) waitInterrupt() {
	<-s.sig
	s.stop <- 1 // say to portaudio to close the stream
	<-s.stopped // wait for the stream to close
}

func (s portaudioSync) stopStream(str *portaudio.Stream) bool {
	select {
	case <-s.stop:
		fmt.Println("stop portaudio streaming")
		chk(str.Stop())
		s.stopped <- 1
		return true
	default:
	}
	return false
}
