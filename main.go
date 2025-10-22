package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
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
	Crawl() error
}

func (q *Queue) Crawl() error {
	u := q.dequeue()
	fmt.Printf("dequeue and crawl: %s\n", u)
	res, err := http.Get(u)
	if err != nil {
		return fmt.Errorf("error fetching this url: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("recieved a non-OK HTTP status: %d %s", res.StatusCode, res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}

	fmt.Println("Body: ", string(body))

	return nil
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
	urlQueue.enqueue(link)

	err := urlQueue.Crawl()
	if err != nil {
		fmt.Fprintf(os.Stdout, "Error crawling %s", link)
	}
}
