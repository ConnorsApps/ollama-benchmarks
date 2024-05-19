package main

import (
	"encoding/csv"
	"io"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

type Run struct {
	Model               *Model
	InitialLoadDuration time.Duration
	AvgDuration         time.Duration
	MaxDuration         time.Duration
	MinDuration         time.Duration
	TotalTestTime       time.Duration
}

func (r *Run) JoinAndAverage(newRun *Run) {
	r.InitialLoadDuration = r.InitialLoadDuration + newRun.InitialLoadDuration/2
	r.AvgDuration = r.AvgDuration + newRun.AvgDuration/2
	r.MaxDuration = r.MaxDuration + newRun.MaxDuration/2
	r.MinDuration = r.MinDuration + newRun.MinDuration/2
	r.TotalTestTime = r.TotalTestTime + newRun.TotalTestTime/2
}

func (r *Run) PrintResult() {
	log.Info().
		Str("avg", r.AvgDuration.String()).
		Str("max", r.MaxDuration.String()).
		Str("min", r.MinDuration.String()).
		Str("init", r.InitialLoadDuration.String()).
		Send()
}

type Runs []*Run

var runCols = []string{
	"model",
	"param_count",
	"storage",
	"initial_load_duration",
	"avg_duration",
	"max_duration",
	"min_duration",
	"total_test_time",
}

func (runs Runs) ToCSV() io.Reader {
	read, write := io.Pipe()
	go func() {
		defer write.Close()
		w := csv.NewWriter(write)
		must(w.Write(runCols))

		for _, run := range runs {
			w.Write([]string{
				run.Model.Name,
				run.Model.ParamCount,
				run.Model.Storage,
				strconv.Itoa(int(run.InitialLoadDuration.Milliseconds())),
				strconv.Itoa(int(run.AvgDuration.Milliseconds())),
				strconv.Itoa(int(run.MaxDuration.Milliseconds())),
				strconv.Itoa(int(run.MinDuration.Milliseconds())),
				strconv.Itoa(int(run.TotalTestTime.Milliseconds())),
			})
		}
		w.Flush()
	}()
	return read
}
