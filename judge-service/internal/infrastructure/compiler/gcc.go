package compiler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// GCCCompiler compiles C code using GCC
type GCCCompiler struct {
	Timeout time.Duration
}

// NewGCCCompiler creates a new GCCCompiler
func NewGCCCompiler(timeout time.Duration) *GCCCompiler {
	return &GCCCompiler{Timeout: timeout}
}

// Compile compiles C source code
func (c *GCCCompiler) Compile(code, language, workDir string) error {
	if language != "c" {
		return fmt.Errorf("unsupported language: %s", language)
	}

	srcPath := filepath.Join(workDir, "main.c")
	outPath := filepath.Join(workDir, "main")

	if err := os.WriteFile(srcPath, []byte(code), 0644); err != nil {
		return fmt.Errorf("failed to write source file: %w", err)
	}

	cmd := exec.Command("gcc", "-O2", "-o", outPath, srcPath)
	cmd.Dir = workDir

	output, err := runWithTimeout(cmd, c.Timeout)
	if err != nil {
		return fmt.Errorf("%s", string(output))
	}

	return nil
}

func runWithTimeout(cmd *exec.Cmd, timeout time.Duration) ([]byte, error) {
	done := make(chan error, 1)
	var output []byte
	var cmdErr error

	go func() {
		output, cmdErr = cmd.CombinedOutput()
		done <- cmdErr
	}()

	select {
	case <-time.After(timeout):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return nil, fmt.Errorf("compilation timeout")
	case <-done:
		return output, cmdErr
	}
}
