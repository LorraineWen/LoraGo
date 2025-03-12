package lora_render

import (
	"encoding/xml"
	"net/http"
)

type XmlRender struct {
	Data any
}

var xmlContentType = "application/xml; charset=utf-8"

func (x *XmlRender) Render(w http.ResponseWriter, status int) error {
	writeContentType(w, xmlContentType)
	w.WriteHeader(status)
	return xml.NewEncoder(w).Encode(x.Data)
}
