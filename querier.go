package timesink

import (
	"github.com/graphql-go/graphql"
)

type Event struct {
	ID      string `json:"id"`
	Subject string `json:"subject"`
}

var EventList []Event

func init() {
	todo1 := Event{ID: "a", Subject: "A todo not to forget"}
	todo2 := Event{ID: "b", Subject: "This is the most important"}
	todo3 := Event{ID: "c", Subject: "Please do this or else"}
	EventList = append(EventList, todo1, todo2, todo3)
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
		},
		"subject": &graphql.Field{
			Type: graphql.String,
		},
	},
})

func SchemaConfig() graphql.SchemaConfig {
	// Schema
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

				idQuery, isOK := params.Args["id"].(string)
				if isOK {
					// Search for el with id
					for _, todo := range EventList {
						if todo.ID == idQuery {
							return todo, nil
						}
					}
				}

				return Event{}, nil
			},
		},
	}
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	return graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
}
