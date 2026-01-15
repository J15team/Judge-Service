package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"judge-service/internal/domain/entity"
)

const sandboxImage = "judge-sandbox:latest"

// DockerPoolSandbox uses pre-warmed Docker containers for fast secure execution
type DockerPoolSandbox struct {
	compileTimeout time.Duration
	workPool       chan string
	poolSize       int
}

// NewDockerPoolSandbox creates a new DockerPoolSandbox
func NewDockerPoolSandbox(compileTimeout time.Duration, poolSize int) *DockerPoolSandbox {
	s := &DockerPoolSandbox{
		compileTimeout: compileTimeout,
		workPool:       make(chan string, poolSize),
		poolSize:       poolSize,
	}
	for i := 0; i < poolSize; i++ {
		dir, _ := os.MkdirTemp("", fmt.Sprintf("judge-pool-%d-*", i))
		os.Chmod(dir, 0777)
		s.workPool <- dir
	}
	return s
}

// CompileAndRun compiles and runs code in Docker sandbox
func (s *DockerPoolSandbox) CompileAndRun(code, language string, testCases []entity.TestCase, timeLimit, memoryLimit int) []entity.JudgeResult {
	results := make([]entity.JudgeResult, len(testCases))

	workDir := <-s.workPool
	defer func() { s.workPool <- workDir }()

	s.cleanWorkDir(workDir)

	srcFile := "main.c"
	if err := os.WriteFile(filepath.Join(workDir, srcFile), []byte(code), 0644); err != nil {
		return s.allError(results, entity.VerdictRE, "failed to write source file")
	}
	os.Chmod(filepath.Join(workDir, srcFile), 0644)

	if compileErr := s.compile(workDir, language); compileErr != "" {
		return s.allError(results, entity.VerdictCE, compileErr)
	}

	// Run test cases concurrently
	var wg sync.WaitGroup
	for i, tc := range testCases {
		wg.Add(1)
		go func(idx int, testCase entity.TestCase) {
			defer wg.Done()
			results[idx] = s.runTestCase(idx, workDir, testCase, timeLimit, memoryLimit)
		}(i, tc)
	}
	wg.Wait()

	return results
}

func (s *DockerPoolSandbox) cleanWorkDir(dir string) {
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		os.RemoveAll(filepath.Join(dir, e.Name()))
	}
}

func (s *DockerPoolSandbox) compile(workDir, language string) string {
	if language != "c" {
		return fmt.Sprintf("unsupported language: %s", language)
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.compileTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "run",
		"--rm",
		"--network", "none",           // No network
		"--memory", "512m",            // Memory limit
		"--memory-swap", "512m",       // No swap
		"--cpus", "1",                 // CPU limit
		"--pids-limit", "50",          // Process limit
		"--ulimit", "fsize=10485760",  // 10MB file size
		"--ulimit", "nofile=64",       // Open files
		"--read-only",                 // Read-only root
		"--tmpfs", "/tmp:rw,noexec,nosuid,size=64m",
		"--security-opt", "no-new-privileges",
		"-v", workDir+":/work:rw",
		"-w", "/work",
		"-u", "1000:1000",             // Non-root user
		sandboxImage,
		"gcc", "-O2", "-static", "-o", "main", "main.c",
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "compilation timeout"
		}
		errMsg := stderr.String()
		if errMsg == "" {
			errMsg = err.Error()
		}
		return errMsg
	}

	return ""
}

func (s *DockerPoolSandbox) runTestCase(index int, workDir string, tc entity.TestCase, timeLimit, memoryLimit int) entity.JudgeResult {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeLimit*3)*time.Millisecond)
	defer cancel()

	startTime := time.Now()

	cmd := exec.CommandContext(ctx, "docker", "run",
		"--rm",
		"-i",
		"--network", "none",
		"--memory", strconv.Itoa(memoryLimit)+"m",
		"--memory-swap", strconv.Itoa(memoryLimit)+"m",
		"--cpus", "1",
		"--pids-limit", "5",
		"--ulimit", "fsize=1048576",   // 1MB output
		"--ulimit", "nofile=16",
		"--read-only",
		"--security-opt", "no-new-privileges",
		"-v", workDir+":/work:ro",     // Read-only mount
		"-w", "/work",
		"-u", "1000:1000",
		sandboxImage,
		"timeout", fmt.Sprintf("%.1f", float64(timeLimit)/1000.0), "./main",
	)

	cmd.Stdin = strings.NewReader(tc.Input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	execTime := time.Since(startTime).Milliseconds()

	result := entity.JudgeResult{
		Index:         index,
		ExecutionTime: execTime,
		ActualOutput:  stdout.String(),
	}

	if ctx.Err() == context.DeadlineExceeded {
		result.Verdict = entity.VerdictTLE
		return result
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code := exitErr.ExitCode()
			if code == 124 { // timeout
				result.Verdict = entity.VerdictTLE
				return result
			}
			if code == 137 { // OOM
				result.Verdict = entity.VerdictMLE
				return result
			}
		}
		result.Verdict = entity.VerdictRE
		result.ErrorMessage = stderr.String()
		return result
	}

	if normalizeOutput(tc.Expected) == normalizeOutput(stdout.String()) {
		result.Verdict = entity.VerdictAC
	} else {
		result.Verdict = entity.VerdictWA
	}

	return result
}

func (s *DockerPoolSandbox) allError(results []entity.JudgeResult, verdict entity.Verdict, msg string) []entity.JudgeResult {
	for i := range results {
		results[i] = entity.JudgeResult{Index: i, Verdict: verdict, ErrorMessage: msg}
	}
	return results
}
