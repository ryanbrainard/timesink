package timesink

import (
	"database/sql"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/kinds"
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
					Type: graphql.ID,
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
	query := "SELECT raw FROM cloud_events WHERE id = $1"
	row := q.db.QueryRow(query, id)

	var raw []byte
	if err := row.Scan(&raw); err != nil {
		q.logger.WithError(err).Error("db query error")
		return nil, err
	}

	event := cloudevents.NewEvent()
	if err := event.UnmarshalJSON(raw); err != nil {
		return nil, err
	}

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
		"data": &graphql.Field{
			Type: JsonType,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				var data map[string]interface{}
				if err := params.Source.(*cloudevents.Event).DataAs(&data); err != nil {
					return nil, err
				}
				return data, nil
			},
		},
	},
})

var JsonType = graphql.NewScalar(
	graphql.ScalarConfig{
		Name:        "JSON",
		Description: "The `JSON` scalar type represents JSON values as specified by [ECMA-404](http://www.ecma-international.org/publications/files/ECMA-ST/ECMA-404.pdf)",
		Serialize: func(value interface{}) interface{} {
			return value
		},
		ParseValue: func(value interface{}) interface{} {
			return value
		},
		ParseLiteral: parseJsonLiteral,
	},
)

func parseJsonLiteral(astValue ast.Value) interface{} {
	kind := astValue.GetKind()

	switch kind {
	case kinds.StringValue:
		return astValue.GetValue()
	case kinds.BooleanValue:
		return astValue.GetValue()
	case kinds.IntValue:
		return astValue.GetValue()
	case kinds.FloatValue:
		return astValue.GetValue()
	case kinds.ObjectValue:
		obj := make(map[string]interface{})
		for _, v := range astValue.GetValue().([]*ast.ObjectField) {
			obj[v.Name.Value] = parseJsonLiteral(v.Value)
		}
		return obj
	case kinds.ListValue:
		list := make([]interface{}, 0)
		for _, v := range astValue.GetValue().([]ast.Value) {
			list = append(list, parseJsonLiteral(v))
		}
		return list
	default:
		return nil
	}
}
