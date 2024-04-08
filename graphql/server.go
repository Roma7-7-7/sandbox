package main

import (
	"net/http"
	"os"

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

	_ = todoist.NewClient(todoistToken, http.DefaultClient)

}
