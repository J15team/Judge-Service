package entity

// Submission represents a code submission to be judged
type Submission struct {
	Code        string     `json:"code"`
	Language    string     `json:"language"`
	TestCases   []TestCase `json:"testCases"`
	TimeLimit   int        `json:"timeLimit"`   // milliseconds
	MemoryLimit int        `json:"memoryLimit"` // MB
}

// JudgeResponse represents the response from the judge service
type JudgeResponse struct {
	Results []JudgeResult `json:"results"`
}
