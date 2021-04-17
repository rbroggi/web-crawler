package crawler

import (
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
)

// Crawler is used to Crawl a web-site
type Crawler interface {
	// Crawl will recursively crawl a page URL. It will only crawl addresses that
	// share the same domain and will not follow links to external sites
	// the visit parameter is a function that performs some logic based on a page and it's url
	Crawl(ctx context.Context, base *url.URL, visit func(u *url.URL, page *html.Node)) error
}

type crawler struct {
}

func (c *crawler) Crawl(ctx context.Context, base *url.URL, visit func(u *url.URL, page *html.Node)) error {
	if base == nil {
		return errors.New("nil base URL cannot be crawled")
	}

	// used to track end of all spawned go-routines
	var wg sync.WaitGroup
	var rw sync.RWMutex
	visited := make(map[string]struct{})

	recursiveVisit(ctx, &rw, &wg, visited, base, visit)

	// waits all go-routines to finish
	wg.Wait()

	return nil
}

func recursiveVisit(ctx context.Context, rw *sync.RWMutex, wg *sync.WaitGroup, visited map[string]struct{}, u *url.URL, visit func(u *url.URL, page *html.Node)) {
	// collect token for spawning new go-routine
	wg.Add(1)
	go func() {
		defer wg.Done()
		// add u to visited pages
		rw.Lock()
		visited[u.String()] = struct{}{}
		rw.Unlock()

		page, err := getPage(ctx, u.String())
		// if error while getting page simply return
		if err != nil {
			log.Errorf("failed to get page %s", u)
			return
		}

		// apply the visit function
		visit(u, page)

		// retrieve all links in the page
		links := GetPageLinks(page)
		for link := range links {
			absLink, err := GetLinkAbsoluteUrl(u, link)
			if err != nil {
				log.Errorf("failed to get absolute link on page %s with relative link %s", u, link)
				continue
			}
			rw.RLock()
			_, ok := visited[absLink.String()]
			rw.RUnlock()
			// if not visited and same url domain, visit it
			if !ok && isSameDomain(u, absLink) {
				// if context cancelled algo recursion stops
				select {
				case <-ctx.Done():
					return
				default:
					recursiveVisit(ctx, rw, wg, visited, absLink, visit)
				}
			}
		}
	}()
}

// NewCrawler creates a structure that implements the Crawler interface
// the maxConcurrency param determines how many go-routines can concurrently
// crawl the site
func NewCrawler() Crawler {
	return &crawler{}
}

// GetLinkAbsoluteUrl parses a link transforming it in an absolute URI
// if the link is itself an absolute path it parses it regardless of the
// input parent URL, otherwise it parses it relative to the parent URL
func GetLinkAbsoluteUrl(parent *url.URL, link string) (*url.URL, error) {
	r, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	// if link is already an absolute link
	if r.IsAbs() {
		return r, nil
	}

	// if link is relative to server root
	if strings.HasPrefix(link, "/") {
		return parent.Parse(link)
	}

	// if link is relative to parent page uri
	return parent.Parse(path.Join(path.Dir(parent.Path), link))
}

// isSameDomain checks whether p, u urls belong to the same domain
// we check that with the host URL property which is usually the
// result of hostname:port
func isSameDomain(p, u *url.URL) bool {
	if p == nil || u == nil {
		return false
	}

	return p.Host == u.Host
}

// GetPageLinks retrieve all links found in a page
// a set (map[string]struct{}) is used to add semantic meaning
// to the method - no duplicated links are going to be retrieved
func GetPageLinks(node *html.Node) map[string]struct{} {
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

// getPage performs an HTTP GET request using the input url and tries
// to parse the result into an html.Node data structure
func getPage(ctx context.Context, url string) (*html.Node, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error while preparing request - %v", err)
	}
	client := http.DefaultClient
	r, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error while getting page - %v", err)
	}
	b, err := html.Parse(r.Body)
	if err != nil {
		return nil, fmt.Errorf("error while html parsing response - %v", err)
	}
	return b, err
}
