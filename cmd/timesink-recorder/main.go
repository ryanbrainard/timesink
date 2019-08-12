package main

import (
	"context"
	cloudevents "github.com/cloudevents/sdk-go"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"go.ryanbrainard.com/timesink"
	"os"
)

func main() {
	logger := logrus.New().WithField("service", "timesink-recorder")

	databaseUrl, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		logger.Panic("DATABASE_URL not set")
	}
	databaseUrl = os.ExpandEnv(databaseUrl)
	r := timesink.NewRecorder(databaseUrl, logger)

	c, err := cloudevents.NewDefaultClient()
	if err != nil {
		logger.WithError(err).Fatal("failed to create client")
	}

	if err := c.StartReceiver(context.Background(), r.HandleEvent); err != nil {
		logger.WithError(err).Fatal("failed to start receiver")
	}
}
