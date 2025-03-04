package validate

import (
	"encoding/xml"
	"io"
	"net/http"
)

type xmlBinder struct{}

func (xmlBinder) Name() string {
	return "xml"
}

func (xmlBinder) Bind(req *http.Request, obj any) error {
	return decodeXML(req.Body, obj)
}

func decodeXML(r io.Reader, obj any) error {
	decoder := xml.NewDecoder(r)
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return validateAllParams(obj)
}
