package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/rbroggi/crawler/crawler"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	// cmd line flags
	rootURL := flag.String("url", "http://localhost:8080/index.html", "URL to be recursively crawled")
	flag.Parse()

	// Create a new context that can be cancelled with ctrl+c
	ctx := signalContext(context.Background())

	// Parsing input URL
	baseURL, err := url.Parse(*rootURL)
	if err != nil {
		log.Errorf("Error while parsing root URL: [%v]", err)
		os.Exit(1)
	}

	c := crawler.NewCrawler()
	// Crawl input URL and for each page prints url + links
	err = c.Crawl(ctx, baseURL, WritePageURLAndLinksToStdOut)
	if err != nil {
		log.Printf("Error while crawling: [%v]\n", err)
		os.Exit(2)
	}
}

// WritePageURLAndLinksToStdOut takes a page url and it's html content
// and writes to stdout the url of the page along with all the
// links in the page in both the raw form (the one in found in the html)
// and in it's absolute form
func WritePageURLAndLinksToStdOut(u *url.URL, page *html.Node) {
	var b strings.Builder
	_, err := fmt.Fprintf(&b, "url: %s\n", u.String())
	if err != nil {
		log.Errorf("Error while writing into strings.Builder")
		return
	}
	links := crawler.GetPageLinks(page)
	for link := range links {
		absLink, err := crawler.GetLinkAbsoluteUrl(u, link)
		if err != nil {
			log.Errorf("Error while parsing link: [%s]", link)
		}
		_, err = fmt.Fprintf(&b, "link: %s | abs link: %s\n", link, absLink)
		if err != nil {
			log.Errorf("Error while writing into strings.Builder")
			return
		}
	}
	fmt.Println(b.String())
}

// signalContext takes a parentCtx and returns
// a ctx decorated with cancelation through SIGINT (ctrl+c)
// feature
func signalContext(parentCtx context.Context) context.Context {
	// Create a new context, with its cancellation function
	// from the original context
	ctx, cancel := context.WithCancel(parentCtx)

	// create a channel to communicate OS Signals
	c := make(chan os.Signal, 1)
	// configure the os.Interrupt (ctrl+c) to be enqueued to the c channel
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		cancel()
		<-time.After(time.Second)
		log.Fatal("[Error] unclean exit")
	}()
	return ctx
}
