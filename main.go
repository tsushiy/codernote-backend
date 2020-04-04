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
	nonAuthRouter.HandleFunc("/note", s.publicNoteGetHandler).Methods("GET")
	nonAuthRouter.HandleFunc("/notes", s.publicNoteListGetHandler).Methods("GET")

	authRouter := router.NewRoute().Subrouter()
	authRouter.Use(authMiddleware)
	authRouter.HandleFunc("/login", s.loginPostHandler).Methods("POST")
	authRouter.HandleFunc("/user/name", s.userNamePostHandler).Methods("POST")
	authRouter.HandleFunc("/user/setting", s.userSettingGetHandler).Methods("GET")
	authRouter.HandleFunc("/user/setting", s.userSettingPostHandler).Methods("POST")
	authRouter.HandleFunc("/user/note", s.authNoteGetHandler).Methods("GET")
	authRouter.HandleFunc("/user/note/{problemNo:[0-9]+}", s.myNoteGetHandler).Methods("GET")
	authRouter.HandleFunc("/user/note/{problemNo:[0-9]+}", s.myNotePostHandler).Methods("POST")
	authRouter.HandleFunc("/user/note/{problemNo:[0-9]+}", s.myNoteDeleteHandler).Methods("DELETE")
	authRouter.HandleFunc("/user/notes", s.myNoteListGetHandler).Methods("GET")
	authRouter.HandleFunc("/user/note/{problemNo:[0-9]+}/tag", s.tagGetHandler).Methods("GET")
	authRouter.HandleFunc("/user/note/{problemNo:[0-9]+}/tag", s.tagPostHandler).Methods("POST")
	authRouter.HandleFunc("/user/note/{problemNo:[0-9]+}/tag", s.tagDeleteHandler).Methods("DELETE")

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
