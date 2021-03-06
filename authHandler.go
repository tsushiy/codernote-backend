package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/google/uuid"
	"github.com/gorilla/mux"

	. "github.com/tsushiy/codernote-backend/db"
)

const (
	defaultNameLen = 24
	letters        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func (s *server) loginPostHandler(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(uidKey).(string)

	var user User
	if err := s.db.
		Where(User{
			UserID: uid,
		}).
		Attrs(User{
			UserID: uid,
			Name:   randStr(defaultNameLen),
		}).
		FirstOrCreate(&user).Error; err != nil {
		log.Println(err)
		http.Error(w, "failed to fetch or create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(user)
}

func (s *server) userNamePostHandler(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(uidKey).(string)

	type changeNameBody struct {
		Name string
	}
	var b changeNameBody
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(b.Name)
	err := validation.Validate(
		name,
		validation.Required,
		validation.Length(3, 30),
		validation.Match(regexp.MustCompile("^[a-zA-Z0-9_]+$")),
	)
	if err != nil {
		http.Error(w, "invalid username", http.StatusBadRequest)
		return
	}

	var user User
	if err := s.db.
		Where(User{
			UserID: uid,
		}).
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

func (s *server) userSettingGetHandler(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(uidKey).(string)

	var detail UserDetail
	if err := s.db.
		Where(UserDetail{
			UserID: uid,
		}).
		FirstOrInit(&detail).Error; err != nil {
		log.Println(err)
		http.Error(w, "failed to get setting", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(detail)
}

func (s *server) userSettingPostHandler(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(uidKey).(string)

	type changeSettingBody struct {
		AtCoderID    string
		CodeforcesID string
		YukicoderID  string
		AOJID        string
		LeetCodeID   string
	}
	var b changeSettingBody
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	elem := reflect.ValueOf(&b).Elem()
	for i := 0; i < elem.NumField(); i++ {
		x := elem.Field(i).Interface().(string)
		if err := validation.Validate(
			x,
			validation.Length(0, 100),
			validation.Match(regexp.MustCompile("^[a-zA-Z0-9_]+$")),
		); err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
	}

	var detail UserDetail
	if err := s.db.
		Where(UserDetail{
			UserID: uid,
		}).
		Assign(map[string]interface{}{
			"user_id":       uid,
			"at_coder_id":   b.AtCoderID,
			"codeforces_id": b.CodeforcesID,
			"yukicoder_id":  b.YukicoderID,
			"aoj_id":        b.AOJID,
			"leet_code_id":  b.LeetCodeID,
		}).
		FirstOrCreate(&detail).Error; err != nil {
		log.Println(err)
		http.Error(w, "failed to change setting", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(detail)
}

func (s *server) authNoteGetHandler(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(uidKey).(string)
	q := r.URL.Query()

	noteID := q.Get("noteId")
	if noteID == "" {
		http.Error(w, "invalid request path", http.StatusBadRequest)
		return
	}

	nfilter := Note{ID: noteID}
	var note Note
	if err := s.db.
		Preload("User").
		Preload("Problem").
		Where(&nfilter).
		Take(&note).Error; err != nil {
		http.Error(w, "note not found", http.StatusNotFound)
		return
	}

	if note.User.UserID != uid && note.Public == 1 {
		http.Error(w, "note not found", http.StatusNotFound)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(note)
}

func (s *server) myNoteGetHandler(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(uidKey).(string)

	vars := mux.Vars(r)
	problemNo, _ := strconv.Atoi(vars["problemNo"])
	if problemNo == 0 {
		http.Error(w, "invalid request path", http.StatusBadRequest)
		return
	}

	pfilter := Problem{No: problemNo}
	ufilter := User{UserID: uid}
	var note Note
	if err := s.db.
		Preload("User").
		Preload("Problem").
		Joins("left join problems on problems.no = notes.problem_no").
		Joins("left join users on users.no = notes.user_no").
		Where(&pfilter).
		Where(&ufilter).
		Take(&note).Error; err != nil {
		log.Println(err)
		http.Error(w, "note not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(note)
}

func (s *server) myNotePostHandler(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(uidKey).(string)

	vars := mux.Vars(r)
	problemNo, _ := strconv.Atoi(vars["problemNo"])
	if problemNo == 0 {
		http.Error(w, "invalid request path", http.StatusBadRequest)
		return
	}

	type notePostBody struct {
		Text   string
		Public bool
	}
	var b notePostBody
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if b.Text == "" {
		http.Error(w, "empty text", http.StatusBadRequest)
		return
	}
	if len(b.Text) > 1024*1024 {
		http.Error(w, "too large text", http.StatusBadRequest)
		return
	}

	var public int
	if b.Public == true {
		public = 2
	} else {
		public = 1
	}

	var problem Problem
	var user User
	if err := s.db.
		Where(Problem{
			No: problemNo,
		}).
		Take(&problem).Error; err != nil {
		http.Error(w, "no problem matched", http.StatusBadRequest)
		return
	}
	if err := s.db.
		Where(User{
			UserID: uid,
		}).
		Take(&user).Error; err != nil {
		http.Error(w, "user not registered", http.StatusBadRequest)
		return
	}

	var note Note
	randID, err := genUUID()
	if err != nil {
		log.Println(err)
		http.Error(w, "failed to create a note", http.StatusInternalServerError)
		return
	}
	if err := s.db.
		Where(Note{
			ProblemNo: problemNo,
			UserNo:    user.No,
		}).
		Attrs(Note{
			ID: randID,
		}).
		Assign(Note{
			Text:   b.Text,
			Public: public,
		}).
		FirstOrCreate(&note).Error; err != nil {
		log.Println(err)
		http.Error(w, "failed to create or update note", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(note)
}

func (s *server) myNoteDeleteHandler(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(uidKey).(string)

	vars := mux.Vars(r)
	problemNo, _ := strconv.Atoi(vars["problemNo"])
	if problemNo == 0 {
		http.Error(w, "invalid request path", http.StatusBadRequest)
		return
	}

	pfilter := Problem{No: problemNo}
	ufilter := User{UserID: uid}
	var note Note
	if err := s.db.
		Joins("left join problems on problems.no = notes.problem_no").
		Joins("left join users on users.no = notes.user_no").
		Where(&pfilter).
		Where(&ufilter).
		Take(&note).Error; err != nil {
		log.Println(err)
		http.Error(w, "note does not exist", http.StatusBadRequest)
		return
	}

	if err := s.db.Delete(&note).Error; err != nil {
		log.Println(err)
		http.Error(w, "failed to delete note", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *server) myNoteListGetHandler(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(uidKey).(string)

	q := r.URL.Query()
	domain := q.Get("domain")
	contestID := q.Get("contestId")
	tag := q.Get("tag")
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
		http.Error(w, "invalid sort order", http.StatusBadRequest)
		return
	}

	pfilter := Problem{
		Domain:    domain,
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
			Joins("left join tag_maps on tag_maps.note_id = notes.id").
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
			Limit(limit).Offset(skip).Order(order).
			Preload("User").
			Preload("Problem").
			Joins("left join problems on problems.no = notes.problem_no").
			Joins("left join users on users.no = notes.user_no").
			Where(&pfilter).
			Where(&ufilter).
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
	uid := r.Context().Value(uidKey).(string)

	vars := mux.Vars(r)
	problemNo, _ := strconv.Atoi(vars["problemNo"])
	if problemNo == 0 {
		http.Error(w, "invalid request path", http.StatusBadRequest)
		return
	}

	pfilter := Problem{No: problemNo}
	ufilter := User{UserID: uid}

	type result struct {
		Key string
	}
	var res []result
	if err := s.db.
		Model(&Note{}).
		Joins("left join problems on problems.no = notes.problem_no").
		Joins("left join users on users.no = notes.user_no").
		Joins("left join tag_maps on tag_maps.note_id = notes.id").
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
	resp := response{Tags: []string{}}
	for _, v := range res {
		resp.Tags = append(resp.Tags, v.Key)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(resp)
}

func (s *server) tagPostHandler(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(uidKey).(string)

	vars := mux.Vars(r)
	problemNo, _ := strconv.Atoi(vars["problemNo"])
	if problemNo == 0 {
		http.Error(w, "invalid request path", http.StatusBadRequest)
		return
	}

	type tagPostBody struct {
		Tag string
	}
	var b tagPostBody
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	key := strings.TrimSpace(b.Tag)
	if key == "" {
		http.Error(w, "empty tag", http.StatusBadRequest)
		return
	}
	if len(key) > 200 {
		http.Error(w, "too large tag", http.StatusBadRequest)
		return
	}
	if isInvalidTag(key) {
		http.Error(w, "invalid tag", http.StatusBadRequest)
		return
	}

	var problem Problem
	if err := s.db.
		Where(Problem{
			No: problemNo,
		}).
		Take(&problem).Error; err != nil {
		http.Error(w, "no problem matched", http.StatusBadRequest)
		return
	}
	var user User
	if err := s.db.
		Where(User{
			UserID: uid,
		}).
		Take(&user).Error; err != nil {
		http.Error(w, "user is not registered", http.StatusBadRequest)
		return
	}
	var tag Tag
	if err := s.db.
		Where(Tag{
			Key: key,
		}).
		FirstOrCreate(&tag).Error; err != nil {
		log.Println(err)
		http.Error(w, "failed to fetch tag info", http.StatusInternalServerError)
		return
	}
	var note Note
	randID, err := genUUID()
	if err != nil {
		log.Println(err)
		http.Error(w, "failed to create a note", http.StatusInternalServerError)
		return
	}
	if err := s.db.
		Where(Note{
			ProblemNo: problemNo,
			UserNo:    user.No,
		}).
		Attrs(Note{
			ID: randID,
		}).
		FirstOrCreate(&note).Error; err != nil {
		log.Println(err)
		http.Error(w, "failed to fetch note", http.StatusInternalServerError)
		return
	}
	var tagMap TagMap
	if err := s.db.
		Where(TagMap{
			NoteID: note.ID,
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
	uid := r.Context().Value(uidKey).(string)

	vars := mux.Vars(r)
	problemNo, _ := strconv.Atoi(vars["problemNo"])
	if problemNo == 0 {
		http.Error(w, "invalid request path", http.StatusBadRequest)
		return
	}

	type tagDeleteBody struct {
		Tag string
	}
	var b tagDeleteBody
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	key := strings.TrimSpace(b.Tag)
	if key == "" {
		http.Error(w, "empty tag", http.StatusBadRequest)
		return
	}
	if len(key) > 200 {
		http.Error(w, "too large tag", http.StatusBadRequest)
		return
	}

	var tag Tag
	if err := s.db.
		Where(Tag{
			Key: key,
		}).
		Take(&tag).Error; err != nil {
		http.Error(w, "tag does not exist", http.StatusBadRequest)
		return
	}

	pfilter := Problem{No: problemNo}
	ufilter := User{UserID: uid}
	var note Note
	if err := s.db.
		Joins("left join problems on problems.no = notes.problem_no").
		Joins("left join users on users.no = notes.user_no").
		Where(&pfilter).
		Where(&ufilter).
		Take(&note).Error; err != nil {
		http.Error(w, "note does not exist", http.StatusBadRequest)
		return
	}
	var tagMap TagMap
	if err := s.db.
		Where(TagMap{
			NoteID: note.ID,
			TagNo:  tag.No,
		}).
		Take(&tagMap).Error; err != nil {
		http.Error(w, "note does not have the tag", http.StatusBadRequest)
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

func genUUID() (string, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func isInvalidTag(s string) bool {
	list := []string{"<", ">", "&", "\"", "'", "/", "!", "?", "=", "$"}
	for _, v := range list {
		isDangerous := strings.Contains(s, v)
		if isDangerous == true {
			return true
		}
	}
	return false
}
