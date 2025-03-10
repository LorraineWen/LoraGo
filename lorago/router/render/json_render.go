package render

import (
	"encoding/json"
	"net/http"
)

type JsonRender struct {
	Data any
}

var jsonContentType string = "application/json; charset=utf-8"

func (j *JsonRender) Render(w http.ResponseWriter, status int) error {
	writeContentType(w, jsonContentType)
	w.WriteHeader(status)
	dataJson, err := json.Marshal(j.Data)
	if err != nil {
		return err
	}
	_, err = w.Write(dataJson)
	return err
}
