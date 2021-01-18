package http

import (
	"encoding/xml"
	"fmt"
	"net/http"

	"github.com/quantonganh/blog"
)

const xmlns = "http://www.sitemaps.org/schemas/sitemap/0.9"

func (s *Server) sitemapHandler(w http.ResponseWriter, r *http.Request) error {
	sitemap := blog.Sitemap{
		XMLNS: xmlns,
		URLs: []blog.URL{
			{
				Loc: s.URL(),
			},
		},
	}

	for _, p := range s.PostService.GetAllPosts() {
		sitemap.URLs = append(sitemap.URLs, blog.URL{
			Loc:     fmt.Sprintf("%s/%s", s.URL(), p.URI),
			LastMod: blog.ToISODate(p.Date),
		})
	}

	output, err := xml.MarshalIndent(sitemap, "  ", "    ")
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(xml.Header + string(output)))
	if err != nil {
		return err
	}

	return nil
}
