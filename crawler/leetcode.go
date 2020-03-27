package crawler

import (
	"encoding/json"
	"log"
	"strconv"

	"github.com/jinzhu/gorm"
	. "github.com/tsushiy/codernote-backend/db"
)

const (
	leetcodeDomain          = "leetcode"
	leetcodeProblemsBaseURL = "https://leetcode.com/api/problems/"
)

type leetcodeProblem struct {
	UserName        string `json:"user_name"`
	NumSolved       int    `json:"num_solved"`
	NumTotal        int    `json:"num_total"`
	AcEasy          int    `json:"ac_easy"`
	AcMedium        int    `json:"ac_medium"`
	AcHard          int    `json:"ac_hard"`
	StatStatusPairs []struct {
		Stat struct {
			QuestionID          int         `json:"question_id"`
			QuestionArticleLive interface{} `json:"question__article__live"`
			QuestionArticleSlug interface{} `json:"question__article__slug"`
			QuestionTitle       string      `json:"question__title"`
			QuestionTitleSlug   string      `json:"question__title_slug"`
			QuestionHide        bool        `json:"question__hide"`
			TotalAcs            int         `json:"total_acs"`
			TotalSubmitted      int         `json:"total_submitted"`
			FrontendQuestionID  int         `json:"frontend_question_id"`
			IsNewQuestion       bool        `json:"is_new_question"`
		} `json:"stat"`
		Status     interface{} `json:"status"`
		Difficulty struct {
			Level int `json:"level"`
		} `json:"difficulty"`
		PaidOnly  bool `json:"paid_only"`
		IsFavor   bool `json:"is_favor"`
		Frequency int  `json:"frequency"`
		Progress  int  `json:"progress"`
	} `json:"stat_status_pairs"`
	FrequencyHigh int    `json:"frequency_high"`
	FrequencyMid  int    `json:"frequency_mid"`
	CategorySlug  string `json:"category_slug"`
}

func updateLeetcodeProblem(db *gorm.DB) error {
	log.Println("Start updating LeetCode contest info")
	categories := []string{"algorithms", "database", "shell", "concurrency"}

	for _, category := range categories {
		url := leetcodeProblemsBaseURL + category
		body, err := fetchAPI(url)
		if err != nil {
			return err
		}
		var ret leetcodeProblem
		if err := json.Unmarshal(body, &ret); err != nil {
			return err
		}
		var problemNoList []int64
		for _, p := range ret.StatStatusPairs {
			var problem Problem
			if err := db.
				Where(Problem{
					Domain:    leetcodeDomain,
					ProblemID: strconv.Itoa(p.Stat.QuestionID),
				}).
				Assign(Problem{
					Domain:     leetcodeDomain,
					ProblemID:  strconv.Itoa(p.Stat.QuestionID),
					ContestID:  category,
					Title:      p.Stat.QuestionTitle,
					Slug:       p.Stat.QuestionTitleSlug,
					FrontendID: strconv.Itoa(p.Stat.FrontendQuestionID),
					Difficulty: strconv.Itoa(p.Difficulty.Level),
				}).
				FirstOrCreate(&problem).Error; err != nil {
				return err
			}
			problemNoList = append(problemNoList, int64(problem.No))
		}

		if err := db.
			Where(Contest{
				Domain:    leetcodeDomain,
				ContestID: category,
			}).
			Assign(Contest{
				Domain:        leetcodeDomain,
				ContestID:     category,
				Title:         category,
				ProblemNoList: problemNoList,
			}).
			FirstOrCreate(&Contest{}).Error; err != nil {
			return err
		}
	}

	return nil
}

func updateLeetcode(db *gorm.DB) error {
	if err := updateLeetcodeProblem(db); err != nil {
		return err
	}
	return nil
}
