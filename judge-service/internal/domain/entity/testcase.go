package entity

// TestCase represents a single test case for code execution
type TestCase struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
}

// JudgeResult represents the result of a single test case execution
type JudgeResult struct {
	Index         int     `json:"index"`
	Verdict       Verdict `json:"verdict"`
	ExecutionTime int64   `json:"executionTime,omitempty"` // milliseconds
	MemoryUsed    int64   `json:"memoryUsed,omitempty"`    // KB
	ActualOutput  string  `json:"actualOutput,omitempty"`
	ErrorMessage  string  `json:"errorMessage,omitempty"`
}
