package crawler

import "time"

const (
	yukicoderProblemsURL = "https://yukicoder.me/api/v1/problems"
	yukicoderContestsURL = "https://yukicoder.me/api/v1/contest/past"
)

type yukicoderProblem struct {
	No          int       `json:"No"`
	ProblemID   int       `json:"ProblemId"`
	Title       string    `json:"Title"`
	AuthorID    int       `json:"AuthorId"`
	TesterID    int       `json:"TesterId"`
	Level       int       `json:"Level"`
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
