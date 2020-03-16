package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	. "github.com/tsushiy/codernote-backend/db"
)

type server struct {
	db *gorm.DB
}

func main() {
	s := &server{}
	s.db = DbConnect(false)
	defer s.db.Close()
	// s.db.LogMode(true)

	router := mux.NewRouter()
	router.Use(loggerMiddleware)

	nonAuthRouter := router.NewRoute().Subrouter()
	nonAuthRouter.HandleFunc("/problems", s.problemsHandler).Methods("GET")
	nonAuthRouter.HandleFunc("/contests", s.contestsHandler).Methods("GET")
	nonAuthRouter.HandleFunc("/notes", s.publicNoteListGetHandler).Methods("GET")

	authRouter := router.NewRoute().Subrouter()
	authRouter.Use(authMiddleware)
	authRouter.HandleFunc("/login", s.loginHandler).Methods("POST")
	authRouter.HandleFunc("/user/changename", s.changeNameHandler).Methods("POST")
	authRouter.HandleFunc("/user/note", s.noteGetHandler).Methods("GET")
	authRouter.HandleFunc("/user/note", s.notePostHandler).Methods("POST")
	authRouter.HandleFunc("/user/notes", s.myNoteListGetHandler).Methods("GET")
	authRouter.HandleFunc("/user/note/tag", s.tagGetHandler).Methods("GET")
	authRouter.HandleFunc("/user/note/tag", s.tagPostHandler).Methods("POST")
	authRouter.HandleFunc("/user/note/tag", s.tagDeleteHandler).Methods("DELETE")

	srv := &http.Server{
		Handler:      router,
		Addr:         "localhost:8000",
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	log.Println("Listen Server ....")
	log.Fatal(srv.ListenAndServe())
}
