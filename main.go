package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
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
	nonAuthRouter.HandleFunc("/healthcheck", s.healthcheckHandler).Methods("GET")
	nonAuthRouter.HandleFunc("/problems", s.problemsGetHandler).Methods("GET")
	nonAuthRouter.HandleFunc("/contests", s.contestsGetHandler).Methods("GET")
	nonAuthRouter.HandleFunc("/notes", s.publicNoteListGetHandler).Methods("GET")

	authRouter := router.NewRoute().Subrouter()
	authRouter.Use(authMiddleware)
	authRouter.HandleFunc("/login", s.loginPostHandler).Methods("POST")
	authRouter.HandleFunc("/user/name", s.namePostHandler).Methods("POST")
	authRouter.HandleFunc("/user/note", s.noteGetHandler).Methods("GET")
	authRouter.HandleFunc("/user/note", s.notePostHandler).Methods("POST")
	authRouter.HandleFunc("/user/notes", s.myNoteListGetHandler).Methods("GET")
	authRouter.HandleFunc("/user/note/tag", s.tagGetHandler).Methods("GET")
	authRouter.HandleFunc("/user/note/tag", s.tagPostHandler).Methods("POST")
	authRouter.HandleFunc("/user/note/tag", s.tagDeleteHandler).Methods("DELETE")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodDelete},
		AllowedHeaders: []string{"*"},
	})

	srv := &http.Server{
		Handler:      c.Handler(router),
		Addr:         ":" + port,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	log.Println("Listen Server ....")
	log.Fatal(srv.ListenAndServe())
}
