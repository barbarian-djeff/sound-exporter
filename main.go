package main

import (
	"fmt"
	"os"
	"os/signal"
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
	stop := make(chan int, 1)
	stopped := make(chan int, 1)

	peakChannel := make(chan data)
	minuteChannel := make(chan data)
	go collectMinutes(minuteChannel)
	go collectPeaks(peakChannel)
	go readFromPortAudio(stop, stopped, peakChannel, minuteChannel)
	go serveVolumes()

	<-sig
	fmt.Println("stopping")
	stop <- 1
	<-stopped
	fmt.Println("stopped")
}
