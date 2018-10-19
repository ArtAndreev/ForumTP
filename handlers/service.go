package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ArtAndreev/ForumTP/queries"
)

func ClearDatabase(w http.ResponseWriter, r *http.Request) {
	err := queries.ClearDatabase()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func GetDatabaseStatus(w http.ResponseWriter, r *http.Request) {
	res, err := queries.GetDatabaseStatus()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	j, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, string(j))
}
