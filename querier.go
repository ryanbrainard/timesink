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

func (q *querier) queryEventById(id string) (*cloudevents.Event, error) {
	query := "SELECT raw FROM cloud_events WHERE id = $1"
	row := q.db.QueryRow(query, id)
	return q.scanEventRow(row)
}

func (q *querier) queryLastEventByGroupVersionKindName(apiVersion, kind, name string) (*cloudevents.Event, error) {
	query := "SELECT raw FROM cloud_events WHERE " +
		"raw -> 'data' ->> 'apiVersion'          = $1 AND " +
		"raw -> 'data' ->> 'kind'                = $2 AND " +
		"raw -> 'data' ->  'metadata' ->> 'name' = $3" +
		"ORDER BY time DESC " + // TODO: how to get correct one at point in time?
		"LIMIT 1"
	row := q.db.QueryRow(query, apiVersion, kind, name)
	return q.scanEventRow(row)
}

func (q *querier) scanEventRow(row *sql.Row) (*cloudevents.Event, error) {
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

func eventParamAsMap(params graphql.ResolveParams) (map[string]interface{}, error) {
	var data map[string]interface{}
	if err := params.Source.(*cloudevents.Event).DataAs(&data); err != nil {
		return nil, err
	}
	return data, nil
}

func SchemaConfig() graphql.SchemaConfig {
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
				return params.Context.Value("q").(*querier).queryEventById(id)
			},
		},
	}
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	return graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
}

var eventType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Event",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.ID,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				return params.Source.(*cloudevents.Event).ID(), nil
			},
		},
		"type": &graphql.Field{
			Type: graphql.String,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				return params.Source.(*cloudevents.Event).Type(), nil
			},
		},
		"source": &graphql.Field{
			Type: graphql.String,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				return params.Source.(*cloudevents.Event).Source(), nil
			},
		},
		"subject": &graphql.Field{
			Type: graphql.String,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				return params.Source.(*cloudevents.Event).Subject(), nil
			},
		},
		"time": &graphql.Field{
			Type: graphql.DateTime,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				return params.Source.(*cloudevents.Event).Time(), nil
			},
		},
		"data": &graphql.Field{
			Type: ObjectLiteralType,
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

func init() {
	eventType.AddFieldConfig("owners", &graphql.Field{
		Type: graphql.NewList(eventType),
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			data, err := eventParamAsMap(params)
			ownerRefs := data["metadata"].(map[string]interface{})["ownerReferences"].([]interface{})
			var owners []*cloudevents.Event
			for _, ownerRef := range ownerRefs {
				owner, err := params.Context.Value("q").(*querier).queryLastEventByGroupVersionKindName(
					ownerRef.(map[string]interface{})["apiVersion"].(string),
					ownerRef.(map[string]interface{})["kind"].(string),
					ownerRef.(map[string]interface{})["name"].(string),
				)
				if err != nil {
					return nil, err
				}
				owners = append(owners, owner)
			}
			return owners, err
		},
	})
}

var ObjectLiteralType = graphql.NewScalar(
	graphql.ScalarConfig{
		Name:        "JSON",
		Description: "The `JSON` scalar type represents JSON values as specified by [ECMA-404](http://www.ecma-international.org/publications/files/ECMA-ST/ECMA-404.pdf)",
		Serialize: func(value interface{}) interface{} {
			return value
		},
		ParseValue: func(value interface{}) interface{} {
			return value
		},
		ParseLiteral: parseObjectLiteral,
	},
)

func parseObjectLiteral(astValue ast.Value) interface{} {
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
			obj[v.Name.Value] = parseObjectLiteral(v.Value)
		}
		return obj
	case kinds.ListValue:
		list := make([]interface{}, 0)
		for _, v := range astValue.GetValue().([]ast.Value) {
			list = append(list, parseObjectLiteral(v))
		}
		return list
	default:
		return nil
	}
}
