package main

import (
	"fmt"
	"log"

	"github.com/hsmtkk/curly-spork/env"
	"github.com/hsmtkk/curly-spork/myconst"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func main() {
	natsHost := env.RequiredString("NATS_HOST")
	natsPort := env.RequiredInt("NATS_PORT")

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	natsURL := fmt.Sprintf("nats://%s:%d", natsHost, natsPort)
	natsConn, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("failed to connect NATS; %v", err)
	}
	defer natsConn.Close()

	ch := make(chan *nats.Msg)
	_, err = natsConn.ChanSubscribe(myconst.WaitingScanQueue, ch)
	if err != nil {
		log.Fatalf("failed to subscribe; %v", err)
	}
	for msg := range ch {
		sha256 := string(msg.Data)
		sugar.Infow("scan-work receive", "sha256", sha256)
		// TODO: submit to hybrid analysis
		taskID := sha256
		sugar.Infow("scan-work send", "taskID", taskID)
		if err := natsConn.Publish(myconst.PollingFalconQueue, []byte(taskID)); err != nil {
			sugar.Errorf("failed to publish; %v", err)
		}
	}
}
