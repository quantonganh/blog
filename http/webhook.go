package http

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/quantonganh/blog"
	"github.com/quantonganh/blog/markdown"
)

type webhookPayload struct {
	Commits []struct {
		ID        string   `json:"id"`
		Message   string   `json:"message"`
		Timestamp string   `json:"timestamp"`
		Added     []string `json:"added"`
		Removed   []string `json:"removed"`
		Modified  []string `json:"modified"`
	} `json:"commits"`
}

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

		var payload webhookPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			return err
		}

		addedPosts, removedFiles, modifiedPosts, err := getChangedPosts(config, payload)
		if err != nil {
			return err
		}

		cmd := exec.Command("sh", "-c", fmt.Sprintf("git -C %s fetch origin && git -C %s reset --hard origin/main", config.Posts.Dir, config.Posts.Dir))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}

		if err := s.reload(config, addedPosts, removedFiles, modifiedPosts); err != nil {
			return err
		}

		return nil
	}
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

func getChangedPosts(config *blog.Config, payload webhookPayload) ([]*blog.Post, []string, []*blog.Post, error) {
	var (
		addedFiles    []string
		removedFiles  []string
		modifiedFiles []string
	)
	for _, commit := range payload.Commits {
		addedFiles = append(addedFiles, commit.Added...)
		removedFiles = append(removedFiles, commit.Removed...)
		modifiedFiles = append(modifiedFiles, commit.Modified...)
	}

	var (
		addedPosts    []*blog.Post
		modifiedPosts []*blog.Post
	)
	for _, file := range addedFiles {
		name := filepath.Join(config.Posts.Dir, file)
		f, err := os.Open(name)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error opening new file %s: %w", name, err)
		}

		newPost, err := markdown.Parse(context.Background(), config.Posts.Dir, f)
		if err != nil {
			return nil, nil, nil, err
		}

		addedPosts = append(addedPosts, newPost)
	}

	for _, file := range modifiedFiles {
		name := filepath.Join(config.Posts.Dir, file)
		f, err := os.Open(name)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error opening modified file %s: %w", name, err)
		}

		modifiedPost, err := markdown.Parse(context.Background(), config.Posts.Dir, f)
		if err != nil {
			return nil, nil, nil, err
		}

		modifiedPosts = append(modifiedPosts, modifiedPost)
	}

	return addedPosts, removedFiles, modifiedPosts, nil
}

func (s *Server) reload(config *blog.Config, addedPosts []*blog.Post, removedFiles []string, modifiedPosts []*blog.Post) error {
	if s.PostService != nil {
		posts := s.PostService.GetAllPosts()
		updatedPosts, err := updatePosts(config, posts, addedPosts, removedFiles, modifiedPosts)
		if err != nil {
			return err
		}
		s.PostService = markdown.NewPostService(updatedPosts)
	}

	if s.SearchService != nil {
		index := s.SearchService.GetIndex()
		batch := index.NewBatch()

		for _, name := range removedFiles {
			if err := index.Delete(name); err != nil {
				return err
			}
		}

		for _, post := range append(addedPosts, modifiedPosts...) {
			if err := s.SearchService.Index(post, batch); err != nil {
				return err
			}
		}

		if err := index.Batch(batch); err != nil {
			return err
		}
	}

	log.Println("Content reloaded successfully.")
	return nil
}

func updatePosts(config *blog.Config, posts []*blog.Post, addedPosts []*blog.Post, removedFiles []string, modifiedPosts []*blog.Post) ([]*blog.Post, error) {
	posts = append(posts, addedPosts...)

	for _, name := range removedFiles {
		for i, post := range posts {
			if post.URI == name {
				posts = append(posts[:i], posts[i+1:]...)
			}
		}
	}

	for _, p := range modifiedPosts {
		for i, post := range posts {
			if post.URI == p.URI {
				posts[i] = p
			}
		}
	}

	return posts, nil
}
