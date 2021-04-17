package main

import (
	"context"
	"flag"
	"github.com/rbroggi/crawler/crawler"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// cmd line flags
	rootURL := flag.String("url", "http://localhost:8080/index.html", "URL to be recursively crawled")
	maxPoolSize := flag.CommandLine.Uint("pool", 100000, "The max number of go-routines that can be created")
	flag.Parse()

	// Create a new context
	ctx := signalContext(context.Background())

	c := crawler.NewCrawler(uint32(*maxPoolSize))
	baseURL, err := url.Parse(*rootURL)
	if err != nil {
		log.Printf("Error while parsing URL: [%v]", err)
		os.Exit(1)
	}
	// Crawl url
	err = c.Crawl(ctx, baseURL)
	if err != nil {
		log.Printf("Error while crawling: [%v]\n", err)
		os.Exit(2)
	}
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
