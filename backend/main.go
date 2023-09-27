package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/graphql-go/graphql"
)

type Developer struct {
	ID        int64     `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	BirthDate time.Time `json:"birth_date"`
	GithubURL string    `json:"github_url,omitempty"`
	Stack     []string  `json:"stack,omitempty"`
}

// We define a type for the object we want to return
var developerType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Developer",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
			},
			"first_name": &graphql.Field{
				Type: graphql.String,
			},
			"last_name": &graphql.Field{
				Type: graphql.String,
			},
			"birth_date": &graphql.Field{
				Type: graphql.DateTime,
			},
			"github_url": &graphql.Field{
				Type: graphql.String,
			},
			"stack": &graphql.Field{
				Type: graphql.NewList(graphql.String),
			},
		},
	},
)

func main() {
	// Step 1: define the query fields
	fields := graphql.Fields{
		"developers": &graphql.Field{
			Type:        graphql.NewList(developerType),
			Description: "Get a list of developers",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.Int,
				},
				"name": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				// Here we code how we get the resources
				developer := Developer{
					ID:        1,
					FirstName: "Caio",
					LastName:  "Teixeira",
					BirthDate: time.Now(),
					GithubURL: "https://github.com/CaioTeixeira95",
				}
				return []Developer{developer}, nil
			},
		},
		"developer": &graphql.Field{
			Type:        developerType,
			Description: "Get a single developers",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.Int,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return Developer{
					ID:        1,
					FirstName: "Caio",
					LastName:  "Teixeira",
					BirthDate: time.Now(),
					GithubURL: "https://github.com/CaioTeixeira95",
				}, nil
			},
		},
	}

	// Step 2: create a query
	query := graphql.NewObject(graphql.ObjectConfig{Name: "Query", Fields: fields})

	// Step 3: create a schema config and create a schema
	schemaConfig := graphql.SchemaConfig{Query: query}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	http.HandleFunc("/graphql", func(rw http.ResponseWriter, req *http.Request) {
		query := req.URL.Query().Get("query")

		// Step 4: execute the query based on the query string sent by the client. If empty it returns error.
		res := graphql.Do(graphql.Params{
			Schema:        schema,
			RequestString: query,
		})

		rw.Header().Set("Content-Type", "application/json")
		if len(res.Errors) > 0 {
			rw.WriteHeader(http.StatusBadRequest)
		}

		if err := json.NewEncoder(rw).Encode(res); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	log.Println("server running at http://localhost:8000")
	http.ListenAndServe(":8000", nil)
}
