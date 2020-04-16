package crawler

import (
	"encoding/json"
	"log"
	"sort"
	"strconv"

	"github.com/jinzhu/gorm"
	. "github.com/tsushiy/codernote-backend/db"
)

const (
	atcoderDomain            = "atcoder"
	atcoderProblemsURL       = "https://kenkoooo.com/atcoder/resources/merged-problems.json"
	atcoderContestsURL       = "https://kenkoooo.com/atcoder/resources/contests.json"
	atcoderContestProblemURL = "https://kenkoooo.com/atcoder/resources/contest-problem.json"
	atcoderDifficultyURL     = "https://kenkoooo.com/atcoder/resources/problem-models.json"
)

var atcoderContestProblemMap = make(map[string][]Problem)

type atcoderProblem struct {
	ProblemID            string      `json:"id"`
	ContestID            string      `json:"contest_id"`
	Title                string      `json:"title"`
	ShortestSubmissionID int         `json:"shortest_submission_id"`
	ShortestProblemID    string      `json:"shortest_problem_id"`
	ShortestContestID    string      `json:"shortest_contest_id"`
	ShortestUserID       string      `json:"shortest_user_id"`
	FastestSubmissionID  int         `json:"fastest_submission_id"`
	FastestProblemID     string      `json:"fastest_problem_id"`
	FastestContestID     string      `json:"fastest_contest_id"`
	FastestUserID        string      `json:"fastest_user_id"`
	FirstSubmissionID    int         `json:"first_submission_id"`
	FirstProblemID       string      `json:"first_problem_id"`
	FirstContestID       string      `json:"first_contest_id"`
	FirstUserID          string      `json:"first_user_id"`
	SourceCodeLength     int         `json:"source_code_length"`
	ExecutionTime        int         `json:"execution_time"`
	Point                interface{} `json:"point"`
	Predict              float64     `json:"predict"`
	SolverCount          int         `json:"solver_count"`
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

type atcoderDifficulty struct {
	Slope            float64 `json:"slope"`
	Intercept        float64 `json:"intercept"`
	Variance         float64 `json:"variance"`
	Difficulty       float64 `json:"difficulty"`
	Discrimination   float64 `json:"discrimination"`
	IrtLoglikelihood float64 `json:"irt_loglikelihood"`
	IrtUsers         int     `json:"irt_users"`
	IsExperimental   bool    `json:"is_experimental"`
}

func updateAtcoderProblems(db *gorm.DB) error {
	log.Println("Start updating AtCoder problem info")
	var problems []atcoderProblem
	{
		body, err := fetchAPI(atcoderProblemsURL)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(body, &problems); err != nil {
			return err
		}
	}
	var difficulties map[string]atcoderDifficulty
	{
		body, err := fetchAPI(atcoderDifficultyURL)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(body, &difficulties); err != nil {
			return err
		}
	}

	for _, v := range problems {
		var problem Problem
		var difficulty float64
		if d, ok := difficulties[v.ProblemID]; ok {
			difficulty = d.Difficulty
		}
		if err := db.
			Where(Problem{
				Domain:    atcoderDomain,
				ProblemID: v.ProblemID,
			}).
			Assign(Problem{
				Domain:     atcoderDomain,
				ProblemID:  v.ProblemID,
				ContestID:  v.ContestID,
				Title:      v.Title,
				Difficulty: strconv.FormatFloat(difficulty, 'f', -1, 64),
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
