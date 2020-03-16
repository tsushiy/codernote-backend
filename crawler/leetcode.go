package crawler

const (
	leetcodeProblemsURL = "https://leetcode.com/api/problems/all/"
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
