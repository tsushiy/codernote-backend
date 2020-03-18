package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	. "github.com/tsushiy/codernote-backend/db"
)

func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Printf("Completed %s %s in %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func (s *server) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *server) problemsGetHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	domain := q.Get("domain")

	var problems []Problem
	if err := s.db.
		Where(Problem{
			Domain: domain,
		}).
		Find(&problems).Error; err != nil {
		log.Println(err)
		http.Error(w, "failed to get problems", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(problems)
}

func (s *server) contestsGetHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	domain := q.Get("domain")
	order := q.Get("order")

	if order == "" || order == "-started" {
		order = "start_time_seconds desc"
	} else if order == "started" {
		order = "start_time_seconds asc"
	} else {
		http.Error(w, "invalid sort order", http.StatusInternalServerError)
		return
	}

	var contests []Contest
	if err := s.db.
		Order(order).
		Where(Contest{
			Domain: domain,
		}).
		Find(&contests).Error; err != nil {
		log.Println(err)
		http.Error(w, "failed to get contests", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(contests)
}

func (s *server) publicNoteGetHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	noteID := q.Get("noteId")
	if noteID == "" {
		http.Error(w, "invalid request path", http.StatusInternalServerError)
		return
	}

	nfilter := Note{ID: noteID}
	var note Note
	if err := s.db.
		Preload("User").
		Preload("Problem").
		Where(&nfilter).
		FirstOrInit(&note).Error; err != nil {
		log.Println(err)
		http.Error(w, "failed to fetch note", http.StatusInternalServerError)
		return
	}

	if note.Public == 1 {
		note = Note{}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(note)
}

func (s *server) publicNoteListGetHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	domain := q.Get("domain")
	problemNo, _ := strconv.Atoi(q.Get("problemNo"))
	contestID := q.Get("contestId")
	tag := q.Get("tag")
	userName := q.Get("userName")
	limit, _ := strconv.Atoi(q.Get("limit"))
	skip, _ := strconv.Atoi(q.Get("skip"))
	order := q.Get("order")

	if 1000 < limit {
		limit = 1000
	} else if limit == 0 {
		limit = 100
	}
	if order == "" || order == "-updated" {
		order = "updated_at desc"
	} else {
		http.Error(w, "invalid sort order", http.StatusInternalServerError)
		return
	}

	pfilter := Problem{
		Domain:    domain,
		No:        problemNo,
		ContestID: contestID,
	}
	ufilter := User{Name: userName}
	tfilter := Tag{Key: tag}
	nfilter := Note{Public: 2}

	// ここ綺麗に書きたい
	count := 0
	if tag == "" {
		if err := s.db.
			Model(&Note{}).
			Joins("left join problems on problems.no = notes.problem_no").
			Joins("left join users on users.no = notes.user_no").
			Where(&pfilter).
			Where(&ufilter).
			Where(&nfilter).
			Count(&count).Error; err != nil {
			log.Println(err)
			http.Error(w, "failed to count", http.StatusInternalServerError)
			return
		}
	} else {
		if err := s.db.
			Model(&Note{}).
			Joins("left join problems on problems.no = notes.problem_no").
			Joins("left join users on users.no = notes.user_no").
			Joins("left join tag_maps on tag_maps.note_id = notes.id").
			Joins("left join tags on tags.no = tag_maps.tag_no").
			Where(&pfilter).
			Where(&ufilter).
			Where(&tfilter).
			Where(&nfilter).
			Count(&count).Error; err != nil {
			log.Println(err)
			http.Error(w, "failed to count", http.StatusInternalServerError)
			return
		}
	}

	var notes []Note
	if tag == "" {
		if err := s.db.
			Limit(limit).Offset(skip).Order(order).
			Preload("User").
			Preload("Problem").
			Joins("left join problems on problems.no = notes.problem_no").
			Joins("left join users on users.no = notes.user_no").
			Where(&pfilter).
			Where(&ufilter).
			Where(&nfilter).
			Find(&notes).Error; err != nil {
			log.Println(err)
			http.Error(w, "failed to fetch note list", http.StatusInternalServerError)
			return
		}
	} else {
		if err := s.db.
			Limit(limit).Offset(skip).Order(order).
			Preload("User").
			Preload("Problem").
			Joins("left join problems on problems.no = notes.problem_no").
			Joins("left join users on users.no = notes.user_no").
			Joins("left join tag_maps on tag_maps.note_id = notes.id").
			Joins("left join tags on tags.no = tag_maps.tag_no").
			Where(&pfilter).
			Where(&ufilter).
			Where(&tfilter).
			Where(&nfilter).
			Find(&notes).Error; err != nil {
			log.Println(err)
			http.Error(w, "failed to fetch note list", http.StatusInternalServerError)
			return
		}
	}

	type noteListResp struct {
		Count int
		Notes []Note
	}
	resp := noteListResp{
		Count: count,
		Notes: notes,
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(resp)
}
