# Web crawler

This program is a command line utility to crawl recursively a url. It features a recursive and concurrent crawling. 

## Design

To enhance testability the program main functionalities were implemented in a dedicated package (`crawler`). 
Along with this package, a very small `main.go` making use of the package is delivered to implement a command-line program. 
The `main.go` program also takes care of creating a cancellable context which is used to signaling cancellation in the 
crawling algorithm based on `ctrl+c` or __SIGINIT__ system call.
The **Crawler** will follow a best effort approach: it will attempt to scrape all the links that are eligible. 
If some HTTP GET fails along the way, the crawler will continue attempting to crawl the rest of the items and will not stop. 
The current program prints to __stdout__ the url and the links found for the scrapped page and all the errors are
logged to __stderr__. Therefore, if you desire to read only the output of the program you can redirect the __stderr__.
In terms of concurrency, the program will spawn new go-routines for each new link to be visited. The required conditions 
for a link to be scraped are:

1. it should not have already been scrapped
2. it should be hosted in the same domain as the root (first) url provided in as the program command line input

Something worth noticing as well is that a link can be provided in 3 different forms:

1. Link with an absolute URI path including the schema and the domain (e.g. `https://my-web-site.com/i/business/`)
1. Link with a path which is relative to the server root folder - starting with '/' (e.g. `/i/business/`)
1. Link with a path which is relative to the current served page (e.g. `i/business/`)

In the `Crawl` method I have decided to use the __second-oder function__ pattern to make the `Crawl` method more 
extensible and easier to test/benchmark. Another advantage of this approach was the ability to reuse the methods
`crawler.GetPageLinks` and `crawler.GetLinkAbsoluteUrl` which are used in the traversal algorithm and in the 
visiting algorithm.

An arbitrary structure was chosen for printing the scraping to __stdout__. You can check that format in the 
`ExampleWritePageURLAndLinksToStdOut`

## Build

To build the command line:

```bash
$ make build
```
## Testing 

To run unit-tests and integration tests that run against the delivered web-server hosted in the `test_data` folder run:

```bash
$ make test
```

This command depends on the presence of [docker](https://www.docker.com/) on your running environment as it will:

1. Build the container used to serve the files under `test_data`
2. Start the container listening on port 8080 of the localhost
3. Run go test which will run several tests which are local unit-tests and several tests which are integration tests that
depend on the running web-server. In the code you can distinguish integration tests because they have the suffix `Integration`
   (e.g. `Test_crawler_Crawl_Cancelation_Integration`)
   
If you want to run only the unit-tests you can run the following command:

```bash
$ make utest
```


## Run

To run the binary:

```bash
$ ./web-crawler -url=<url_to_be_crawled> 
```

## Build, test and run with docker-compose

By running the following command the project will be built, will be unit-tested against a dummy local web-server 
(run within its own docker container) and the command-line will be launched against this local web-server:

```bash
$ docker-compose up --build --abort-on-container-exit
```

Also a [github actions workflow](https://github.com/features/actions) is put in place for a small CI of this repository 
to facilitate the reader to check the product outputs when run against the sample web-server.

## Tests

* Even though I support the idea of unit-testing only the public-interface (exported-methods) of a given package 
  as a way to allow implementation details
  to vary without impacting behavior from customer/client code,
  in this repo I ended-up testing some non-exported methods as a strategy to test some corner-cases and make
  the final product behavior more predictable.

## Concurrency 

In this program in order to reach high-performance on the crawling step, the program makes heavy use of go-routines. 
In the routine model of this program each new page is scraped by a different go-routine, which will also be responsible
for parse the html page and detect new links. In the go-routines processing there are two points of synchronization to
avoid data-races in the access (read or write) of the set of already scrapted urls (represented in this program as a `map[string]struct{}`).
Each go-routine therefore will:
1. scrape an HTML page after having inserted it in the set of __scraped__ urls
2. parse the HTML content of the scraped page 
3. execute the 'visit' function on that `html.Node` structure - in our case the visit method is only printing to __stdout__ 
    the page url along with the found links (without considering if those links need to be scraped)
4. extract all links in the html page
5. convert the extracted links to absolute links
6. for all the eligible links, spawn new go-routines to perform this same
   set of functionalities (from 1 to 6)
   
The algorithm follows this recursive pattern.
In order to minimize synchronization blocks we use a `sync.RWMutex` around the `map[string]struct{}` accesses. With this 
data-structures we can afford to have several go-routines reading concurrently the map and only synchronize access when there 
are mix write-read access patterns.

Here below you have a scratch of how the concurrency model works:

![concurrency-diagram](http://www.plantuml.com/plantuml/proxy?cache=no&src=https://raw.githubusercontent.com/rbroggi/web-crawler/master/concurrency.plantuml)

## limitations

* in the concurrency pattern implemented there is no limitation on the concurrency level of the program as new go-routines
  will be spawned recursively across the page scraping. This situation could lead to an excessive memory use and eventually
  to program crashes.
* not having a concurrency level setting makes it more difficult to put in place a benchmark test to evaluate the real 
  benefit of concurrency in this url scraping recursive algorithm.
* this program does not recognize anchors containing `#` as special local-page references and therefore treats it as a 
  link whose path is relative to the current page. This could be easily enhanced by filtering out links starting with `#`.

### References

* [Building a Web Crawler with Go to detect duplicate titles](https://flaviocopes.com/golang-web-crawler/)
* [Go exmples test](https://blog.golang.org/examples)
* [Absolute vs. Relative paths/links](https://www.coffeecup.com/help/articles/absolute-vs-relative-pathslinks/#:~:text=The%20first%20difference%20you'll,file%20or%20a%20file%20path.)
* [Cancel http.Request using Context](https://ferencfbin.medium.com/golang-cancel-http-request-using-context-1f45aeba6464)