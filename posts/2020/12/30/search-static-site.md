---
title: How to add full text search to my static website?
date: 2020-12-30
description:
categories:
    - Programming
tags:
    - bleve
    - full-text search
    - golang
---
After building the website, I want to add search function.
There are some ways to do it:

- [ElasticSearch](https://www.elastic.co/guide/en/elasticsearch/reference/current/getting-started.html): setup and maintenance cost
- [MongoDB](https://docs.mongodb.com/manual/text-search/): have to insert posts into database
- [Google Custom Search Engine](https://programmablesearchengine.google.com/about/): ads?

So, is there any search engine written in Go?

- [Bleve](https://github.com/blevesearch/bleve)
- [riot](https://github.com/go-ego/riot)

`Bleve` has more star and be maintained more often than `riot`, I would like to give it a try first.

### Backend

Looking at the [documentation](https://blevesearch.com/docs/Getting%20Started/), there are three steps to add search to your website:

- [Open a new index](https://github.com/quantonganh/blog/blob/144fdd13e62faa0aa74f488967c187c0a8e52cc4/markdown/search.go#L38-L55):

```go
func createOrOpenIndex(posts []*blog.Post, indexPath string) (bleve.Index, error) {
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

		if err := deletePostsFromIndex(posts, index); err != nil {
			return nil, err
		}
	}
	return index, nil
}
```

Index will be created when starting, so it must be updated if there is a post has been deleted:

```go
func deletePostsFromIndex(posts []*blog.Post, index bleve.Index) error {
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
```


- [Index posts](https://github.com/quantonganh/blog/blob/144fdd13e62faa0aa74f488967c187c0a8e52cc4/markdown/search.go#L17-L36):

```go
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
```

- [Search](https://github.com/quantonganh/blog/blob/144fdd13e62faa0aa74f488967c187c0a8e52cc4/markdown/search.go#L92-L113):

```go
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
```

[bleve CLI](https://blevesearch.com/docs/bleve/) can be used to query the index:

```shell
$ bleve query posts.bleve "linux"
5 matches, showing 1 through 5, took 625.866µs
    1. 2019/09/19/about (0.318970)
	Content
		…developer. Then my passionate on Linux and open source led me to a different direction, a system administrator.
With a strong background in Linux system administration, I was gradually moving to backe…
    2. 2019/10/30/docker-exited-127 (0.233462)
	Content
		…/span>
<span style="margin-right:0.4em;padding:0 0.4em 0 0.4em;color:#495050">5</span>COPY build/linux/&lt;binary&gt; .
<span style="margin-right:0.4em;padding:0 0.4em 0 0.4em;color:#495050">6</span>
…
```

### Frontend

Add a search box using [forms](https://getbootstrap.com/docs/4.6/components/forms/):

```html
<form class="form-inline my-2 my-lg-0" action="/search">
    <input class="form-control mr-sm-2" type="search" placeholder="Search" aria-label="Search" name="q">
    <button class="btn btn-outline-success my-2 my-sm-0" type="submit">Search</button>
</form>
```

The [action](https://developer.mozilla.org/en-US/docs/Web/HTML/Element/form#attr-action) attribute defines [where the data gets sent](https://github.com/quantonganh/blog/blob/665f775620c8df1922fa1a48c6a6ffb78ffb4759/http/server.go#L106):

```go
	s.router.HandleFunc("/search", s.Error(s.searchHandler))
```

[searchHandler](https://github.com/quantonganh/blog/blob/8d16352ce5f562f28f0665b993a8650ca2992f86/http/search.go#L7-L32) will:

- parse form
- get the form value
- perform search
- render posts

References:

- [Let's build a Full-Text Search engine](https://artem.krylysov.com/blog/2020/07/28/lets-build-a-full-text-search-engine/)
- [Bleve _source example](https://gist.github.com/michaeljs1990/25f9949c10ab64d971422dda44feea0e)