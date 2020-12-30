package main

import (
	"net/url"
	"os"
)

type remark struct {
	URL     *url.URL
	PageURL *url.URL
}

func getRemarkURL() (*remark, error) {
	remarkURL, err := url.Parse(os.Getenv("REMARK_URL"))
	if err != nil {
		return nil, err
	}
	pageURL, err := url.Parse(os.Getenv("PAGE_URL"))
	if err != nil {
		return nil, err
	}

	return &remark{
		URL:     remarkURL,
		PageURL: pageURL,
	}, nil
}
