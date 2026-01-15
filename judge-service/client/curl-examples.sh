#!/bin/bash

# Judge Service API テスト用 curl コマンド集
# 使用前に JUDGE_URL を設定してください

JUDGE_URL="${JUDGE_URL:-http://localhost:8081}"

echo "=== Health Check ==="
curl -s "${JUDGE_URL}/api/health" | jq .

echo ""
echo "=== Hello World (AC expected) ==="
curl -s -X POST "${JUDGE_URL}/api/judge" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "#include <stdio.h>\nint main() { printf(\"Hello\\n\"); return 0; }",
    "language": "c",
    "testCases": [
      {"input": "", "expected": "Hello\n"}
    ],
    "timeLimit": 2000,
    "memoryLimit": 256
  }' | jq .

echo ""
echo "=== Sum Program (AC expected) ==="
curl -s -X POST "${JUDGE_URL}/api/judge" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "#include <stdio.h>\nint main() { int a,b; scanf(\"%d %d\", &a, &b); printf(\"%d\\n\", a+b); return 0; }",
    "language": "c",
    "testCases": [
      {"input": "1 2\n", "expected": "3\n"},
      {"input": "10 20\n", "expected": "30\n"},
      {"input": "-5 10\n", "expected": "5\n"}
    ],
    "timeLimit": 2000,
    "memoryLimit": 256
  }' | jq .

echo ""
echo "=== Wrong Answer (WA expected) ==="
curl -s -X POST "${JUDGE_URL}/api/judge" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "#include <stdio.h>\nint main() { printf(\"Wrong\\n\"); return 0; }",
    "language": "c",
    "testCases": [
      {"input": "", "expected": "Hello\n"}
    ],
    "timeLimit": 2000,
    "memoryLimit": 256
  }' | jq .

echo ""
echo "=== Compile Error (CE expected) ==="
curl -s -X POST "${JUDGE_URL}/api/judge" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "#include <stdioo.h>\nint main() { return 0; }",
    "language": "c",
    "testCases": [
      {"input": "", "expected": ""}
    ],
    "timeLimit": 2000,
    "memoryLimit": 256
  }' | jq .

echo ""
echo "=== Time Limit Exceeded (TLE expected) ==="
curl -s -X POST "${JUDGE_URL}/api/judge" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "#include <stdio.h>\nint main() { while(1); return 0; }",
    "language": "c",
    "testCases": [
      {"input": "", "expected": ""}
    ],
    "timeLimit": 1000,
    "memoryLimit": 256
  }' | jq .

echo ""
echo "=== Runtime Error (RE expected) ==="
curl -s -X POST "${JUDGE_URL}/api/judge" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "#include <stdio.h>\nint main() { int *p = 0; *p = 1; return 0; }",
    "language": "c",
    "testCases": [
      {"input": "", "expected": ""}
    ],
    "timeLimit": 2000,
    "memoryLimit": 256
  }' | jq .
