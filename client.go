package main

import (
	"context"
	"time"

	"github.com/fatih/color"
	"github.com/ollama/ollama/api"
	"github.com/rs/zerolog/log"
)

func New(ollama *api.Client) *Client {
	return &Client{
		ollama: ollama,
		color:  color.New(color.FgHiCyan),
	}
}

type Client struct {
	ollama *api.Client
	color  *color.Color
}

var keepAlive = &api.Duration{Duration: time.Second * 30}

func (c *Client) prompt(model, msg string) *api.ChatResponse {
	var res = make(chan *api.ChatResponse)
	must(
		c.ollama.Chat(context.Background(), &api.ChatRequest{
			Model:     model,
			Stream:    ptr(false),
			KeepAlive: keepAlive,
			Messages: []api.Message{
				{
					Role:    "user",
					Content: msg,
				},
			},
		}, func(cr api.ChatResponse) error {
			go func() {
				res <- &cr
			}()
			return nil
		}),
	)

	return <-res
}

func (c *Client) TestModel(model *Model) *Run {
	var (
		start            = time.Now()
		run              = &Run{Model: model}
		allQuestionsTook time.Duration
	)
	// Load model into memory with simple first query
	log.Info().Str("model", model.Name).Str("size", model.Storage).Msg("Loading into memory...")
	run.InitialLoadDuration = c.prompt(model.Name, "hi").Metrics.LoadDuration

	for _, question := range questions {
		println(question)
		res := c.prompt(model.Name, question)

		allQuestionsTook += res.Metrics.TotalDuration

		if res.TotalDuration > run.MaxDuration {
			run.MaxDuration = res.TotalDuration
		}

		if run.MinDuration > res.TotalDuration || run.MinDuration == 0 {
			run.MinDuration = res.TotalDuration
		}

		c.color.Println("🤖: " + res.Message.Content)
	}

	run.AvgDuration = time.Duration(int64(allQuestionsTook) / int64(len(questions)))
	run.TotalTestTime = time.Since(start)

	return run
}
