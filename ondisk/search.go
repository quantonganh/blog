package ondisk

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

func indexPosts(posts []*blog.Post, indexPath string) (bleve.Index, error) {
	var index bleve.Index
	mapping := bleve.NewIndexMapping()
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		index, err = bleve.NewUsing(indexPath, mapping, scorch.Name, scorch.Name, nil)
		if err != nil {
			return nil, errors.Errorf("failed to create index at %s: %v", indexPath, err)
		}
	} else if err == nil {
		index, err = bleve.OpenUsing(indexPath, nil)
		if err != nil {
			return nil, errors.Errorf("failed to open index at %s: %v", indexPath, err)
		}
	}

	batch := index.NewBatch()
	for _, post := range posts {
		doc := document.Document{
			ID: post.URI,
		}
		if err := mapping.MapDocument(&doc, post); err != nil {
			return nil, errors.Errorf("failed to map document: %v", err)
		}

		var b bytes.Buffer
		enc := gob.NewEncoder(&b)
		if err := enc.Encode(post); err != nil {
			return nil, errors.Errorf("failed to encode post: %v", err)
		}

		field := document.NewTextFieldWithIndexingOptions("_source", nil, b.Bytes(), document.StoreField)
		if err := batch.IndexAdvanced(doc.AddField(field)); err != nil {
			return nil, errors.Errorf("failed to add index to the batch: %v", err)
		}
	}

	if err := index.Batch(batch); err != nil {
		return nil, errors.Errorf("failed to index batch: %v", err)
	}

	return index, nil
}

func (ps *postService) Search(value string) ([]*blog.Post, error) {
	query := bleve.NewMatchQuery(value)
	request := bleve.NewSearchRequest(query)
	request.Fields = []string{"_source"}
	searchResults, err := ps.index.Search(request)
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
