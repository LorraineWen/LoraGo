package render

import "net/http"

type Render interface {
	Render(w http.ResponseWriter) error
}

func writeContentType(w http.ResponseWriter, responseType string) {
	w.Header().Set("Content-Type", responseType)
}
