package crawler

import (
	"encoding/json"
	"log"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	. "github.com/tsushiy/codernote-backend/db"
)

const (
	codeforcesDomain      = "codeforces"
	codeforcesProblemsURL = "https://codeforces.com/api/problemset.problems"
	codeforcesContestsURL = "https://codeforces.com/api/contest.list?gym=false"
)

var codeforcesProblems codeforcesProblem
var codeforcesContestProblemMap = make(map[string][]Problem)

type codeforcesProblem struct {
	Status string `json:"status"`
	Result struct {
		Problems []struct {
			ContestID int      `json:"contestId"`
			Index     string   `json:"index"`
			Name      string   `json:"name"`
			Type      string   `json:"type"`
			Tags      []string `json:"tags"`
			Points    float64  `json:"points,omitempty"`
			Rating    int      `json:"rating,omitempty"`
		} `json:"problems"`
		ProblemStatistics []struct {
			ContestID   int    `json:"contestId"`
			Index       string `json:"index"`
			SolvedCount int    `json:"solvedCount"`
		} `json:"problemStatistics"`
	} `json:"result"`
}

type codeforcesContest struct {
	Status string `json:"status"`
	Result []struct {
		ID                  int    `json:"id"`
		Name                string `json:"name"`
		Type                string `json:"type"`
		Phase               string `json:"phase"`
		Frozen              bool   `json:"frozen"`
		DurationSeconds     int    `json:"durationSeconds"`
		StartTimeSeconds    int    `json:"startTimeSeconds"`
		RelativeTimeSeconds int    `json:"relativeTimeSeconds"`
	} `json:"result"`
}

type ProblemFronContest struct {
	Status string `json:"status"`
	Result struct {
		Contest struct {
			ID                  int    `json:"id"`
			Name                string `json:"name"`
			Type                string `json:"type"`
			Phase               string `json:"phase"`
			Frozen              bool   `json:"frozen"`
			DurationSeconds     int    `json:"durationSeconds"`
			StartTimeSeconds    int    `json:"startTimeSeconds"`
			RelativeTimeSeconds int    `json:"relativeTimeSeconds"`
		} `json:"contest"`
		Problems []struct {
			ContestID int      `json:"contestId"`
			Index     string   `json:"index"`
			Name      string   `json:"name"`
			Type      string   `json:"type"`
			Points    float64  `json:"points"`
			Rating    int      `json:"rating"`
			Tags      []string `json:"tags"`
		} `json:"problems"`
		Rows []struct {
			Party struct {
				ContestID int `json:"contestId"`
				Members   []struct {
					Handle string `json:"handle"`
				} `json:"members"`
				ParticipantType  string `json:"participantType"`
				Ghost            bool   `json:"ghost"`
				Room             int    `json:"room"`
				StartTimeSeconds int    `json:"startTimeSeconds"`
			} `json:"party"`
			Rank                  int     `json:"rank"`
			Points                float64 `json:"points"`
			Penalty               int     `json:"penalty"`
			SuccessfulHackCount   int     `json:"successfulHackCount"`
			UnsuccessfulHackCount int     `json:"unsuccessfulHackCount"`
			ProblemResults        []struct {
				Points                    float64 `json:"points"`
				RejectedAttemptCount      int     `json:"rejectedAttemptCount"`
				Type                      string  `json:"type"`
				BestSubmissionTimeSeconds int     `json:"bestSubmissionTimeSeconds"`
			} `json:"problemResults"`
		} `json:"rows"`
	} `json:"result"`
}

type codeforcesProblemsType []struct {
	ContestID int      `json:"contestId"`
	Index     string   `json:"index"`
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	Points    float64  `json:"points"`
	Rating    int      `json:"rating"`
	Tags      []string `json:"tags"`
}

type codeforcesContestType struct {
	ID                  int    `json:"id"`
	Name                string `json:"name"`
	Type                string `json:"type"`
	Phase               string `json:"phase"`
	Frozen              bool   `json:"frozen"`
	DurationSeconds     int    `json:"durationSeconds"`
	StartTimeSeconds    int    `json:"startTimeSeconds"`
	RelativeTimeSeconds int    `json:"relativeTimeSeconds"`
}

func updateCodeforcesProblems(db *gorm.DB) error {
	log.Println("Start updating codeforces problem info")
	body, err := fetchAPI(codeforcesProblemsURL)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, &codeforcesProblems); err != nil {
		return err
	}
	problems := codeforcesProblems

	for _, v := range problems.Result.Problems {
		var problem Problem
		contestID := strconv.Itoa(v.ContestID)
		if err := db.
			Where(Problem{
				Domain:    codeforcesDomain,
				ProblemID: v.Index,
				ContestID: contestID,
			}).
			Assign(Problem{
				Domain:    codeforcesDomain,
				ProblemID: v.Index,
				ContestID: contestID,
				Title:     v.Name,
			}).
			FirstOrCreate(&problem).Error; err != nil {
			return err
		}
		codeforcesContestProblemMap[contestID] = append(codeforcesContestProblemMap[contestID], problem)
	}

	for _, v := range codeforcesContestProblemMap {
		sort.Slice(v, func(i, j int) bool { return v[i].ProblemID < v[j].ProblemID })
	}

	return nil
}

func updateCodeforcesContests(db *gorm.DB) error {
	log.Println("Start updating codeforces contest info")
	body, err := fetchAPI(codeforcesContestsURL)
	if err != nil {
		return err
	}
	var ret codeforcesContest
	if err := json.Unmarshal(body, &ret); err != nil {
		return err
	}
	contests := ret.Result

	var contestMap = make(map[string]codeforcesContestType)
	for _, v := range contests {
		contestMap[strconv.Itoa(v.ID)] = v
	}

	count := 0
	for _, v := range contests {
		if v.Phase != "FINISHED" {
			continue
		}
		contestID := strconv.Itoa(v.ID)
		var problemNoList []int64
		for _, problem := range codeforcesContestProblemMap[contestID] {
			problemNoList = append(problemNoList, int64(problem.No))
		}
		count++
		if count%5 == 0 {
			time.Sleep(1 * time.Second)
		}
		problems, err := getProblemsFromContest(contestID)
		if err != nil {
			continue
		}
		if len(problems) != len(problemNoList) {
			var newProblemNoList []int64
			for _, p1 := range problems {
				for _, p2 := range codeforcesProblems.Result.Problems {
					c1 := contestMap[strconv.Itoa(p1.ContestID)]
					c2 := contestMap[strconv.Itoa(p2.ContestID)]
					if p1.Name == p2.Name && c1.StartTimeSeconds == c2.StartTimeSeconds {
						var problem Problem
						if err := db.
							Where(Problem{
								Domain:    codeforcesDomain,
								ProblemID: p2.Index,
								ContestID: strconv.Itoa(p2.ContestID),
								Title:     p2.Name,
							}).
							First(&problem).Error; err != nil {
							return err
						}
						newProblemNoList = append(newProblemNoList, int64(problem.No))
					}
				}
			}
			problemNoList = newProblemNoList[:]
		}
		rated := "-"
		if isDiv1(v.Name) && isDiv2(v.Name) {
			rated = "12"
		} else if isDiv1(v.Name) {
			rated = "1"
		} else if isDiv2(v.Name) {
			rated = "2"
		} else if isDiv3(v.Name) {
			rated = "3"
		}
		if err := db.
			Where(Contest{
				Domain:    codeforcesDomain,
				ContestID: contestID,
			}).
			Assign(Contest{
				Domain:           codeforcesDomain,
				ContestID:        contestID,
				Title:            v.Name,
				StartTimeSeconds: v.StartTimeSeconds,
				DurationSeconds:  v.DurationSeconds,
				Rated:            rated,
				ProblemNoList:    problemNoList,
			}).
			FirstOrCreate(&Contest{}).Error; err != nil {
			return err
		}
	}

	return nil
}

func getProblemsFromContest(contestID string) (codeforcesProblemsType, error) {
	url := "https://codeforces.com/api/contest.standings?contestId=" + contestID + "&from=1&count=1"
	var body []byte
	var err error
	for i := 0; i < 3; i++ {
		body, err = fetchAPI(url)
		if err != nil {
			log.Printf("Cannot fetch problems %d/3. contestID: %s", i+1, contestID)
			time.Sleep(1 * time.Second)
			continue
		}
	}
	if err != nil {
		return nil, err
	}
	var ret ProblemFronContest
	if err := json.Unmarshal(body, &ret); err != nil {
		return nil, err
	}

	problems := ret.Result.Problems
	sort.Slice(problems, func(i, j int) bool { return problems[i].Index < problems[j].Index })

	return problems, nil
}

func isDiv1(title string) bool {
	return regexp.MustCompile("Div.( ?)1").MatchString(title)
}

func isDiv2(title string) bool {
	return regexp.MustCompile("Div.( ?)2").MatchString(title)
}

func isDiv3(title string) bool {
	return regexp.MustCompile("Div.( ?)3").MatchString(title)
}

func updateCodeforces(db *gorm.DB) error {
	if err := updateCodeforcesProblems(db); err != nil {
		return err
	}
	if err := updateCodeforcesContests(db); err != nil {
		return err
	}
	return nil
}
