package benchmark

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const helloWorldC = `#include <stdio.h>
int main() {
    printf("Hello, World!\n");
    return 0;
}`

const sumC = `#include <stdio.h>
int main() {
    int a, b;
    scanf("%d %d", &a, &b);
    printf("%d\n", a + b);
    return 0;
}`

func BenchmarkCompileHelloWorld(b *testing.B) {
	workDir, _ := os.MkdirTemp("", "bench-*")
	defer os.RemoveAll(workDir)

	srcPath := filepath.Join(workDir, "main.c")
	os.WriteFile(srcPath, []byte(helloWorldC), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outPath := filepath.Join(workDir, fmt.Sprintf("main_%d", i))
		cmd := exec.Command("gcc", "-O2", "-o", outPath, srcPath)
		cmd.Run()
	}
}

func BenchmarkExecuteHelloWorld(b *testing.B) {
	workDir, _ := os.MkdirTemp("", "bench-*")
	defer os.RemoveAll(workDir)

	srcPath := filepath.Join(workDir, "main.c")
	outPath := filepath.Join(workDir, "main")
	os.WriteFile(srcPath, []byte(helloWorldC), 0644)
	exec.Command("gcc", "-O2", "-o", outPath, srcPath).Run()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := exec.Command(outPath)
		cmd.Run()
	}
}

func BenchmarkCompileAndExecute(b *testing.B) {
	for i := 0; i < b.N; i++ {
		workDir, _ := os.MkdirTemp("", "bench-*")
		srcPath := filepath.Join(workDir, "main.c")
		outPath := filepath.Join(workDir, "main")
		
		os.WriteFile(srcPath, []byte(sumC), 0644)
		exec.Command("gcc", "-O2", "-o", outPath, srcPath).Run()
		
		cmd := exec.Command(outPath)
		cmd.Stdin = strings.NewReader("3 5\n")
		cmd.Run()
		
		os.RemoveAll(workDir)
	}
}

func BenchmarkFullJudge5TestCases(b *testing.B) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"1 2\n", "3\n"},
		{"10 20\n", "30\n"},
		{"100 200\n", "300\n"},
		{"-5 10\n", "5\n"},
		{"0 0\n", "0\n"},
	}

	for i := 0; i < b.N; i++ {
		workDir, _ := os.MkdirTemp("", "bench-*")
		srcPath := filepath.Join(workDir, "main.c")
		outPath := filepath.Join(workDir, "main")
		
		os.WriteFile(srcPath, []byte(sumC), 0644)
		exec.Command("gcc", "-O2", "-o", outPath, srcPath).Run()
		
		for _, tc := range testCases {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			cmd := exec.CommandContext(ctx, outPath)
			cmd.Stdin = strings.NewReader(tc.input)
			var stdout bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Run()
			cancel()
		}
		
		os.RemoveAll(workDir)
	}
}

func TestMeasureLatency(t *testing.T) {
	workDir, _ := os.MkdirTemp("", "bench-*")
	defer os.RemoveAll(workDir)

	srcPath := filepath.Join(workDir, "main.c")
	outPath := filepath.Join(workDir, "main")
	os.WriteFile(srcPath, []byte(sumC), 0644)

	// Measure compile time
	start := time.Now()
	exec.Command("gcc", "-O2", "-o", outPath, srcPath).Run()
	compileTime := time.Since(start)

	// Measure execution time (5 test cases)
	testInputs := []string{"1 2\n", "10 20\n", "100 200\n", "-5 10\n", "0 0\n"}
	start = time.Now()
	for _, input := range testInputs {
		cmd := exec.Command(outPath)
		cmd.Stdin = strings.NewReader(input)
		var stdout bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Run()
	}
	execTime := time.Since(start)

	t.Logf("Compile time: %v", compileTime)
	t.Logf("Execute time (5 cases sequential): %v", execTime)
	t.Logf("Total: %v", compileTime+execTime)
}
