package executor

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"judge-service/internal/domain/service"
)

// IsolateExecutor executes code using isolate sandbox
type IsolateExecutor struct {
	BoxID int
}

// NewIsolateExecutor creates a new IsolateExecutor
func NewIsolateExecutor(boxID int) *IsolateExecutor {
	return &IsolateExecutor{BoxID: boxID}
}

// Execute runs the compiled binary in isolate sandbox
func (e *IsolateExecutor) Execute(workDir, input string, timeLimit, memoryLimit int) (*service.ExecutionResult, error) {
	// Initialize isolate box
	initCmd := exec.Command("isolate", "--box-id", strconv.Itoa(e.BoxID), "--init")
	boxPath, err := initCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to init isolate: %w", err)
	}
	boxDir := strings.TrimSpace(string(boxPath))
	defer exec.Command("isolate", "--box-id", strconv.Itoa(e.BoxID), "--cleanup").Run()

	// Copy executable to box
	srcExe := filepath.Join(workDir, "main")
	dstExe := filepath.Join(boxDir, "box", "main")
	if err := copyFile(srcExe, dstExe); err != nil {
		return nil, fmt.Errorf("failed to copy executable: %w", err)
	}
	os.Chmod(dstExe, 0755)

	// Write input file
	inputPath := filepath.Join(boxDir, "box", "input.txt")
	if err := os.WriteFile(inputPath, []byte(input), 0644); err != nil {
		return nil, fmt.Errorf("failed to write input: %w", err)
	}

	// Meta file for results
	metaPath := filepath.Join(workDir, "meta.txt")

	// Run in isolate
	timeLimitSec := float64(timeLimit) / 1000.0
	wallTimeSec := timeLimitSec * 2

	runCmd := exec.Command("isolate",
		"--box-id", strconv.Itoa(e.BoxID),
		"--time", fmt.Sprintf("%.1f", timeLimitSec),
		"--wall-time", fmt.Sprintf("%.1f", wallTimeSec),
		"--mem", strconv.Itoa(memoryLimit*1024), // KB
		"--processes=10",
		"--stdin=input.txt",
		"--meta", metaPath,
		"--run", "/box/main",
	)

	output, _ := runCmd.CombinedOutput()

	// Parse meta file
	result := &service.ExecutionResult{
		Output: string(output),
	}

	if meta, err := parseMeta(metaPath); err == nil {
		if t, ok := meta["time"]; ok {
			if f, err := strconv.ParseFloat(t, 64); err == nil {
				result.ExecutionTime = int64(f * 1000)
			}
		}
		if m, ok := meta["max-rss"]; ok {
			if i, err := strconv.ParseInt(m, 10, 64); err == nil {
				result.MemoryUsed = i
			}
		}
		if status, ok := meta["status"]; ok {
			switch status {
			case "TO":
				result.TimedOut = true
			case "SG", "RE":
				result.ExitCode = 1
			}
		}
		if exitcode, ok := meta["exitcode"]; ok {
			if i, err := strconv.Atoi(exitcode); err == nil {
				result.ExitCode = i
			}
		}
	}

	return result, nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0755)
}

func parseMeta(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	meta := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), ":", 2)
		if len(parts) == 2 {
			meta[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return meta, scanner.Err()
}
