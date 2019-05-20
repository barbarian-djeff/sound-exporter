package main

import (
	"html/template"
	"time"
)

const (
	Blue  template.CSS = "color: #0000ff;"
	Red   template.CSS = "color: #ff0000;"
	Green template.CSS = "color: #00ff00;"
	Black template.CSS = ""
)

type Peak struct {
	Time    string
	Current Volume
	Average Average
}

func newPeak(t time.Time, curVol int, curCol template.CSS, avgVol float64, avgCol template.CSS) Peak {
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

func newMinute(t time.Time, a float64, c template.CSS) Minute {
	return Minute{
		t.Format("15:04"),
		Average{a, c},
	}
}

type TemplateData struct {
	Message          string
	MaxAverageVolume float64
	VolumeThreshold  float64
	Peaks            []Peak
	Minutes          []Minute
}

type Volume struct {
	Value int
	Color template.CSS
}

type Average struct {
	Value float64
	Color template.CSS
}
