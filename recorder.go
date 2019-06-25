package timesink

import (
	"database/sql"
	cloudevents "github.com/cloudevents/sdk-go"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type recorder struct {
	db     *sql.DB
	logger *logrus.Entry
}

func NewRecorder(databaseUrl string, logger *logrus.Entry) *recorder {
	l := logger.WithField("component", "Recorder")

	db, err := sql.Open("postgres", databaseUrl)
	if err != nil {
		l.Fatal("DB connection error:", err)
	}

	return &recorder{
		db:     db,
		logger: l,
	}
}

func (r *recorder) HandleEvent(event cloudevents.Event) {
	l := r.logger.WithField("fn", "HandleEvent").WithField("id", event.ID())
	l.WithField("at", "start").Info()
	defer l.WithField("at", "finish").Info()

	if err := event.Validate(); err != nil {
		r.logger.WithError(err).Error("validation error")
		return
	}

	raw, err := event.MarshalJSON()
	if err != nil {
		r.logger.WithError(err).Error("json marshall error")
		return
	}

	// TODO: move to preparedstatement
	insert := "INSERT INTO cloud_events (time, id, type, source, subject, raw) " +
		"VALUES($1, $2, $3, $4, $5, $6)"

	_, err = r.db.Exec(insert,
		event.Time(),
		event.ID(),
		event.Type(),
		event.Source(),
		event.Subject(),
		raw,
	)
	if err != nil {
		r.logger.WithError(err).Error("db insert error")
		return
	}
}
