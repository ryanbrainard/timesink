package main

import (
	"context"
	"log"
	cloudevents "github.com/cloudevents/sdk-go"
	"database/sql"
	_ "github.com/lib/pq"
)

func handler(event cloudevents.Event) {
	db, err := sql.Open("postgres", "postgres://postgres:password@timescaledb/cloudeventexplorer?sslmode=disable")
	if err != nil {
		log.Fatal("DB connection error:", err)
	}

	if err := event.Validate(); err == nil {
		raw, err := event.MarshalJSON()
		if err != nil {
			log.Printf("error marshelling the event: %v", err)
		}

		log.Printf("validated event. about to insert")

		_, err = db.Exec("INSERT INTO cloud_events (time, id, type, source, subject, raw) VALUES($1, $2, $3, $4, $5, $6)", event.Time(), event.ID(), event.Type(), event.Source(), event.Subject(), raw)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("inserted event")
	} else {
		log.Printf("error validating the event: %v", err)
	}
}

func main() {
	c, err := cloudevents.NewDefaultClient()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	log.Fatalf("failed to start receiver: %s", c.StartReceiver(context.Background(), handler))
}
