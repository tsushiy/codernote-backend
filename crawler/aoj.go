package main

const (
	aojCoursesURL = "https://judgeapi.u-aizu.ac.jp/courses"
	aojProblemsURL = "https://judgeapi.u-aizu.ac.jp/problems?page=0&size=20000"
)

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