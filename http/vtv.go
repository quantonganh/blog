package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

const wordsPerLine = 5

func (s *Server) vtvHandler(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case http.MethodGet:
		return s.Renderer.RenderVTV(w, "", 0, [][]string{})
	case http.MethodPost:
		letters := r.FormValue("letters")
		vtvHost := os.Getenv("VTV_HOST")
		resp, err := http.Post(vtvHost, "application/json", bytes.NewBuffer([]byte(letters)))
		if err != nil {
			return errors.Wrap(err, "failed to send post request")
		}

		var words []string
		if err := json.NewDecoder(resp.Body).Decode(&words); err != nil {
			return errors.Wrap(err, "failed to decode the response body")
		}

		total := len(words)
		rows := make([][]string, 0)
		for i := 0; i < total; i += wordsPerLine {
			end := i + wordsPerLine
			if end > total {
				end = total
			}
			rows = append(rows, words[i:end])
		}
		return s.Renderer.RenderVTV(w, letters, total, rows)
	default:
		return errors.New("invalid method")
	}
}
