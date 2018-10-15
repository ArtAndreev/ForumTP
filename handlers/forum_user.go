package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/ArtAndreev/ForumTP/models"
	"github.com/ArtAndreev/ForumTP/queries"
)

func CreateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	u := &models.BaseForumUser{}
	err := cleanBody(r, u)
	if err != nil {
		if err == ErrWrongJSON {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	u.Nickname = vars["nickname"]

	err = queries.CreateUser(u)
	if err != nil {
		// better handling...
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	j, err := json.Marshal(u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, string(j))
}
