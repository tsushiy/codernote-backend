package crawler

import (
	"context"

	_ "github.com/lib/pq"
	. "github.com/tsushiy/codernote-backend/db"
)

type PubSubMessage struct {
	Data []byte `json:"data"`
}

func Crawl(ctx context.Context, m PubSubMessage) error {
	db := DbConnect(true)
	defer db.Close()
	// db.LogMode(true)

	updateAtcoder(db)
	return nil
}
