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

func CreateThread(w http.ResponseWriter, r *http.Request) {
	t := &models.Thread{}
	err := cleanBody(r, t)
	if err != nil {
		if err == ErrWrongJSON {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	t.Forum = mux.Vars(r)["slug"]

	res, err := queries.CreateThread(t)
	if err != nil {
		switch err.(type) {
		case *queries.NullFieldError:
			j, jErr := json.Marshal(models.ErrorMessage{Message: err.Error()})
			if jErr != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, string(j))
		case *queries.UniqueFieldValueAlreadyExistsError:
			j, err := json.Marshal(res)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusConflict)
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

	// if record has been inserted successfully
	j, err := json.Marshal(res)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, string(j))
}

func GetThreads(w http.ResponseWriter, r *http.Request) {
	params := &models.ThreadQueryParams{}
	decoder.IgnoreUnknownKeys(true)
	err := decoder.Decode(params, r.URL.Query())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	res, err := queries.GetAllThreadsInForum(mux.Vars(r)["slug"], params)
	if err != nil {
		switch err.(type) {
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

	j, jErr := json.Marshal(res)
	if jErr != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, string(j))
}
