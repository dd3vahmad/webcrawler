package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sync"
)

type Queue struct {
	elements []string
	count    int
	mu       sync.Mutex
}

func (q *Queue) enqueue(u string) {
	q.mu.Lock()
	q.count++
	q.elements = append(q.elements, u)
	q.mu.Unlock()
}

func (q *Queue) dequeue() string {
	q.mu.Lock()
	q.count--
	u := q.elements[0]
	q.elements = q.elements[1:]
	q.mu.Unlock()

	return u
}

type Crawler interface {
	Crawl(s *Set) error
	Crawled(url string) bool
}

func (q *Queue) Crawl(s *Set) error {
	url := q.dequeue()
	fmt.Printf("dequeue and crawl: %s\n", url)
	res, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error fetching this url: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("recieved a non-OK HTTP status: %d %s", res.StatusCode, res.Status)
	}

	const byteLimit = 50000
	body, err := io.ReadAll(io.LimitReader(res.Body, int64(byteLimit)))
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}

	s.SetAsCrawled(url)
	bodyStr := string(body)
	regx := regexp.MustCompile(`https?://[^\s"'<>]+`)

	matches := regx.FindAllStringSubmatch(bodyStr, -1)
	for _, match := range matches {
		url := match[0]
		fmt.Println("Match: ", url)
	}

	return nil
}

type Set struct {
	elements map[string]bool
	count    int
	rwmu     sync.RWMutex
	mu       sync.Mutex
}

func (s *Set) IsCrawled(url string) bool {
	s.rwmu.RLock()
	defer s.rwmu.RUnlock()
	return s.elements[url]
}

func (s *Set) SetAsCrawled(url string) {
	s.mu.Lock()
	if s.IsCrawled(url) {
		s.mu.Unlock()
		return
	}
	fmt.Println("Passed the isCrawled check")

	s.count++
	s.elements[url] = true
	s.mu.Unlock()
}

func main() {
	var link string
	fmt.Println("Enter the url to crawl: ")

	if _, err := fmt.Scanln(&link); err != nil {
		fmt.Fprintln(os.Stdout, "Can't get the url: ", err)
		return
	}

	if _, err := url.ParseRequestURI(link); err != nil {
		fmt.Fprintln(os.Stdout, "Oops!, that's an invalid URL")
		return
	}

	urlQueue := Queue{elements: []string{}, count: 0}
	crawledUrls := Set{elements: make(map[string]bool), count: 0}

	urlQueue.enqueue(link)
	err := urlQueue.Crawl(&crawledUrls)
	if err != nil {
		fmt.Fprintf(os.Stdout, "Error crawling %s %s\n", link, err)
	}
}
