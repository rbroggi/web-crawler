package crawler

import (
	"context"
	"errors"
	"golang.org/x/net/html"
	"net/url"
)

// Crawler is used to Crawl a web-site
type Crawler interface {
	// Crawl will recursively crawl a page URL. It will only crawl addresses that
	// share the same domain and will not follow links to external sites
	Crawl(ctx context.Context, url *url.URL) error
}

type crawler struct {
	token <-chan struct{}
}

func (c *crawler) Crawl(ctx context.Context, url *url.URL) error {
	return errors.New("Not implemented")
}

// NewCrawler creates a structure that implements the Crawler interface
// the maxConcurrency param determines how many go-routines can concurrently
// crawl the site
func NewCrawler(maxConcurrency uint32) Crawler {
	return &crawler {
		token: make(chan struct{}, maxConcurrency),
	}
}


// getPageLinks retrieve all links found in a page
// a set (map[string]struct{}) is used to add semantic meaning
// to the method - no duplicated links are going to be retrieved
func getPageLinks(node *html.Node) map[string]struct{} {
	m := make(map[string]struct{})
	// nothing to retrieve for nil node
	if node == nil {
		return m
	}
	return getPageLinksRecursive(node, m)
}

// getPageLinksRecursive retrieve all links found in a page
// a set (map[string]struct{}) is used to add semantic meening
// to the method - no duplicated links are going to be retrieved
func getPageLinksRecursive(node *html.Node, links map[string]struct{}) map[string]struct{} {
	// check if we are in a <a></a> html element
	if node.Type == html.ElementNode && node.Data == "a" {
		// cycle through element attributes
		for _, a := range node.Attr {
			// for href element
			if a.Key == "href" {
				// if link not yet considered put it on the set
				if _, ok := links[a.Val]; !ok {
					links[a.Val] = struct{}{}
				}
			}
		}
	}
	for n := node.FirstChild; n != nil; n = n.NextSibling {
		links = getPageLinksRecursive(n, links)
	}
	return links
}
