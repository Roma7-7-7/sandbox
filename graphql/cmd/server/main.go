package main

import (
	"net/http"
	"os"

	gqlhandler "github.com/graphql-go/graphql-go-handler"

	"github.com/Roma7-7-7/sandbox/graphql/graph"
	"github.com/Roma7-7-7/sandbox/graphql/pkg/todoist"
)

const defaultPort = "8080"

func main() {
	todoistToken := os.Getenv("TODOIST_TOKEN")
	if todoistToken == "" {
		panic("TODOIST_TOKEN is required")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	schema, err := graph.NewSchema()

	if err != nil {
		panic(err)
	}

	h := gqlhandler.New(&gqlhandler.Config{
		Schema: schema,
		Pretty: true,
	})

	client := todoist.NewClient(todoistToken, http.DefaultClient)
	http.HandleFunc("/graphql", func(rw http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(rw, r.WithContext(graph.WithResolveContext(r.Context(), client)))
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
