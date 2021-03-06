package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
)

type User struct {
	No        int    `gorm:"primary_key" json:"-"`
	UserID    string `gorm:"unique;not null"`
	Name      string `gorm:"unique;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserDetail struct {
	UserID       string `gorm:"primary_key"`
	AtCoderID    string
	CodeforcesID string
	YukicoderID  string
	AOJID        string
	LeetCodeID   string
}

type Contest struct {
	No               int `gorm:"primary_key"`
	Domain           string
	ContestID        string
	Title            string
	StartTimeSeconds int
	DurationSeconds  int
	Rated            string
	ProblemNoList    pq.Int64Array `gorm:"type:integer[]"`
}

type Problem struct {
	No         int `gorm:"primary_key"`
	Domain     string
	ProblemID  string
	ContestID  string
	Title      string
	Slug       string `json:"Slug,omitempty"`
	FrontendID string `json:"FrontendID,omitempty"`
	Difficulty string
}

type Note struct {
	ID        string `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Text      string
	ProblemNo int
	Problem   Problem `gorm:"foreignkey:ProblemNo"`
	UserNo    int     `json:"-"`
	User      User    `gorm:"foreignkey:UserNo"`
	Public    int     `gorm:"default:1"`
}

type Tag struct {
	No  int    `gorm:"primary_key" json:"-"`
	Key string `gorm:"unique;not null"`
}

type TagMap struct {
	No     int `gorm:"primary_key" json:"-"`
	NoteID string
	TagNo  int
}

func DbConnect(migrate bool) *gorm.DB {
	host := getEnv("POSTGRE_HOST", "localhost")
	port := getEnv("POSTGRE_PORT", "5432")
	user := getEnv("POSTGRE_USER", "postgres")
	pass := getEnv("POSTGRE_PASS", "passwd")
	dbname := getEnv("POSTGRE_DBNAME", "codernote-dev")

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		host, port, user, dbname, pass)
	log.Printf("Try connect %s", connStr)
	for i := 0; i < 3; i++ {
		db, err := gorm.Open("postgres", connStr)
		if err != nil {
			log.Printf("Cannot connect to db %d/3", i+1)
			time.Sleep(5 * time.Second)
			continue
		}

		if migrate {
			db.AutoMigrate(&User{}, &UserDetail{}, &Contest{}, &Problem{}, &Note{}, &Tag{}, &TagMap{})
		}
		return db
	}

	log.Fatal("Cannot connect db 3 times")

	return nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
