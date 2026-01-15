package entity

// Verdict represents the result of code execution
type Verdict string

const (
	VerdictAC  Verdict = "AC"  // Accepted
	VerdictWA  Verdict = "WA"  // Wrong Answer
	VerdictTLE Verdict = "TLE" // Time Limit Exceeded
	VerdictMLE Verdict = "MLE" // Memory Limit Exceeded
	VerdictRE  Verdict = "RE"  // Runtime Error
	VerdictCE  Verdict = "CE"  // Compilation Error
)
