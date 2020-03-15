package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	. "github.com/tsushiy/codernote-backend/db"
)

var db *gorm.DB

func main() {
	db = DbConnect()
	defer db.Close()
	// db.LogMode(true)

	updateAtcoder(db)
}
