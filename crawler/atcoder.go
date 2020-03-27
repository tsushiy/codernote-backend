package crawler

import (
	"encoding/json"
	"log"
	"sort"

	"github.com/jinzhu/gorm"
	. "github.com/tsushiy/codernote-backend/db"
)

const (
	atcoderDomain            = "atcoder"
	atcoderProblemsURL       = "https://kenkoooo.com/atcoder/resources/problems.json"
	atcoderContestsURL       = "https://kenkoooo.com/atcoder/resources/contests.json"
	atcoderContestProblemURL = "https://kenkoooo.com/atcoder/resources/contest-problem.json"
)

var atcoderContestProblemMap = make(map[string][]Problem)

type atcoderProblem struct {
	ProblemID string `json:"id"`
	ContestID string `json:"contest_id"`
	Title     string `json:"title"`
}

type atcoderContest struct {
	ContestID        string `json:"id"`
	StartEpochSecond int    `json:"start_epoch_second"`
	DurationSecond   int    `json:"duration_second"`
	Title            string `json:"title"`
	RateChange       string `json:"rate_change"`
}

type atcoderContestProblem struct {
	ContestID string `json:"contest_id"`
	ProblemID string `json:"problem_id"`
}

func updateAtcoderProblems(db *gorm.DB) error {
	log.Println("Start updating AtCoder problem info")
	body, err := fetchAPI(atcoderProblemsURL)
	if err != nil {
		return err
	}
	var problems []atcoderProblem
	if err := json.Unmarshal(body, &problems); err != nil {
		return err
	}

	for _, v := range problems {
		var problem Problem
		if err := db.
			Where(Problem{
				Domain:    atcoderDomain,
				ProblemID: v.ProblemID,
			}).
			Assign(Problem{
				Domain:    atcoderDomain,
				ProblemID: v.ProblemID,
				ContestID: v.ContestID,
				Title:     v.Title,
			}).
			FirstOrCreate(&problem).Error; err != nil {
			return err
		}
		atcoderContestProblemMap[v.ContestID] = append(atcoderContestProblemMap[v.ContestID], problem)
	}

	return nil
}

func fetchAtcoderContestProblem(db *gorm.DB) error {
	log.Println("Start fetching AtCoder contest-problem pair")
	body, err := fetchAPI(atcoderContestProblemURL)
	if err != nil {
		return err
	}
	var pairs []atcoderContestProblem
	if err := json.Unmarshal(body, &pairs); err != nil {
		return err
	}
	for _, v := range pairs {
		var problem Problem
		if err := db.
			Where(Problem{
				Domain:    atcoderDomain,
				ProblemID: v.ProblemID,
			}).
			Take(&problem).Error; err != nil {
			return err
		}
		if v.ContestID != problem.ContestID {
			atcoderContestProblemMap[v.ContestID] = append(atcoderContestProblemMap[v.ContestID], problem)
		}
	}
	for _, v := range atcoderContestProblemMap {
		sort.Slice(v, func(i, j int) bool { return v[i].ProblemID < v[j].ProblemID })
	}
	return nil
}

func updateAtcoderContests(db *gorm.DB) error {
	log.Println("Start updating AtCoder contest info")
	body, err := fetchAPI(atcoderContestsURL)
	if err != nil {
		return err
	}
	var contests []atcoderContest
	if err := json.Unmarshal(body, &contests); err != nil {
		return err
	}

	for _, v := range contests {
		var problemNoList []int64
		for _, problem := range atcoderContestProblemMap[v.ContestID] {
			problemNoList = append(problemNoList, int64(problem.No))
		}
		if len(problemNoList) == 0 {
			continue
		}
		if err := db.
			Where(Contest{
				Domain:    atcoderDomain,
				ContestID: v.ContestID,
			}).
			Assign(Contest{
				Domain:           atcoderDomain,
				ContestID:        v.ContestID,
				Title:            v.Title,
				StartTimeSeconds: v.StartEpochSecond,
				DurationSeconds:  v.DurationSecond,
				Rated:            v.RateChange,
				ProblemNoList:    problemNoList,
			}).
			FirstOrCreate(&Contest{}).Error; err != nil {
			return err
		}
	}

	return nil
}

func updateAtcoder(db *gorm.DB) error {
	if err := updateAtcoderProblems(db); err != nil {
		return err
	}
	if err := fetchAtcoderContestProblem(db); err != nil {
		return err
	}
	if err := updateAtcoderContests(db); err != nil {
		return err
	}
	return nil
}
