package main

import (
	"github.com/tsushiy/codernote-backend/crawler"
)

func main() {
	crawler.CrawlAll(nil, crawler.PubSubMessage{})
}
