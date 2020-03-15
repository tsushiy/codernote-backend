package main

import (
	"encoding/json"
	"log"
	"net/http"
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

func (s *server) problemsHandler(w http.ResponseWriter, r *http.Request) {
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

func (s *server) contestsHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	domain := q.Get("domain")

	var contests []Contest
	if err := s.db.
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

func (s *server) publicNoteListGetHandler(w http.ResponseWriter, r *http.Request) {
	type noteListGetBody struct {
		Domain    string
		ContestID string
		ProblemID string
		Tag       string
		UserName  string
		Limit     int
		Skip      int
		Order     string
	}
	var b noteListGetBody
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		log.Println(err)
		http.Error(w, "invalid request body", http.StatusInternalServerError)
		return
	}

	if 1000 < b.Limit {
		b.Limit = 1000
	} else if b.Limit < 20 {
		b.Limit = 20
	}
	order := ""
	if b.Order == "" || b.Order == "-updated" {
		order = "updated_at desc"
	} else {
		http.Error(w, "invalid sort order", http.StatusInternalServerError)
		return
	}

	pfilter := Problem{
		Domain:    b.Domain,
		ProblemID: b.ProblemID,
		ContestID: b.ContestID,
	}
	ufilter := User{Name: b.UserName}
	tfilter := Tag{Key: b.Tag}
	nfilter := Note{Public: true}

	// ここ綺麗に書きたい
	count := 0
	if b.Tag == "" {
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
			Joins("left join tag_maps on tag_maps.note_no = notes.no").
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
	if b.Tag == "" {
		if err := s.db.Limit(b.Limit).Offset(b.Skip).
			Preload("User").
			Preload("Problem").
			Joins("left join problems on problems.no = notes.problem_no").
			Joins("left join users on users.no = notes.user_no").
			Where(&pfilter).
			Where(&ufilter).
			Where(&nfilter).
			Limit(b.Limit).Offset(b.Skip).Order(order).
			Find(&notes).Error; err != nil {
			log.Println(err)
			http.Error(w, "failed to fetch note list", http.StatusInternalServerError)
			return
		}
	} else {
		if err := s.db.Limit(b.Limit).Offset(b.Skip).
			Preload("User").
			Preload("Problem").
			Joins("left join problems on problems.no = notes.problem_no").
			Joins("left join users on users.no = notes.user_no").
			Joins("left join tag_maps on tag_maps.note_no = notes.no").
			Joins("left join tags on tags.no = tag_maps.tag_no").
			Where(&pfilter).
			Where(&ufilter).
			Where(&tfilter).
			Where(&nfilter).
			Limit(b.Limit).Offset(b.Skip).Order(order).
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
