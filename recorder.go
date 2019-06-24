package timesink

import (
	"database/sql"
	cloudevents "github.com/cloudevents/sdk-go"
	_ "github.com/lib/pq"
	"log"
)

type recorder struct {
	db *sql.DB
}

func NewRecorder(databaseUrl string) *recorder {
	db, err := sql.Open("postgres", databaseUrl)
	if err != nil {
		log.Fatal("DB connection error:", err)
	}

	return &recorder{db: db}
}

func (h *recorder) HandleEvent(event cloudevents.Event) {
	log.Printf("component=recorder.HandleEvent id=%q at=start", event.ID())

	if err := event.Validate(); err != nil {
		log.Printf("component=recorder.HandleEvent id=%q at=validation.error err=%q", event.ID(), err)
	}

	raw, err := event.MarshalJSON()
	if err != nil {
		log.Printf("component=recorder.HandleEvent id=%q at=json.error err=%q", event.ID(), err)
	}

	// TODO: move to preparedstatement
	sql := "INSERT INTO cloud_events (time, id, type, source, subject, raw) VALUES($1, $2, $3, $4, $5, $6)"

	_, err = h.db.Exec(sql,
		event.Time(),
		event.ID(),
		event.Type(),
		event.Source(),
		event.Subject(),
		raw,
	)
	if err != nil {
		log.Printf("component=recorder.HandleEvent id=%q at=insert.error err=%q", event.ID(), err)
	}

	log.Printf("component=recorder.HandleEvent id=%q at=finish", event.ID())
}
