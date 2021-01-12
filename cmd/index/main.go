package main

import (
	"log"

	"github.com/quantonganh/blog/http"
	"github.com/quantonganh/blog/ondisk"
)

func main() {
	posts, err := ondisk.GetAllPosts("posts")
	if err != nil {
		log.Fatal(err)
	}

	s := &http.Server{
		PostService: ondisk.NewPostService(posts),
	}
	_, err = s.PostService.IndexPosts(http.IndexPath)
	if err != nil {
		log.Fatal(err)
	}
}
