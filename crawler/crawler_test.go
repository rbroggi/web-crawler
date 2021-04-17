package crawler

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
	"reflect"
	"strings"
	"testing"
)

func Test_getPageLinks(t *testing.T) {
	tests := map[string]struct {
		htmlStr string
		want map[string]struct{}
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
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			node, err := html.Parse(strings.NewReader(tt.htmlStr))
			// well formatted html
			assert.Nil(t, err)
			// test for map equality
			links := getPageLinks(node)
			assert.True(t, reflect.DeepEqual(links, tt.want))
		})
	}
}

