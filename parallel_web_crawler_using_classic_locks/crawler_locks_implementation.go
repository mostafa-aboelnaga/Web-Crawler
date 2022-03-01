package main

import (
	"fmt"  // to implement formatted I/O like in language C
	"sync" // to provide basic synchronization primitives such as mutual exclusion locks.
	"time" // to provide functionality for measuring and displaying time.
)

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher) {

	// TODO: Fetch URLs in parallel.
	// 			DONE using goroutines
	// TODO: Don't fetch the same URL twice. (visited_set has to be a thread-safe data structure)
	// 			DONE by letting each goroutine use a lock namely "setLock", in order to lock the visited_set

	if depth <= 0 {
		return
	}

	// each goroutine tries to have a lock on the visited set
	setLock.Lock()
	// but if there exists a lock already, everything will get blocked (already locked)

	visited := visited_set[url] // removed the blank identifier, it was noted that it was unnecessary

	if !visited {
		body, urls, err := fetcher.Fetch(url)
		visited_set[url] = true
		if err != nil {
			fmt.Println(err)

			wg.Done()        // if an error occurs, we need to decrease the counter, meaning that we are done with that goroutine
			setLock.Unlock() // we also unlock, just before returning, being done with this goroutine

			return
		}
		fmt.Printf("found:[depth:%d] %s %q\n", depth, url, body)
		for _, u := range urls {
			wg.Add(1) // increasing the counter to wait for this goroutine
			go Crawl(u, depth-1, fetcher)
		}
	}

	wg.Done()        // decreasing the counter, meaning that we are done with that goroutine
	setLock.Unlock() // unlocking, being done with this goroutine
}

func main() {
	wg.Add(1) // adding a wait group in advance (increasing the counter), to count in for the main thread

	Crawl("http://golang.org/", 4, fetcher)

	wg.Wait() // blocking the next statements, until the counter of our waitgroup gets to 0

	println("============DONE=============")

	for k := range visited_set {
		println(k)
	}

	// time.Sleep(time.Duration(30) * time.Second)
	// no need for this delay/sleep, we already print everything
	// after ensuring that we are done with fetching and all that
}

// declaring the visited set and the required locks to sync everything!
var visited_set = make(map[string]bool) // the set or list that holds our visited/fetched URLs
var setLock = &sync.Mutex{}             // the lock needed on the visited set
var wg sync.WaitGroup                   // to ensure the coordination between the goroutines

///////////////////////////////////////////////////////////////////////////////////////////////

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

//////////////////////////////////////////////////////////////////////////////////////////////

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

// every url has a body + urls
type fakeResult struct {
	body string
	urls []string
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"http://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"http://golang.org/pkg/",
			"http://golang.org/cmd/",
		},
	},
	"http://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"http://golang.org/",
			"http://golang.org/cmd/",
			"http://golang.org/pkg/fmt/",
			"http://golang.org/pkg/os/",
		},
	},
	"http://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
	"http://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {

	fmt.Printf("Fetching: %s\n", url)

	time.Sleep(500 * time.Millisecond)

	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

///////////////////////////////////////////////////////////////////////////////////////////////
