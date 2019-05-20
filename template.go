package main

import "time"

type Color string

const (
	Blue  Color = "color: #0000ff;"
	Red   Color = "color: #ff0000;"
	Green Color = "color: #00ff00;"
	Black Color = ""
)

type Peak struct {
	Time    string
	Current Volume
	Average Average
}

func newPeak(t time.Time, curVol int, curCol Color, avgVol float64, avgCol Color) Peak {
	return Peak{
		t.Format("15:04:05"),
		Volume{curVol, curCol},
		Average{avgVol, avgCol},
	}
}

type Minute struct {
	Time    string
	Average Average
}

func newMinute(t time.Time, a float64, c Color) Minute {
	return Minute{
		t.Format("15:04"),
		Average{a, c},
	}
}

type TemplateData struct {
	Message          string
	MaxAverageVolume int
	VolumeThreshold  float64
	Peaks            []Peak
	Minutes          []Minute
}

type Volume struct {
	Value int
	Color Color
}

type Average struct {
	Value float64
	Color Color
}
