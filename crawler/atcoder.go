package main

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

var contestProblemMap = make(map[string][]string)

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
		fetched := &Problem{
			Domain:    atcoderDomain,
			ProblemID: v.ProblemID,
			ContestID: v.ContestID,
			Title:     v.Title,
		}
		in := &Problem{
			Domain:    atcoderDomain,
			ProblemID: v.ProblemID,
		}

		if err := cmpAndUpdate(fetched, in, db); err != nil {
			return err
		}
	}

	return nil
}

func fetchAtcoderContestProblem() error {
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
		contestProblemMap[v.ContestID] = append(contestProblemMap[v.ContestID], v.ProblemID)
	}
	for _, v := range contestProblemMap {
		sort.Slice(v, func(i, j int) bool { return v[i] < v[j] })
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
		fetched := &Contest{
			Domain:           atcoderDomain,
			ContestID:        v.ContestID,
			Title:            v.Title,
			StartTimeSeconds: v.StartEpochSecond,
			DurationSeconds:  v.DurationSecond,
			ProblemIDList:    []string{},
		}
		in := &Contest{
			Domain:    atcoderDomain,
			ContestID: v.ContestID,
		}
		fetched.ProblemIDList = contestProblemMap[v.ContestID]

		if err := cmpAndUpdate(fetched, in, db); err != nil {
			return err
		}
	}

	return nil
}

func updateAtcoder(db *gorm.DB) error {
	if err := updateAtcoderProblems(db); err != nil {
		return err
	}
	if err := fetchAtcoderContestProblem(); err != nil {
		return err
	}
	if err := updateAtcoderContests(db); err != nil {
		return err
	}
	return nil
}
