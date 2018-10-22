package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/ArtAndreev/ForumTP/models"
	"github.com/ArtAndreev/ForumTP/queries"
)

func VoteForPost(w http.ResponseWriter, r *http.Request) {
	v := &models.Vote{}
	err := cleanBody(r, v)
	if err != nil {
		if err == ErrWrongJSON {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	path := mux.Vars(r)["slug_or_id"]

	res, err := queries.VoteForPost(v, path)
	if err != nil {
		switch err.(type) {
		case *queries.NullFieldError, *queries.ValidationError:
			j, jErr := json.Marshal(models.ErrorMessage{Message: err.Error()})
			if jErr != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, string(j))
		case *queries.RecordNotFoundError:
			j, jErr := json.Marshal(models.ErrorMessage{Message: err.Error()})
			if jErr != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, string(j))
		default:
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// if records have been inserted successfully
	j, err := json.Marshal(res)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, string(j))
}
