package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func cleanBody(r *http.Request, s interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, s)
	if err != nil {
		return ErrWrongJSON
	}

	return nil
}
