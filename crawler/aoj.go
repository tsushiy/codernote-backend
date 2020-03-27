package crawler

import (
	"encoding/json"
	"log"

	"github.com/jinzhu/gorm"
	. "github.com/tsushiy/codernote-backend/db"
)

const (
	aojDomain                  = "aoj"
	aojFilterURL               = "https://judgeapi.u-aizu.ac.jp/problems/filters"
	aojCoursesURL              = "https://judgeapi.u-aizu.ac.jp/courses"
	aojProblemsURL             = "https://judgeapi.u-aizu.ac.jp/problems?page=0&size=20000"
	aojCategoryProblemsBaseURL = "https://judgeapi.u-aizu.ac.jp/problems/cl/"
	aojCourseProblemsBaseURL   = "https://judgeapi.u-aizu.ac.jp/problems/courses/"
)

type aojFilter struct {
	Volumes  []int    `json:"volumes"`
	LargeCls []string `json:"largeCls"`
}

type aojCourse struct {
	Filter  interface{} `json:"filter"`
	Courses []struct {
		ID             int         `json:"id"`
		Serial         int         `json:"serial"`
		ShortName      string      `json:"shortName"`
		Name           string      `json:"name"`
		Type           string      `json:"type"`
		UserScore      int         `json:"userScore"`
		MaxScore       int         `json:"maxScore"`
		Progress       float64     `json:"progress"`
		Image          string      `json:"image"`
		NumberOfTopics interface{} `json:"numberOfTopics"`
		Topics         interface{} `json:"topics"`
		Description    string      `json:"description"`
	} `json:"courses"`
}

type aojProblem struct {
	ID                 string  `json:"id"`
	Available          int     `json:"available"`
	Doctype            int     `json:"doctype"`
	Name               string  `json:"name"`
	ProblemTimeLimit   int     `json:"problemTimeLimit"`
	ProblemMemoryLimit int     `json:"problemMemoryLimit"`
	MaxScore           int     `json:"maxScore"`
	SolvedUser         int     `json:"solvedUser"`
	Submissions        int     `json:"submissions"`
	Recommendations    int     `json:"recommendations"`
	IsSolved           bool    `json:"isSolved"`
	Bookmark           bool    `json:"bookmark"`
	Recommend          bool    `json:"recommend"`
	SuccessRate        float64 `json:"successRate"`
	Score              float64 `json:"score"`
	UserScore          int     `json:"userScore"`
}

type aojCategoryProblems struct {
	Progress         float64 `json:"progress"`
	NumberOfProblems int     `json:"numberOfProblems"`
	NumberOfSolved   int     `json:"numberOfSolved"`
	Problems         []struct {
		ID                 string  `json:"id"`
		Available          int     `json:"available"`
		Doctype            int     `json:"doctype"`
		Name               string  `json:"name"`
		ProblemTimeLimit   int     `json:"problemTimeLimit"`
		ProblemMemoryLimit int     `json:"problemMemoryLimit"`
		MaxScore           int     `json:"maxScore"`
		SolvedUser         int     `json:"solvedUser"`
		Submissions        int     `json:"submissions"`
		Recommendations    int     `json:"recommendations"`
		IsSolved           bool    `json:"isSolved"`
		Bookmark           bool    `json:"bookmark"`
		Recommend          bool    `json:"recommend"`
		SuccessRate        float64 `json:"successRate"`
		Score              float64 `json:"score"`
		UserScore          int     `json:"userScore"`
	} `json:"problems"`
}

type aojCourseProblems struct {
	Progress         float64 `json:"progress"`
	NumberOfProblems int     `json:"numberOfProblems"`
	NumberOfSolved   int     `json:"numberOfSolved"`
	Problems         []struct {
		ID                 string  `json:"id"`
		Available          int     `json:"available"`
		Doctype            int     `json:"doctype"`
		Name               string  `json:"name"`
		ProblemTimeLimit   int     `json:"problemTimeLimit"`
		ProblemMemoryLimit int     `json:"problemMemoryLimit"`
		MaxScore           int     `json:"maxScore"`
		SolvedUser         int     `json:"solvedUser"`
		Submissions        int     `json:"submissions"`
		Recommendations    int     `json:"recommendations"`
		IsSolved           bool    `json:"isSolved"`
		Bookmark           bool    `json:"bookmark"`
		Recommend          bool    `json:"recommend"`
		SuccessRate        float64 `json:"successRate"`
		Score              float64 `json:"score"`
		UserScore          int     `json:"userScore"`
	} `json:"problems"`
}

func getAOJCategories() ([]string, error) {
	body, err := fetchAPI(aojFilterURL)
	if err != nil {
		return nil, err
	}
	var filter aojFilter
	if err := json.Unmarshal(body, &filter); err != nil {
		return nil, err
	}
	return filter.LargeCls, nil
}

func getAOJCourses() ([]string, error) {
	body, err := fetchAPI(aojCoursesURL)
	if err != nil {
		return nil, err
	}
	var ret aojCourse
	if err := json.Unmarshal(body, &ret); err != nil {
		return nil, err
	}
	var courses []string
	for _, v := range ret.Courses {
		courses = append(courses, v.ShortName)
	}
	return courses, nil
}

func updateAOJProblems(db *gorm.DB) error {
	log.Println("Start updating aoj problem info")
	body, err := fetchAPI(aojProblemsURL)
	if err != nil {
		return err
	}
	var problems []aojProblem
	if err := json.Unmarshal(body, &problems); err != nil {
		return err
	}

	for _, v := range problems {
		var problem Problem
		if err := db.
			Where(Problem{
				Domain:    aojDomain,
				ProblemID: v.ID,
			}).
			Assign(Problem{
				Domain:    aojDomain,
				ProblemID: v.ID,
				Title:     v.Name,
			}).
			FirstOrCreate(&problem).Error; err != nil {
			return err
		}
	}

	return nil
}

func updateAOJContests(db *gorm.DB) error {
	log.Println("Start updating aoj contest info")
	categories, err := getAOJCategories()
	if err != nil {
		return nil
	}

	for _, v := range categories {
		url := aojCategoryProblemsBaseURL + v
		body, err := fetchAPI(url)
		if err != nil {
			return err
		}
		var ret aojCategoryProblems
		if err := json.Unmarshal(body, &ret); err != nil {
			return err
		}
		problems := ret.Problems

		var problemNoList []int64
		for _, p := range problems {
			var problem Problem
			if err := db.
				Where(Problem{
					Domain:    aojDomain,
					ProblemID: p.ID,
					Title:     p.Name,
				}).
				First(&problem).Error; err != nil {
				if gorm.IsRecordNotFoundError(err) {
					log.Println(err)
					continue
				} else {
					return err
				}
			}
			problemNoList = append(problemNoList, int64(problem.No))
		}

		if err := db.
			Where(Contest{
				Domain:    aojDomain,
				ContestID: v,
			}).
			Assign(Contest{
				Domain:        aojDomain,
				ContestID:     v,
				Title:         v,
				ProblemNoList: problemNoList,
			}).
			FirstOrCreate(&Contest{}).Error; err != nil {
			return err
		}
	}

	courses, err := getAOJCourses()
	if err != nil {
		return nil
	}

	for _, v := range courses {
		url := aojCourseProblemsBaseURL + v
		body, err := fetchAPI(url)
		if err != nil {
			return err
		}
		var ret aojCourseProblems
		if err := json.Unmarshal(body, &ret); err != nil {
			return err
		}
		problems := ret.Problems

		var problemNoList []int64
		for _, p := range problems {
			var problem Problem
			if err := db.
				Where(Problem{
					Domain:    aojDomain,
					ProblemID: p.ID,
					Title:     p.Name,
				}).
				First(&problem).Error; err != nil {
				if gorm.IsRecordNotFoundError(err) {
					log.Println(err)
					continue
				} else {
					return err
				}
			}
			problemNoList = append(problemNoList, int64(problem.No))
		}

		if err := db.
			Where(Contest{
				Domain:    aojDomain,
				ContestID: v,
			}).
			Assign(Contest{
				Domain:        aojDomain,
				ContestID:     v,
				Title:         v,
				ProblemNoList: problemNoList,
			}).
			FirstOrCreate(&Contest{}).Error; err != nil {
			return err
		}
	}

	return nil
}

func updateAOJ(db *gorm.DB) error {
	if err := updateAOJProblems(db); err != nil {
		return err
	}
	if err := updateAOJContests(db); err != nil {
		return err
	}
	return nil
}
