package main

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis"
	"github.com/hsmtkk/curly-spork/env"
	"github.com/hsmtkk/curly-spork/filerepo"
	"github.com/hsmtkk/curly-spork/hybridanalysis"
	"github.com/hsmtkk/curly-spork/myconst"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func main() {
	redisHost := env.RequiredString("REDIS_HOST")
	redisPort := env.RequiredInt("REDIS_PORT")
	natsHost := env.RequiredString("NATS_HOST")
	natsPort := env.RequiredInt("NATS_PORT")
	apiKey := env.RequiredString("API_KEY")

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	redisAddr := fmt.Sprintf("%s:%d", redisHost, redisPort)
	fileClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       myconst.FileDB,
	})
	defer fileClient.Close()

	natsURL := fmt.Sprintf("nats://%s:%d", natsHost, natsPort)
	natsConn, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("failed to connect NATS; %v", err)
	}
	defer natsConn.Close()

	var fileGetter fileGetter = filerepo.New(sugar, fileClient)
	var fileSubmitter fileSubmitter = hybridanalysis.NewClient(sugar, apiKey)
	var submitFileParser submitFileParser = hybridanalysis.NewParser(sugar)

	ctx := context.Background()

	ch := make(chan *nats.Msg)
	_, err = natsConn.ChanSubscribe(myconst.WaitingScanQueue, ch)
	if err != nil {
		log.Fatalf("failed to subscribe; %v", err)
	}
	for msg := range ch {
		sha256 := string(msg.Data)
		sugar.Infow("scan-work receive", "sha256", sha256)
		data, err := fileGetter.GetFile(ctx, sha256)
		if err != nil {
			sugar.Errorf("failed to get file; %v", err)
		}
		respBytes, err := fileSubmitter.SubmitFile(sha256, data)
		if err != nil {
			sugar.Errorf("failed to submit; %v", err)
		}
		taskID, err := submitFileParser.ParseSubmitFile(respBytes)
		if err != nil {
			sugar.Errorf("failed to parse; %v", err)
		}
		sugar.Infow("scan-work send", "taskID", taskID)
		if err := natsConn.Publish(myconst.PollingFalconQueue, []byte(taskID)); err != nil {
			sugar.Errorf("failed to publish; %v", err)
		}
	}
}

type fileGetter interface {
	GetFile(ctx context.Context, sha256 string) ([]byte, error)
}

type fileSubmitter interface {
	SubmitFile(fileName string, data []byte) ([]byte, error)
}

type submitFileParser interface {
	ParseSubmitFile(respBytes []byte) (string, error)
}
