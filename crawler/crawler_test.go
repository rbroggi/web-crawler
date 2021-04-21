package crawler

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
	"net/url"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
)

// getBaseURLStr retrieves from the env variable `CRAWLER_BASE_URL`
// the baseURL for the web-server delivered along with this software
// for integration-tests purposes. The default value is "http://localhost:8080/"
func getBaseURLStr() string {
	baseUrl, ok := os.LookupEnv("CRAWLER_BASE_URL")
	if !ok {
		baseUrl = "http://localhost:8080/"
	}
	return baseUrl
}

// getURL is utility function for Integration tests it parses a url string
// into a *url.URL and if the parsing fails it panics stoping the test
func getURL(urlStr string) *url.URL {
	baseURL, err := url.Parse(urlStr)
	if err != nil {
		panic(fmt.Sprintf("error while parsing url [%s]", urlStr))
	}
	return baseURL
}

// getBaseURLIndex will retrieve the baseURL for the server to be target
// and adds to it the suffix "index.html" which is the root page for the
// web-server that is delivered for testability purpose along with this
// software in the folder `test_data`
func getBaseURLIndex() string {
	return getBaseURLStr() + "index.html"
}

// pageTitle given a reference to a html.Node, scans it until it
// finds the title tag, and returns its value
// reference an credits to : https://flaviocopes.com/golang-web-crawler/
func pageTitle(n *html.Node) string {
	var title string
	if n.Type == html.ElementNode && n.Data == "title" {
		return n.FirstChild.Data
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		title = pageTitle(c)
		if title != "" {
			break
		}
	}
	return title
}

func Test_getPageLinks(t *testing.T) {
	tests := map[string]struct {
		htmlStr string
		want    map[string]struct{}
	}{
		"two_valid_links": {
			htmlStr: `<!doctype html>
	                  <html>
						  <head></head>
						  <body>
							<p>index</p>
							<p><a href="index.html">index</a></p>
							<p><a href="page1.html">p1</a></p>
						  </body>
					  </html>`,
			want: map[string]struct{}{
				"index.html": {},
				"page1.html": {},
			},
		},
		"one_valid_link": {
			htmlStr: `<!doctype html>
	                  <html>
						  <head></head>
						  <body>
							<p>index</p>
							<p><a href="index.html">index</a></p>
						  </body>
					  </html>`,
			want: map[string]struct{}{
				"index.html": {},
			},
		},
		"3_valid_links": {
			htmlStr: `<!doctype html>
	                  <html>
						  <head></head>
						  <body>
							<p>index</p>
                            <div><a href="index2.html"/></div>
							<p><a href="index.html">index</a></p>
							<p><a href="index3.html">index</a></p>
						  </body>
					  </html>`,
			want: map[string]struct{}{
				"index.html":  {},
				"index2.html": {},
				"index3.html": {},
			},
		},
		"no_duplicate_outputs": {
			htmlStr: `<!doctype html>
	                  <html>
						  <head></head>
						  <body>
							<p>index</p>
                            <div><a href="index.html"/></div>
							<p><a href="index.html">index</a></p>
							<p><a href="index.html">index</a></p>
						  </body>
					  </html>`,
			want: map[string]struct{}{
				"index.html": {},
			},
		},
		"no_link_empty_map": {
			htmlStr: `<!doctype html>
	                  <html>
						  <head></head>
						  <body>
							<p>index</p>
                            <div></div>
							<p>test</p>
							<p>test</p>
						  </body>
					  </html>`,
			want: map[string]struct{}{},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			node, err := html.Parse(strings.NewReader(tt.htmlStr))
			// well formatted html
			assert.Nil(t, err)
			// test for map equality
			links := GetPageLinks(node)
			assert.True(t, reflect.DeepEqual(links, tt.want))
		})
	}
}

func Test_getLinkAbsoluteUrl(t *testing.T) {
	tests := map[string]struct {
		par     string
		link    string
		wantErr bool
		want    string
	}{
		"absolute_path": {
			par:     "https://my-web-site.com",
			link:    "https://my-web-site.com/i/business",
			wantErr: false,
			want:    "https://my-web-site.com/i/business",
		},
		"absolute_path_error": {
			par:     "https://my-web-site.com",
			link:    "://my-web-site.com/i/business",
			wantErr: true,
			want:    "does_not_matter",
		},
		"relative_path": {
			par:     "https://my-web-site.com",
			link:    "i/business",
			wantErr: false,
			want:    "https://my-web-site.com/i/business",
		},
		"relative_path_with_starting_slash": {
			par:     "https://my-web-site.com",
			link:    "/i/business",
			wantErr: false,
			want:    "https://my-web-site.com/i/business",
		},
		"relative_path_with_starting_from_root_of_server": {
			par:     "https://my-web-site.com/test",
			link:    "/i/business",
			wantErr: false,
			want:    "https://my-web-site.com/i/business",
		},
		"relative_path_with_starting_from_served_page": {
			par:     "https://my-web-site.com/test",
			link:    "i/business",
			wantErr: false,
			want:    "https://my-web-site.com/i/business",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			par, err := url.Parse(tt.par)
			assert.Nil(t, err)

			want, err := url.Parse(tt.want)
			assert.Nil(t, err)

			got, err := GetLinkAbsoluteUrl(par, tt.link)

			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Equal(t, want, got)
			}
		})
	}
}

func Test_isSameDomain(t *testing.T) {
	tests := map[string]struct {
		par   string
		child string
		want  bool
	}{
		"same_scheme_same_domain_different_path_should_be_considered_same_domain": {
			par:   "https://my-web-site.com/i/business",
			child: "https://my-web-site.com/i/business/test",
			want:  true,
		},
		"different_scheme_same_domain_different_path_should_be_considered_same_domain": {
			par:   "https://my-web-site.com/i/business",
			child: "http://my-web-site.com/i/business/test",
			want:  true,
		},
		"different_host_will_lead_to_different_domain": {
			par:   "https://my-web-serverb.com/i/business",
			child: "http://my-web-servera.com/i/business",
			want:  false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			par, err := url.Parse(tt.par)
			assert.Nil(t, err)

			child, err := url.Parse(tt.child)
			assert.Nil(t, err)

			assert.Equal(t, tt.want, isSameDomain(par, child))
		})
	}
}

// Test_crawler_Crawl_Integration depends on the startup of the attached `test_data/Dockerfile` container
// this test will perform an actual integration test of the crawl transversal method. The goal is to
// transverse the web-server and collect for each of the page, it's title. If the traversing is correct
// all the pages (except the orphans) will be in the collected map
func Test_crawler_Crawl_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	tests := map[string]struct {
		c Crawler
	}{
		"recursive_crawler": {
			c: NewCrawler(),
		},
		"ciclic_crawler": {
			c: NewCyclicCrawler(0),
		},
		"serial_ciclic_crawler": {
			c: NewCyclicCrawler(1),
		},
		"serial_ciclic_crawler_concurrent": {
			c: NewCyclicCrawler(10),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {

			// build a visit function that collects in a map all the
			// page titles
			m := make(map[string]struct{})
			mut := sync.Mutex{}

			// visit is a function that given a page collects
			// it's title and adds it to a set of titles
			visit := func(u *url.URL, page *html.Node) {
				title := pageTitle(page)
				// add the title to the map
				mut.Lock()
				m[title] = struct{}{}
				mut.Unlock()
			}
			err := tt.c.Crawl(context.Background(), getURL(getBaseURLIndex()), visit)
			assert.Nil(t, err)

			// want is the titles of all the pages on the web-server
			want := map[string]struct{}{
				"index":  {},
				"page1":  {},
				"page2":  {},
				"page3":  {},
				"page11": {},
			}

			// check equality on collected titles and expected titles
			assert.True(t, reflect.DeepEqual(m, want))
		})
	}
}

// Test_crawler_Crawl_Cancelation_Integration will test that the crawler is cancellable
// this test leverages on a visit function that collects the very first title and issues
// a context cancellation which will stop the whole algorithm therefore only collecting
// one single title out of the web-server
func Test_crawler_Crawl_Cancelation_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	// build a crawler
	c := &crawler{}
	// build a visit function that collects in a map all the
	// page titles
	m := make(map[string]struct{})
	mut := sync.Mutex{}

	// creating ctx with cancellation feature
	ctx, cancel := context.WithCancel(context.Background())

	// visit is a function that given a page collects
	// it's title and adds it to a set of titles
	visit := func(u *url.URL, page *html.Node) {
		title := pageTitle(page)
		// add the title to the map
		mut.Lock()
		m[title] = struct{}{}
		mut.Unlock()

		// issues the cancellation after first title was collected
		// rest of the pages should not get scraped
		cancel()
	}

	err := c.Crawl(ctx, getURL(getBaseURLIndex()), visit)
	assert.Nil(t, err)

	// want contains only the title of the root page because the cancellation was
	// issued before the scrapping could finish the first node
	want := map[string]struct{}{
		"index": {},
	}

	// check equality on collected titles and expected titles
	assert.True(t, reflect.DeepEqual(m, want))
}
