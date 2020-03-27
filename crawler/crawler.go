package crawler

import (
	"context"
	"log"

	_ "github.com/lib/pq"
	. "github.com/tsushiy/codernote-backend/db"
)

type PubSubMessage struct {
	Data []byte `json:"data"`
}

func CrawlAll(ctx context.Context, m PubSubMessage) error {
	db := DbConnect(true)
	defer db.Close()
	// db.LogMode(true)

	if err := updateAtcoder(db); err != nil {
		log.Println(err)
	}
	if err := updateCodeforces(db); err != nil {
		log.Println(err)
	}
	if err := updateYukicoder(db); err != nil {
		log.Println(err)
	}
	if err := updateAOJ(db); err != nil {
		log.Println(err)
	}
	if err := updateLeetcode(db); err != nil {
		log.Println(err)
	}
	return nil
}

func CrawlAtcoder(ctx context.Context, m PubSubMessage) error {
	db := DbConnect(true)
	defer db.Close()
	if err := updateAtcoder(db); err != nil {
		log.Println(err)
	}
	return nil
}

func CrawlCodeforces(ctx context.Context, m PubSubMessage) error {
	db := DbConnect(true)
	defer db.Close()
	if err := updateCodeforces(db); err != nil {
		log.Println(err)
	}
	return nil
}

func CrawlYukicoder(ctx context.Context, m PubSubMessage) error {
	db := DbConnect(true)
	defer db.Close()
	if err := updateYukicoder(db); err != nil {
		log.Println(err)
	}
	return nil
}

func CrawlAOJ(ctx context.Context, m PubSubMessage) error {
	db := DbConnect(true)
	defer db.Close()
	if err := updateAOJ(db); err != nil {
		log.Println(err)
	}
	return nil
}

func CrawlLeetcode(ctx context.Context, m PubSubMessage) error {
	db := DbConnect(true)
	defer db.Close()
	if err := updateLeetcode(db); err != nil {
		log.Println(err)
	}
	return nil
}
