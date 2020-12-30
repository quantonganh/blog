package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/document"
	"github.com/blevesearch/bleve/index/scorch"
	"github.com/blevesearch/bleve/mapping"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

func (b *Blog) indexPosts(path string) (bleve.Index, error) {
	indexMapping := bleve.NewIndexMapping()
	index, err := bleve.NewUsing(path, indexMapping, scorch.Name, scorch.Name, nil)
	if err != nil {
		return nil, errors.Errorf("failed to create index at %s: %v", path, err)
	}

	g, ctx := errgroup.WithContext(context.Background())
	for _, post := range b.posts {
		post := post // https://golang.org/doc/faq#closures_and_goroutines
		g.Go(func() error {
			return indexPost(ctx, indexMapping, index, post)
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return index, nil
}

func indexPost(ctx context.Context, mapping *mapping.IndexMappingImpl, index bleve.Index, post *Post) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		doc := document.Document{
			ID: post.URI,
		}
		if err := mapping.MapDocument(&doc, post); err != nil {
			return errors.Errorf("failed to map document: %v", err)
		}

		var b bytes.Buffer
		enc := gob.NewEncoder(&b)
		if err := enc.Encode(post); err != nil {
			return errors.Errorf("failed to encode post: %v", err)
		}

		field := document.NewTextFieldWithIndexingOptions("_source", nil, b.Bytes(), document.StoreField)
		batch := index.NewBatch()
		if err := batch.IndexAdvanced(doc.AddField(field)); err != nil {
			return errors.Errorf("failed to add index to the batch: %v", err)
		}
		if err := index.Batch(batch); err != nil {
			return errors.Errorf("failed to index batch: %v", err)
		}

		return nil
	}
}

func (b *Blog) search(index bleve.Index, value string) ([]*Post, error) {
	query := bleve.NewMatchQuery(value)
	request := bleve.NewSearchRequest(query)
	request.Fields = []string{"_source"}
	searchResults, err := index.Search(request)
	if err != nil {
		return nil, errors.Errorf("failed to execute a search request: %v", err)
	}

	var searchPosts []*Post
	for _, result := range searchResults.Hits {
		var post *Post
		b := bytes.NewBuffer([]byte(fmt.Sprintf("%v", result.Fields["_source"])))
		dec := gob.NewDecoder(b)
		if err = dec.Decode(&post); err != nil {
			return nil, errors.Errorf("failed to decode post: %v", err)
		}
		searchPosts = append(searchPosts, post)
	}

	return searchPosts, nil
}
