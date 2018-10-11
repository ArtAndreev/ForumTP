package main

import (
	"log"
	"net/http"

	"github.com/ArtAndreev/ForumTP/queries"
)

func main() {
	db := queries.InitDB("docker:docker@localhost:5432", "docker")
	defer db.Close()

	log.Println("starting server at:", 5000)
	http.ListenAndServe(":5000", nil)
}
