package main

import (
	"context"
	cloudevents "github.com/cloudevents/sdk-go"
	_ "github.com/lib/pq"
	"go.ryanbrainard.com/timesink/recorder"
	"log"
	"os"
)

func main() {
	databaseUrl, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		panic("DATABASE_URL not set")
	}
	r := recorder.New(databaseUrl)

	c, err := cloudevents.NewDefaultClient()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	if err := c.StartReceiver(context.Background(), r.HandleEvent); err != nil {
		log.Fatalf("failed to start receiver: %s", err)
	}
}
