package main

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/hsmtkk/curly-spork/env"
	"github.com/hsmtkk/curly-spork/hybridanalysis"
	"github.com/hsmtkk/curly-spork/myconst"
	"github.com/hsmtkk/curly-spork/reportrepo"
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
	reportClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       myconst.ReportDB,
	})
	defer reportClient.Close()

	natsURL := fmt.Sprintf("nats://%s:%d", natsHost, natsPort)
	natsConn, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("failed to connect NATS; %v", err)
	}
	defer natsConn.Close()

	var reportSummaryGetter reportSummaryGetter = hybridanalysis.NewClient(sugar, apiKey)
	var reportSummaryParser reportSummaryParser = hybridanalysis.NewParser(sugar)
	var reportPutter reportPutter = reportrepo.New(sugar, reportClient)

	ctx := context.Background()

	ch := make(chan *nats.Msg)
	_, err = natsConn.ChanSubscribe(myconst.PollingFalconQueue, ch)
	if err != nil {
		log.Fatalf("failed to subscribe; %v", err)
	}
	for msg := range ch {
		go func(msg *nats.Msg) {
			jobID := string(msg.Data)
			sugar.Infow("poll-work receive", "jobID", jobID)
			reportBytes, err := reportSummaryGetter.ReportSummary(jobID)
			if err != nil {
				sugar.Errorf("failed to get report; %v", err)
				return
			}
			state, sha256, err := reportSummaryParser.ParseReportSummary(reportBytes)
			if err != nil {
				sugar.Errorf("failed to parse report; %v", err)
				return
			}
			sugar.Infow("parsed", "state", state, "sha256", sha256)
			if state == "IN_PROGRESS" || state == "IN_QUEUE" {
				sugar.Infow("requeue", "jobID", jobID)
				if err := natsConn.Publish(myconst.PollingFalconQueue, []byte(jobID)); err != nil {
					sugar.Errorf("failed to publish; %v", err)
					return
				}
			} else {
				sugar.Infow("finish", "jobID", jobID, "sha256", sha256)
				if err := reportPutter.PutReport(ctx, sha256, string(reportBytes)); err != nil {
					sugar.Errorf("failed to put report; %v", err)
					return
				}
			}
		}(msg)
	}
}

type reportSummaryGetter interface {
	ReportSummary(jobID string) ([]byte, error)
}

type reportSummaryParser interface {
	ParseReportSummary(respBytes []byte) (string, string, error)
}

type reportPutter interface {
	PutReport(ctx context.Context, sha256, report string) error
}
