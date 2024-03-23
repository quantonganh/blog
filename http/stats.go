package http

import (
	"net/http"

	"github.com/quantonganh/blog/ui/html"
)

func (s *Server) statsHandler(w http.ResponseWriter, r *http.Request) error {
	top10VisitedPages, err := s.StatService.Top10VisitedPages()
	if err != nil {
		return err
	}

	top10Countries, err := s.StatService.Top10Countries()
	if err != nil {
		return err
	}

	top10Referers, err := s.StatService.Top10Referers(s.Domain)
	if err != nil {
		return err
	}

	top10Browsers, err := s.StatService.Top10Browsers()
	if err != nil {
		return err
	}

	top10OperatingSystems, err := s.StatService.Top10OperatingSystems()
	if err != nil {
		return err
	}

	tmpl := html.Parse(nil, "stats.html")
	data := map[string]interface{}{
		"top10VisitedPages":     top10VisitedPages,
		"top10Countries":        top10Countries,
		"top10Referers":         top10Referers,
		"top10Browsers":         top10Browsers,
		"top10OperatingSystems": top10OperatingSystems,
	}
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		return err
	}

	return nil
}
