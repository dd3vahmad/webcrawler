package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sync"
	"github.com/PuerkitoBio/goquery"
)

type Queue struct {
	elements []string
	count    int
	total    int
	mu       sync.Mutex
}

func (q *Queue) enqueue(u string) {
	q.mu.Lock()
	q.count++
	q.total++
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
	IsCrawled(url string) bool
}

func (q *Queue) Crawl(s *Set) error {
	for q.count > 0 && s.count < 500 {
		current := q.dequeue()

		res, err := http.Get(current)
		if err != nil {
			continue
		}

		if res.StatusCode != http.StatusOK {
			res.Body.Close()
			continue
		}

		body, err := io.ReadAll(io.LimitReader(res.Body, 50000))
		if err != nil {
			continue
		}

		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			fmt.Println("Oops an error occurred %v\n", err)
			continue
		}
		title := doc.Find("title").Text()
		res.Body.Close()
		
		regx := regexp.MustCompile(`https?://[^\s"'<>]+`)
		links := regx.FindAllString(string(body), -1)

		for _, link := range links {
			if s.IsCrawled(link) {
				continue
			}
			s.SetAsCrawled(link)
			q.enqueue(link)
			fmt.Printf("Count: %d | %s -> %s\n", q.total, link, title)
		}
	}
	return nil
}

type Set struct {
	elements map[string]bool
	count    int
	mu       sync.Mutex
}

func (s *Set) IsCrawled(url string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.elements[url]
}

func (s *Set) SetAsCrawled(url string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.elements[url] {
		return
	}
	s.elements[url] = true
	s.count++
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

	urlQueue := Queue{elements: []string{}, count: 0, total: 0}
	crawledUrls := Set{elements: make(map[string]bool), count: 0}

	urlQueue.enqueue(link)
	err := urlQueue.Crawl(&crawledUrls)

	fmt.Println("\n")
	fmt.Println("------------Crawler Stats------------")
	queuedTotal := urlQueue.total
	queuedCount := urlQueue.count
	crawledCount := crawledUrls.count
	fmt.Printf(">> Total queued: %d\n", queuedTotal)
	fmt.Printf(">> To be crawled (Queue): %d\n", queuedCount)
	fmt.Printf(">> Crawled: %d\n", crawledCount)

	if err != nil {
		fmt.Fprintf(os.Stdout, "Error crawling %s %s\n", link, err)
	}
}
