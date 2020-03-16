package main

import (
	"github.com/tsushiy/codernote-backend/crawler"
)

func main() {
	crawler.Crawl(nil, crawler.PubSubMessage{})
}
