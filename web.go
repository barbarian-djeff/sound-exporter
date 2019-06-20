package main

import (
	"github.com/wcharczuk/go-chart"
	"go.uber.org/zap"
	"html/template"
	"net/http"
	"time"
)

func serveVolumes() {
	http.HandleFunc("/chart.png", drawChart)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("./html/index.html"))
		mux.Lock()
		defer mux.Unlock()
		data := TemplateData{
			message,
			maxVolume,
			avgVolume,
			volumeThreshold,
			peaks,
			minutes,
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, data)
	})
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./html/css/"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./html/js/"))))
	logger.Fatal("fail to serve", zap.Error(http.ListenAndServe("localhost:8090", nil)))
}

func drawChart(res http.ResponseWriter, req *http.Request) {
	graph := chart.Chart{
		XAxis: chart.XAxis{
			Style:          chart.StyleShow(),
			ValueFormatter: hmsFormatter,
		},
		YAxis: chart.YAxis{
			Style: chart.StyleShow(),
		},
		Series: []chart.Series{
			serie,
		},
	}
	res.Header().Set("Content-Type", "image/png")
	graph.Render(chart.PNG, res)
}

func hmsFormatter(v interface{}) string {
	if typed, isTyped := v.(float64); isTyped {
		return time.Unix(0, int64(typed)).Format("15:04:05")
	}
	return "x"
}
