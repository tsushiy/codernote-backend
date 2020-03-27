package crawler

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	. "github.com/tsushiy/codernote-backend/db"
)

const (
	yukicoderDomain      = "yukicoder"
	yukicoderProblemsURL = "https://yukicoder.me/api/v1/problems"
	yukicoderContestsURL = "https://yukicoder.me/api/v1/contest/past"
)

type yukicoderProblem struct {
	No          int       `json:"No"`
	ProblemID   int       `json:"ProblemId"`
	Title       string    `json:"Title"`
	AuthorID    int       `json:"AuthorId"`
	TesterID    int       `json:"TesterId"`
	Level       float32   `json:"Level"`
	ProblemType int       `json:"ProblemType"`
	Tags        string    `json:"Tags"`
	Date        time.Time `json:"Date"`
}

type yukicoderContest struct {
	ID            int       `json:"Id"`
	Name          string    `json:"Name"`
	Date          time.Time `json:"Date"`
	EndDate       time.Time `json:"EndDate"`
	ProblemIDList []int     `json:"ProblemIdList"`
}

var yukicoderProblemNoMap = make(map[string]int)

func updateYukicoderProblems(db *gorm.DB) error {
	log.Println("Start updating yukicoder problem info")
	body, err := fetchAPI(yukicoderProblemsURL)
	if err != nil {
		return err
	}
	var problems []yukicoderProblem
	if err := json.Unmarshal(body, &problems); err != nil {
		return err
	}

	for _, v := range problems {
		var problem Problem
		problemID := strconv.Itoa(v.ProblemID)
		if err := db.
			Where(Problem{
				Domain:    yukicoderDomain,
				ProblemID: problemID,
			}).
			Assign(Problem{
				Domain:    yukicoderDomain,
				ProblemID: problemID,
				Title:     v.Title,
			}).
			FirstOrCreate(&problem).Error; err != nil {
			return err
		}
		yukicoderProblemNoMap[problemID] = problem.No
	}

	return nil
}

func updateYukicoderContests(db *gorm.DB) error {
	log.Println("Start updating yukicoder contest info")
	body, err := fetchAPI(yukicoderContestsURL)
	if err != nil {
		return err
	}
	var contests []yukicoderContest
	if err := json.Unmarshal(body, &contests); err != nil {
		return err
	}

	for _, v := range contests {
		var problemNoList []int64
		for _, id := range v.ProblemIDList {
			problemNo := yukicoderProblemNoMap[strconv.Itoa(id)]
			problemNoList = append(problemNoList, int64(problemNo))
			if err := db.
				Where(Problem{
					No: problemNo,
				}).
				Assign(Problem{
					ContestID: strconv.Itoa(v.ID),
				}).
				FirstOrCreate(&Problem{}).Error; err != nil {
				return err
			}
		}
		if err := db.
			Where(Contest{
				Domain:    yukicoderDomain,
				ContestID: strconv.Itoa(v.ID),
			}).
			Assign(Contest{
				Domain:           yukicoderDomain,
				ContestID:        strconv.Itoa(v.ID),
				Title:            v.Name,
				StartTimeSeconds: int(v.Date.Unix()),
				DurationSeconds:  int(v.EndDate.Unix()) - int(v.Date.Unix()),
				ProblemNoList:    problemNoList,
			}).
			FirstOrCreate(&Contest{}).Error; err != nil {
			return err
		}
	}

	return nil
}

func updateYukicoder(db *gorm.DB) error {
	if err := updateYukicoderProblems(db); err != nil {
		return err
	}
	if err := updateYukicoderContests(db); err != nil {
		return err
	}
	return nil
}
