package markdown

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/document"
	"github.com/blevesearch/bleve/index/scorch"
	"github.com/pkg/errors"

	"github.com/quantonganh/blog"
)

type searchService struct {
	index bleve.Index
}

// NewSearchService returns new search service
func NewSearchService(indexPath string, posts []*blog.Post) (blog.SearchService, error) {
	var index bleve.Index
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		index, err = bleve.NewUsing(indexPath, bleve.NewIndexMapping(), scorch.Name, scorch.Name, nil)
		if err != nil {
			return nil, errors.Errorf("failed to create index at %s: %v", indexPath, err)
		}
	} else if err == nil {
		index, err = bleve.OpenUsing(indexPath, nil)
		if err != nil {
			return nil, errors.Errorf("failed to open index at %s: %v", indexPath, err)
		}

		if err := deletePostsFromIndex(index, posts); err != nil {
			return nil, err
		}
	}

	batch := index.NewBatch()
	for i, post := range posts {
		post.ID = i

		if err := indexPost(post, batch); err != nil {
			return nil, err
		}
	}

	if err := index.Batch(batch); err != nil {
		return nil, errors.Wrapf(err, "failed to index batch")
	}

	return &searchService{
		index: index,
	}, nil
}

func indexPost(post *blog.Post, batch *bleve.Batch) error {
	doc := document.Document{
		ID: post.URI,
	}
	if err := bleve.NewIndexMapping().MapDocument(&doc, post); err != nil {
		return errors.Errorf("failed to map document: %v", err)
	}

	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	if err := enc.Encode(post); err != nil {
		return errors.Errorf("failed to encode post: %v", err)
	}

	field := document.NewTextFieldWithIndexingOptions("_source", nil, b.Bytes(), document.StoreField)
	if err := batch.IndexAdvanced(doc.AddField(field)); err != nil {
		return errors.Errorf("failed to add index to the batch: %v", err)
	}
	return nil
}

func deletePostsFromIndex(index bleve.Index, posts []*blog.Post) error {
	count, err := index.DocCount()
	if err != nil {
		return errors.Errorf("failed to get number of documents in the index: %v", err)
	}

	query := bleve.NewMatchAllQuery()
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Size = int(count)
	searchResults, err := index.Search(searchRequest)
	if err != nil {
		return errors.Errorf("failed to find all documents in the index: %v", err)
	}
	for i := 0; i < len(searchResults.Hits); i++ {
		uri := searchResults.Hits[i].ID
		if !isContains(posts, uri) {
			if err := index.Delete(uri); err != nil {
				return err
			}
		}
	}
	return nil
}

func isContains(posts []*blog.Post, uri string) bool {
	for _, post := range posts {
		if post.URI == uri {
			return true
		}
	}

	return false
}

func (ss *searchService) Search(value string) ([]*blog.Post, error) {
	query := bleve.NewMatchQuery(value)
	request := bleve.NewSearchRequest(query)
	request.Fields = []string{"_source"}

	size, err := ss.index.DocCount()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to count number of documents in the index")
	}
	request.Size = int(size)

	searchResults, err := ss.index.Search(request)
	if err != nil {
		return nil, errors.Errorf("failed to execute a search request: %v", err)
	}

	var searchPosts []*blog.Post
	for _, result := range searchResults.Hits {
		var post *blog.Post
		b := bytes.NewBuffer([]byte(fmt.Sprintf("%v", result.Fields["_source"])))
		dec := gob.NewDecoder(b)
		if err = dec.Decode(&post); err != nil {
			return nil, errors.Errorf("failed to decode post: %v", err)
		}
		searchPosts = append(searchPosts, post)
	}

	return searchPosts, nil
}

func (ss *searchService) CloseIndex() error {
	return ss.index.Close()
}
