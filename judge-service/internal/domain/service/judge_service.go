package service

import "judge-service/internal/domain/entity"

// Compiler compiles source code
type Compiler interface {
	Compile(code, language, workDir string) error
}

// Executor executes compiled code in sandbox
type Executor interface {
	Execute(workDir, input string, timeLimit, memoryLimit int) (*ExecutionResult, error)
}

// ExecutionResult holds the result of code execution
type ExecutionResult struct {
	Output        string
	ExecutionTime int64 // milliseconds
	MemoryUsed    int64 // KB
	ExitCode      int
	TimedOut      bool
	MemoryExceeded bool
}

// Comparator compares expected and actual output
type Comparator interface {
	Compare(expected, actual string) bool
}

// JudgeService orchestrates the judging process
type JudgeService struct {
	compiler   Compiler
	executor   Executor
	comparator Comparator
}

// NewJudgeService creates a new JudgeService
func NewJudgeService(compiler Compiler, executor Executor, comparator Comparator) *JudgeService {
	return &JudgeService{
		compiler:   compiler,
		executor:   executor,
		comparator: comparator,
	}
}

// Judge executes and judges a submission
func (s *JudgeService) Judge(submission *entity.Submission, workDir string) *entity.JudgeResponse {
	response := &entity.JudgeResponse{
		Results: make([]entity.JudgeResult, len(submission.TestCases)),
	}

	// Compile
	if err := s.compiler.Compile(submission.Code, submission.Language, workDir); err != nil {
		for i := range submission.TestCases {
			response.Results[i] = entity.JudgeResult{
				Index:        i,
				Verdict:      entity.VerdictCE,
				ErrorMessage: err.Error(),
			}
		}
		return response
	}

	// Execute each test case
	for i, tc := range submission.TestCases {
		result := s.executeTestCase(i, tc, workDir, submission.TimeLimit, submission.MemoryLimit)
		response.Results[i] = result
	}

	return response
}

func (s *JudgeService) executeTestCase(index int, tc entity.TestCase, workDir string, timeLimit, memoryLimit int) entity.JudgeResult {
	execResult, err := s.executor.Execute(workDir, tc.Input, timeLimit, memoryLimit)
	if err != nil {
		return entity.JudgeResult{
			Index:        index,
			Verdict:      entity.VerdictRE,
			ErrorMessage: err.Error(),
		}
	}

	if execResult.TimedOut {
		return entity.JudgeResult{
			Index:         index,
			Verdict:       entity.VerdictTLE,
			ExecutionTime: execResult.ExecutionTime,
		}
	}

	if execResult.MemoryExceeded {
		return entity.JudgeResult{
			Index:      index,
			Verdict:    entity.VerdictMLE,
			MemoryUsed: execResult.MemoryUsed,
		}
	}

	if execResult.ExitCode != 0 {
		return entity.JudgeResult{
			Index:        index,
			Verdict:      entity.VerdictRE,
			ActualOutput: execResult.Output,
		}
	}

	if s.comparator.Compare(tc.Expected, execResult.Output) {
		return entity.JudgeResult{
			Index:         index,
			Verdict:       entity.VerdictAC,
			ExecutionTime: execResult.ExecutionTime,
			MemoryUsed:    execResult.MemoryUsed,
			ActualOutput:  execResult.Output,
		}
	}

	return entity.JudgeResult{
		Index:         index,
		Verdict:       entity.VerdictWA,
		ExecutionTime: execResult.ExecutionTime,
		MemoryUsed:    execResult.MemoryUsed,
		ActualOutput:  execResult.Output,
	}
}
