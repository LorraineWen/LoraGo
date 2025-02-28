package render

import (
	"encoding/xml"
	"net/http"
)

type XmlRender struct {
	Data any
}

var xmlContentType = "application/xml; charset=utf-8"

func (x *XmlRender) Render(w http.ResponseWriter) error {
	writeContentType(w, xmlContentType)
	return xml.NewEncoder(w).Encode(x.Data)
}
