package main

import (
	"fmt"
	"net/url"
	"os"
)

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

	fmt.Printf("Crawling %s\n", link)
}
