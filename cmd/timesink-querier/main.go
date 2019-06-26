package main

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"go.ryanbrainard.com/timesink"
	"net/http"
)

func main() {
	schema, _ := graphql.NewSchema(timesink.SchemaConfig())

	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	http.Handle("/graphql", h)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
