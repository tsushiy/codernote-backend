package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"

	. "github.com/tsushiy/codernote-backend/db"
)

const (
	defaultNameLen = 24
	letters        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func (s *server) loginPostHandler(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value("uid").(string)

	var user User
	if err := s.db.
		Where(User{UserID: uid}).
		Attrs(User{UserID: uid, Name: randStr(defaultNameLen)}).
		FirstOrCreate(&user).Error; err != nil {
		log.Println(err)
		http.Error(w, "failed to fetch or create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(user)
}

func (s *server) namePostHandler(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value("uid").(string)
	type changeNameBody struct {
		Name string
	}
	var b changeNameBody
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		log.Println(err)
		http.Error(w, "invalid request body", http.StatusInternalServerError)
		return
	}

	// username must be between 3 and 30 alphanumeric characters
	name := strings.TrimSpace(b.Name)
	err := validation.Validate(
		name,
		validation.Required,
		validation.Length(3, 30),
		is.Alphanumeric,
	)
	if err != nil {
		log.Println(err)
		http.Error(w, "invalid username", http.StatusInternalServerError)
		return
	}

	var user User
	if err := s.db.
		Where(User{UserID: uid}).
		Assign(User{
			UserID: uid,
			Name:   name,
		}).
		FirstOrCreate(&user).Error; err != nil {
		log.Println(err)
		http.Error(w, "failed to change username", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(user)
}

func (s *server) noteGetHandler(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value("uid").(string)
	q := r.URL.Query()
	domain := q.Get("domain")
	contestID := q.Get("contestId")
	problemID := q.Get("problemId")
	if domain == "" || contestID == "" || problemID == "" {
		http.Error(w, "invalid request body", http.StatusInternalServerError)
		return
	}

	pfilter := Problem{
		Domain:    domain,
		ProblemID: problemID,
		ContestID: contestID,
	}
	ufilter := User{
		UserID: uid,
	}
	var note Note
	if err := s.db.
		Preload("User").
		Preload("Problem").
		Joins("left join problems on problems.no = notes.problem_no").
		Joins("left join users on users.no = notes.user_no").
		Where(&pfilter).
		Where(&ufilter).
		FirstOrInit(&note).Error; err != nil {
		log.Println(err)
		http.Error(w, "failed to fetch note", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(note)
}

func (s *server) notePostHandler(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value("uid").(string)
	type notePostBody struct {
		Domain    string
		ContestID string
		ProblemID string
		Text      string
		Public    bool
	}
	var b notePostBody
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		log.Println(err)
		http.Error(w, "invalid request body", http.StatusInternalServerError)
		return
	}
	if b.Domain == "" || b.ContestID == "" || b.ProblemID == "" {
		http.Error(w, "invalid problem", http.StatusInternalServerError)
		return
	}

	if b.Text == "" {
		http.Error(w, "empty text", http.StatusInternalServerError)
		return
	}
	if len(b.Text) > 1024*1024 {
		http.Error(w, "too large text", http.StatusInternalServerError)
		return
	}

	var problem Problem
	var user User
	if err := s.db.
		Where(Problem{
			Domain:    b.Domain,
			ProblemID: b.ProblemID,
			ContestID: b.ContestID,
		}).
		Take(&problem).Error; err != nil {
		log.Println(err)
		http.Error(w, "no problem matched", http.StatusInternalServerError)
		return
	}
	if err := s.db.
		Where(User{UserID: uid}).
		Take(&user).Error; err != nil {
		log.Println(err)
		http.Error(w, "user not registered", http.StatusInternalServerError)
		return
	}

	var note Note
	if err := s.db.
		Where(Note{
			ProblemNo: problem.No,
			UserNo:    user.No,
		}).
		Assign(Note{
			Text:   b.Text,
			Public: b.Public,
		}).
		FirstOrCreate(&note).Error; err != nil {
		log.Println(err)
		http.Error(w, "failed to create or update note", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *server) myNoteListGetHandler(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value("uid").(string)
	q := r.URL.Query()
	domain := q.Get("domain")
	contestID := q.Get("contestId")
	problemID := q.Get("problemId")
	tag := q.Get("tag")
	limit, _ := strconv.Atoi(q.Get("limit"))
	skip, _ := strconv.Atoi(q.Get("skip"))
	order := q.Get("order")

	if 1000 < limit {
		limit = 1000
	} else if limit < 20 {
		limit = 20
	}
	if order == "" || order == "-updated" {
		order = "updated_at desc"
	} else {
		http.Error(w, "invalid sort order", http.StatusInternalServerError)
		return
	}

	pfilter := Problem{
		Domain:    domain,
		ProblemID: problemID,
		ContestID: contestID,
	}
	ufilter := User{UserID: uid}
	tfilter := Tag{Key: tag}

	// ここ綺麗に書きたい
	count := 0
	if tag == "" {
		if err := s.db.
			Model(&Note{}).
			Joins("left join problems on problems.no = notes.problem_no").
			Joins("left join users on users.no = notes.user_no").
			Where(&pfilter).
			Where(&ufilter).
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
			Count(&count).Error; err != nil {
			log.Println(err)
			http.Error(w, "failed to count", http.StatusInternalServerError)
			return
		}
	}

	var notes []Note
	if tag == "" {
		if err := s.db.
			Preload("User").
			Preload("Problem").
			Joins("left join problems on problems.no = notes.problem_no").
			Joins("left join users on users.no = notes.user_no").
			Where(&pfilter).
			Where(&ufilter).
			Limit(limit).Offset(skip).Order(order).
			Find(&notes).Error; err != nil {
			log.Println(err)
			http.Error(w, "failed to fetch note list", http.StatusInternalServerError)
			return
		}
	} else {
		if err := s.db.
			Preload("User").
			Preload("Problem").
			Joins("left join problems on problems.no = notes.problem_no").
			Joins("left join users on users.no = notes.user_no").
			Joins("left join tag_maps on tag_maps.note_no = notes.no").
			Joins("left join tags on tags.no = tag_maps.tag_no").
			Where(&pfilter).
			Where(&ufilter).
			Where(&tfilter).
			Limit(limit).Offset(skip).Order(order).
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

func (s *server) tagGetHandler(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value("uid").(string)
	q := r.URL.Query()
	domain := q.Get("domain")
	contestID := q.Get("contestId")
	problemID := q.Get("problemId")
	if domain == "" || contestID == "" || problemID == "" {
		http.Error(w, "invalid problem", http.StatusInternalServerError)
		return
	}

	pfilter := Problem{
		Domain:    domain,
		ProblemID: problemID,
		ContestID: contestID,
	}
	ufilter := User{UserID: uid}

	type result struct {
		Key string
	}
	var res []result
	if err := s.db.
		Model(&Note{}).
		Joins("left join problems on problems.no = notes.problem_no").
		Joins("left join users on users.no = notes.user_no").
		Joins("left join tag_maps on tag_maps.note_no = notes.no").
		Joins("left join tags on tags.no = tag_maps.tag_no").
		Where(&pfilter).
		Where(&ufilter).
		Select("key").
		Scan(&res).Error; err != nil {
		log.Println(err)
		http.Error(w, "failed to get tags", http.StatusInternalServerError)
		return
	}
	type response struct {
		Tags []string
	}
	resp := response{}
	for _, v := range res {
		resp.Tags = append(resp.Tags, v.Key)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(resp)
}

func (s *server) tagPostHandler(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value("uid").(string)
	type tagPostBody struct {
		Domain    string
		ContestID string
		ProblemID string
		Tag       string
	}
	var b tagPostBody
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		log.Println(err)
		http.Error(w, "invalid request body", http.StatusInternalServerError)
		return
	}
	if b.Domain == "" || b.ContestID == "" || b.ProblemID == "" {
		http.Error(w, "invalid problem", http.StatusInternalServerError)
		return
	}

	b.Tag = strings.TrimSpace(b.Tag)
	if b.Tag == "" {
		http.Error(w, "empty tag", http.StatusInternalServerError)
		return
	}
	if len(b.Tag) > 200 {
		http.Error(w, "too large tag", http.StatusInternalServerError)
		return
	}

	var problem Problem
	if err := s.db.
		Where(Problem{
			Domain:    b.Domain,
			ProblemID: b.ProblemID,
			ContestID: b.ContestID,
		}).
		Take(&problem).Error; err != nil {
		log.Println(err)
		http.Error(w, "no problem matched", http.StatusInternalServerError)
		return
	}
	var user User
	if err := s.db.
		Where(User{UserID: uid}).
		Take(&user).Error; err != nil {
		log.Println(err)
		http.Error(w, "user is not registered", http.StatusInternalServerError)
		return
	}
	var tag Tag
	if err := s.db.
		Where(Tag{Key: b.Tag}).
		FirstOrCreate(&tag).Error; err != nil {
		log.Println(err)
		http.Error(w, "failed to fetch tag info", http.StatusInternalServerError)
		return
	}
	var note Note
	if err := s.db.
		Where(Note{
			ProblemNo: problem.No,
			UserNo:    user.No,
		}).
		FirstOrCreate(&note).Error; err != nil {
		log.Println(err)
		http.Error(w, "failed to fetch note", http.StatusInternalServerError)
		return
	}
	var tagMap TagMap
	if err := s.db.
		Where(TagMap{
			NoteNo: note.No,
			TagNo:  tag.No,
		}).
		FirstOrCreate(&tagMap).Error; err != nil {
		log.Println(err)
		http.Error(w, "failed to create note-tag map", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *server) tagDeleteHandler(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value("uid").(string)
	type tagDeleteBody struct {
		Domain    string
		ContestID string
		ProblemID string
		Tag       string
	}
	var b tagDeleteBody
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		log.Println(err)
		http.Error(w, "invalid request body", http.StatusInternalServerError)
		return
	}
	if b.Domain == "" || b.ContestID == "" || b.ProblemID == "" {
		http.Error(w, "invalid problem", http.StatusInternalServerError)
		return
	}

	b.Tag = strings.TrimSpace(b.Tag)
	if b.Tag == "" {
		http.Error(w, "empty tag", http.StatusInternalServerError)
		return
	}
	if len(b.Tag) > 200 {
		http.Error(w, "too large tag", http.StatusInternalServerError)
		return
	}

	var tag Tag
	if err := s.db.
		Where(Tag{Key: b.Tag}).
		Take(&tag).Error; err != nil {
		log.Println(err)
		http.Error(w, "tag does not exist", http.StatusInternalServerError)
		return
	}

	pfilter := Problem{
		Domain:    b.Domain,
		ProblemID: b.ProblemID,
		ContestID: b.ContestID,
	}
	ufilter := User{UserID: uid}
	var note Note
	if err := s.db.
		Joins("left join problems on problems.no = notes.problem_no").
		Joins("left join users on users.no = notes.user_no").
		Where(&pfilter).
		Where(&ufilter).
		Take(&note).Error; err != nil {
		log.Println(err)
		http.Error(w, "note does not exist", http.StatusInternalServerError)
		return
	}
	var tagMap TagMap
	if err := s.db.
		Where(TagMap{
			NoteNo: note.No,
			TagNo:  tag.No,
		}).
		Take(&tagMap).Error; err != nil {
		log.Println(err)
		http.Error(w, "note does not have the tag", http.StatusInternalServerError)
		return
	}
	if err := s.db.Delete(&tagMap).Error; err != nil {
		log.Println(err)
		http.Error(w, "failed to delete tag", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

var randSrc = rand.NewSource(time.Now().UnixNano())

func randStr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Int63()%int64(len(letters))]
	}
	return string(b)
}
