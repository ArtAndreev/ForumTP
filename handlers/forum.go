package handlers

import (
	"fmt"
	"net/http"

	"github.com/ArtAndreev/ForumTP/models"
	"github.com/ArtAndreev/ForumTP/queries"
)

func CreateForum(w http.ResponseWriter, r *http.Request) {
	f := &models.BaseForum{}
	err := cleanBody(r, f)
	if err != nil {
		if err == ErrWrongJSON {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	queries.CreateForum(f)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, f)
}
