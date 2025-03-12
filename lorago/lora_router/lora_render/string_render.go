package lora_render

import (
	"fmt"
	"github.com/LorraineWen/lorago/lora_util"
	"net/http"
)

type StringRender struct {
	Format string
	Data   []any
}

var plainContentType string = "text/plain; charset=utf-8"

func (s *StringRender) Render(w http.ResponseWriter, status int) error {
	writeContentType(w, plainContentType)
	w.WriteHeader(status)
	if len(s.Data) > 0 {
		_, err := fmt.Fprintf(w, s.Format, s.Data...)
		if err != nil {
			return err
		}
		return nil
	}
	_, err := w.Write(lora_util.StringToByte(s.Format))
	return err
}
