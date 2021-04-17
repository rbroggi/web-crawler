package main

import (
	"golang.org/x/net/html"
	"net/url"
	"strings"
)

func ExampleWritePageURLAndLinksToStdOut() {
	u, err := url.Parse("https://my-web-site.com/root/parent")
	if err != nil {
		panic("error parsing url")
	}

	htmlStr := `<!doctype html>
		        <html>
			    	<head></head>
			    	<body>
						<p>index</p>
						<p><a href="/index.html">index</a></p>
						<p><a href="index.html">index</a></p>
						<p><a href="https://another-web-site.com/root">p1</a></p>
			  		</body>
			    </html>`

	node, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		panic("error parsing html")
	}

	WritePageURLAndLinksToStdOut(u, node)

	// Unordered output:
	// url: https://my-web-site.com/root/parent
	// link: /index.html | abs link: https://my-web-site.com/index.html
	// link: index.html | abs link: https://my-web-site.com/root/index.html
	// link: https://another-web-site.com/root | abs link: https://another-web-site.com/root
}
