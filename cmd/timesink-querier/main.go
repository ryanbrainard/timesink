package main

import (
	"context"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/sirupsen/logrus"
	"go.ryanbrainard.com/timesink"
	"net/http"
	"os"
)

func main() {
	logger := logrus.New().WithField("service", "timesink-recorder")

	databaseUrl, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		logger.Panic("DATABASE_URL not set")
	}
	q := timesink.NewQuerier(databaseUrl, logger)

	schema, err := graphql.NewSchema(timesink.SchemaConfig())
	if err != nil {
		panic(err)
	}

	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		h.ContextHandler(context.WithValue(context.Background(), "q", q), w, r)
	})

	if err := http.ListenAndServe("127.0.0.1:8080", nil); err != nil {
		panic(err)
	}
}
