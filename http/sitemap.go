package http

import (
	"encoding/xml"
	"fmt"
	"net/http"

	"github.com/quantonganh/blog"
	"github.com/quantonganh/blog/http/mw"
)

const xmlns = "http://www.sitemaps.org/schemas/sitemap/0.9"

func (s *Server) SitemapHandler(posts []*blog.Post) mw.ErrHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		scheme := "http"
		if xForwardedProto := r.Header.Get("X-Forwarded-Proto"); xForwardedProto != "" {
			scheme = xForwardedProto
		}

		sitemap := blog.Sitemap{
			XMLNS: xmlns,
			URLs: []blog.URL{
				{
					Loc: fmt.Sprintf("%s://%s", scheme, r.Host),
				},
			},
		}

		for _, p := range posts {
			sitemap.URLs = append(sitemap.URLs, blog.URL{
				Loc:     fmt.Sprintf("%s://%s/%s", scheme, r.Host, p.URI),
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
}
