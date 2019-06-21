package main

import (
	"flag"
	"os"
	"os/signal"
)

func main() {
	fileName := flag.String("record", "", "record data into this filename")
	flag.Parse()
	f := openFiles(*fileName)
	defer closeFiles(f)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
	stop := make(chan int, 1)
	stopped := make(chan int, 1)

	peakChannel := make(chan data)
	minuteChannel := make(chan data)
	fileChannel := make(chan data)

	go collectMinutes(minuteChannel)
	go collectPeaks(peakChannel)
	go writeFile(fileChannel, f)
	go readFromPortAudio(stop, stopped, peakChannel, minuteChannel, fileChannel)
	go serveVolumes()

	<-sig
	stop <- 1 // say to portaudio to close the stream
	<-stopped // wait for the stream to close
}
