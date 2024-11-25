package main

import (
	"fmt"
	"log"

	"github.com/gocolly/colly/v2"
)

func main() {
	// Create a new collector
	c := colly.NewCollector()

	// Verify setup with a simple print
	c.OnHTML("title", func(e *colly.HTMLElement) {
		fmt.Printf("Page Title: %s\n", e.Text)
	})

	// Visit test page
	err := c.Visit("http://go.dev")
	if err != nil {
		log.Fatal(err)
	}
}
