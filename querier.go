package timesink

import (
	"database/sql"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/graphql-go/graphql"
	"github.com/sirupsen/logrus"
)

func NewQuerier(databaseUrl string, logger *logrus.Entry) *querier {
	l := logger.WithField("component", "Querier")

	db, err := sql.Open("postgres", databaseUrl)
	if err != nil {
		l.Fatal("DB connection error:", err)
	}

	return &querier{
		db:     db,
		logger: l,
	}
}

type querier struct {
	db     *sql.DB
	logger *logrus.Entry
}

func (q *querier) SchemaConfig() graphql.SchemaConfig {
	fields := graphql.Fields{
		"hello": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return "world", nil
			},
		},
		"event": &graphql.Field{
			Type:        eventType,
			Description: "Get single event",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				id, ok := params.Args["id"].(string)
				if !ok {
					return nil, nil
				}
				return q.queryEvent(id)
			},
		},
	}
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	return graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
}

func (q *querier) queryEvent(id string) (*cloudevents.Event, error) {
	query := "SELECT id, subject FROM cloud_events WHERE id = $1"
	row := q.db.QueryRow(query, id)

	var qId string
	var qSubject string
	if err := row.Scan(&qId, &qSubject); err != nil {
		q.logger.WithError(err).Error("db query error")
		return nil, err
	}

	event := cloudevents.NewEvent()
	event.SetID(qId)
	event.SetSubject(qSubject)

	return &event, nil
}

// define custom GraphQL ObjectType `eventType` for our Golang struct `Event`
// Note that
// - the fields in our eventType maps with the json tags for the fields in our struct
// - the field type matches the field type in our struct
var eventType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Event",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				return params.Source.(*cloudevents.Event).ID(), nil
			},
		},
		"subject": &graphql.Field{
			Type: graphql.String,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				return params.Source.(*cloudevents.Event).Subject(), nil
			},
		},
	},
})
