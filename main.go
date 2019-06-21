package main

import (
	"flag"
)

func main() {
	fileName := flag.String("record", "", "record data into this filename")
	flag.Parse()
	f := openFiles(*fileName)
	defer closeFiles(f)

	peakChannel := make(chan data)
	minuteChannel := make(chan data)
	fileChannel := make(chan data)

	go collectMinutes(minuteChannel)
	go collectPeaks(peakChannel)
	go writeFile(fileChannel, f)
	go serveVolumes()

	s := newSync()
	go readFromPortAudio(s, peakChannel, minuteChannel, fileChannel)
	s.waitInterrupt()
}
