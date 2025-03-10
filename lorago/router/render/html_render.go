package render

import (
	"html/template"
	"net/http"
)

// 用于html模板加载到内存中时，对加载的模板进行存储
type HtmlTemplateRender struct {
	Template *template.Template
}

var htmlContentType string = "text/html; charset=utf-8"

type HtmlRender struct {
	Data       any
	Template   *template.Template
	Name       string
	IsTemplate bool
}

func (h *HtmlRender) Render(w http.ResponseWriter, status int) error {
	writeContentType(w, htmlContentType)
	w.WriteHeader(status)
	if !h.IsTemplate {
		_, err := w.Write([]byte(h.Data.(string)))
		return err
	}
	err := h.Template.ExecuteTemplate(w, h.Name, h.Data)
	return err
}
