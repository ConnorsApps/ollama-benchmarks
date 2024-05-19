package main

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/ollama/ollama/api"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	_ "github.com/joho/godotenv/autoload"
)

const (
	minioApi    = "minio-api.connorskees.com"
	minioBucket = "youtube"
)

type Model struct {
	Name       string
	ParamCount string
	Storage    string
}

var (
	ollamaApi, _ = url.Parse("https://ollama-api.connorskees.com")
	models       = []*Model{
		{Name: "codellama:7b", ParamCount: "7B", Storage: "3.8GB"},
		{Name: "llama2-uncensored:7b", ParamCount: "7B", Storage: "3.8GB"},
		{Name: "llama3:latest", ParamCount: "8B", Storage: "4.7GB"},
		{Name: "tinyllama:latest", ParamCount: "1.1B", Storage: "640MB"},
	}
	questions = []string{
		"What is the capital of France?",
		"Who is the CEO of Apple?",
		"What is the meaning of life?",
		"What is the best way to get to work?",
		"How do I make a cup of coffee?",
		"What is the capital of Australia?",
		"Can you tell me about a famous painting by Leonardo da Vinci?",
		"Who is the CEO of Tesla?",
		"Can you give me a recipe for a delicious vegetarian dish?",
		"What is the capital of Canada?",
		"How do I properly format my phone number for international calls?",
		"What is the best way to stay healthy during the pandemic?",
	}
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Kitchen})
}

func main() {
	// Initialize minio client object.
	minioClient, err := minio.New(minioApi, &minio.Options{
		Creds: credentials.NewStaticV4(
			os.Getenv("MINIO_KEY_ID"), os.Getenv("MINIO_SECRET_KEY"), "",
		),
		Secure: true,
	})
	must(err)

	var (
		ollama         = api.NewClient(ollamaApi, http.DefaultClient)
		c              = New(ollama)
		runs           = make(Runs, len(models))
		done, shutdown = make(chan interface{}), make(chan os.Signal, 1)
	)

	log.Info().Timestamp().Msg("Start")

	signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		<-shutdown
		log.Info().Timestamp().Msg("End")

		var filename = "run" + time.Now().Format(time.RFC3339) + ".csv"
		_, err = minioClient.PutObject(context.Background(), minioBucket, filename, runs.ToCSV(), -1, minio.PutObjectOptions{})
		must(err)
		log.Info().Str("filename", filename).Msg("Uploaded to Minio")
		done <- nil
	}()

	color.New(color.BgBlack).Add(color.FgWhite).Println("--- Round 1 ---")

	for i, model := range models {
		run := c.TestModel(model)
		run.PrintResult()
		runs[i] = run
	}

	color.New(color.BgBlack).Add(color.FgWhite).Println("--- Round 2 ---")

	for i, model := range models {
		run := c.TestModel(model)
		run.PrintResult()
		// Average the two runs together
		runs[i].JoinAndAverage(run)
	}

	shutdown <- nil
	<-done
}
