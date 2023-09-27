package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/graphql-go/graphql"
	_ "github.com/lib/pq"
)

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
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatalf("DATABASE_URL is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	repository := NewDefaultDeveloperRepository(db)

	// Step 1: define the query and mutation fields
	queryFields := graphql.Fields{
		// http://localhost:8000/graphql?query={developers{first_name,last_name,stack}}
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
			Resolve: ListDevelopers(repository),
		},
		// http://localhost:8000/graphql?query={developer(id:1){first_name,last_name,stack}}
		"developer": &graphql.Field{
			Type:        developerType,
			Description: "Get a single developers",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.Int,
				},
			},
			Resolve: DeveloperDetails(repository),
		},
	}

	mutationFields := graphql.Fields{
		// http://localhost:8000/graphql?query=mutation+_{create(first_name:"Caio",last_name:"Teixeira",github_url:"https://github.com/CaioTeixeira95",stack:["go", "python"]){id,first_name,last_name,github_url,stack}}
		"create": &graphql.Field{
			Type:        developerType,
			Description: "Create new developer",
			Args: graphql.FieldConfigArgument{
				"first_name": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"last_name": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"github_url": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"stack": &graphql.ArgumentConfig{
					Type: graphql.NewList(graphql.String),
				},
			},
			Resolve: CreateDeveloper(repository),
		},
		// http://localhost:8000/graphql?query=mutation+_{update(id:1,first_name:"Caio",last_name:"Teixeira",github_url:"https://github.com/CaioTeixeira95",stack:["go", "python"]){id,first_name,last_name,github_url,stack}}
		"update": &graphql.Field{
			Type:        developerType,
			Description: "Update a developer",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.Int,
				},
				"first_name": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"last_name": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"github_url": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"stack": &graphql.ArgumentConfig{
					Type: graphql.NewList(graphql.String),
				},
			},
			Resolve: UpdateDeveloper(repository),
		},
	}

	// Step 2: create a query and mutation
	query := graphql.NewObject(graphql.ObjectConfig{Name: "Query", Fields: queryFields})
	mutation := graphql.NewObject(graphql.ObjectConfig{Name: "Mutation", Fields: mutationFields})

	// Step 3: create a schema config and a schema
	schemaConfig := graphql.SchemaConfig{Query: query, Mutation: mutation}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	http.HandleFunc("/graphql", func(rw http.ResponseWriter, req *http.Request) {
		query := req.URL.Query().Get("query")

		// Step 4: execute the query based on the query string sent by the client. If the request string is empty an error is returned.
		res := graphql.Do(graphql.Params{
			Schema:        schema,
			RequestString: query,
			Context:       req.Context(),
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
	if err = http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}
}
