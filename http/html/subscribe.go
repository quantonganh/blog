package html

import (
	"html/template"
	"net/http"

	"github.com/flosch/pongo2"
	"github.com/pkg/errors"
)

type subscribe struct {
	tmpl *template.Template
}

func NewSubscribe(tmpl *template.Template) *subscribe {
	return &subscribe{
		tmpl: tmpl,
	}
}

func (s *subscribe) Render(w http.ResponseWriter, message string) error {
	data := pongo2.Context{"message": message}

	if err := s.tmpl.ExecuteTemplate(w, "subscribe", data); err != nil {
		return errors.Errorf("failed to execute template: %v", err)
	}

	return nil
}
