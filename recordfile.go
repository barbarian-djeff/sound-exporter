package main

import (
	"encoding/csv"
	"fmt"
	"go.uber.org/zap"
	"os"
)

func writeFile(dataCh chan data, writers []*os.File) {
	o := make([]*csv.Writer, len(writers))
	for i, w := range writers {
		o[i] = csv.NewWriter(w)
	}
	for {
		d := <-dataCh
		line := []string{
			fmt.Sprintf("%s", d.time.String()),
			fmt.Sprintf("%d", d.volume),
			fmt.Sprintf("%f", d.avg),
		}
		for _, w := range o {
			if err := w.Write(line); err != nil {
				logger.Error("fail to write in csv file", zap.Error(err))
			}
			w.Flush()
		}
	}
}
