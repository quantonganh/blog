package http

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/quantonganh/blog"
	"github.com/quantonganh/blog/markdown"
)

func (s *Server) webhookHandler(config *blog.Config) appHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}

		signature := r.Header.Get("X-Hub-Signature-256")
		if err := verifySignature(signature, body, config.Webhook.Secret); err != nil {
			return &Error{
				Message: err.Error(),
				Status:  http.StatusUnauthorized,
			}
		}

		cmd := exec.Command("git", "-C", config.Posts.Dir, "pull")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}

		posts, err := markdown.GetAllPosts(config.Posts.Dir)
		if err != nil {
			return err
		}
		s.reload(config, posts)

		return nil
	}
}

func (s *Server) reload(config *blog.Config, posts []*blog.Post) {
	if s.PostService != nil {
		s.PostService = markdown.NewPostService(posts)
	}

	if s.SearchService != nil {
		indexPath := path.Join(path.Dir(config.Posts.Dir), path.Base(config.Posts.Dir)+".bleve")
		searchService, err := markdown.NewSearchService(indexPath, posts)
		if err != nil {
			log.Printf("Failed to reload search service: %v\n", err)
			return
		}
		s.SearchService = searchService
	}

	log.Println("Content reloaded successfully.")
}

func verifySignature(signature string, payload []byte, secret string) error {
	parts := strings.SplitN(signature, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("X-Hub-Signature-256: expected 2 parts, got %d", len(parts))
	}

	algo, hash := parts[0], parts[1]
	if algo != "sha256" {
		return fmt.Errorf("X-Hub-Signature-256 algorithm: expected sha256, got %s", algo)
	}

	hmacHash := hmac.New(sha256.New, []byte(secret))
	_, err := hmacHash.Write(payload)
	if err != nil {
		return err
	}
	calculateHash := hex.EncodeToString(hmacHash.Sum(nil))

	if !hmac.Equal([]byte(calculateHash), []byte(hash)) {
		return errors.New("Request signature didn't match")
	}

	return nil
}
