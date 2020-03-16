package crawler

const (
	codeforcesProblemsURL = "https://codeforces.com/api/problemset.problems"
	codeforcesContestsURL = "https://codeforces.com/api/contest.list?gym=false"
)

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
