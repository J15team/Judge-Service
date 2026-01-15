# Judge Service 設計仕様

## 概要

Judge Service は、提出されたコードをサンドボックス環境で実行し、テストケースに対する判定結果を返す独立マイクロサービス。

## AWS アーキテクチャ

```
Users → API Gateway → Main Backend (ECS) → Judge Service (ECS) → Sandbox (isolate)
                              ↓
                      RDS PostgreSQL
```

### デプロイ構成

| コンポーネント | AWS サービス | 詳細                                        |
| -------------- | ------------ | ------------------------------------------- |
| Judge Service  | ECS EC2      | 既存クラスタ (j15-backend-cluster-dev)      |
| EC2 Instance   | t3.small     | ecs-judge-service-dev (i-090a993fda9fe3fb4) |
| コンテナ       | ECR          | judge-service-dev リポジトリ                |
| ネットワーク   | VPC          | Public Subnet (10.0.1.0/24)                 |
| サービス間通信 | 内部 HTTP    | Main Backend → Judge Service (Port 8081)    |
| ログ           | CloudWatch   | /ecs/judge-service-dev                      |

### ECS タスク定義

```json
{
  "family": "judge-service-dev",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["EC2"],
  "cpu": "1024",
  "memory": "2048",
  "containerDefinitions": [
    {
      "name": "judge-service",
      "image": "127214181395.dkr.ecr.ap-northeast-1.amazonaws.com/judge-service-dev:latest",
      "portMappings": [
        {
          "containerPort": 8081,
          "protocol": "tcp"
        }
      ],
      "privileged": true,
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/judge-service-dev",
          "awslogs-region": "ap-northeast-1",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ]
}
```

**注意:** Docker-in-Docker を使用するため `privileged: true` が必要。EC2 起動タイプを使用。

## 技術スタック

| コンポーネント | 技術選定         | 理由                       |
| -------------- | ---------------- | -------------------------- |
| 言語           | Go               | 軽量、高速、並行処理に強い |
| フレームワーク | Gin or Echo      | シンプルで高性能           |
| サンドボックス | Docker-in-Docker | セキュアな隔離環境         |
| コンパイラ     | GCC              | C 言語対応                 |

## 内部アーキテクチャ

```
┌─────────────────────────────────────────────────────────────┐
│                   Judge Service (ECS)                        │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐   │
│  │   API Layer  │  │   Compiler   │  │  Result Comparator│   │
│  │  (HTTP/JSON) │  │   (GCC)      │  │                  │   │
│  └──────┬───────┘  └──────┬───────┘  └────────┬─────────┘   │
│         │                 │                    │             │
│         ▼                 ▼                    ▼             │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │                    Executor (Sandbox)                   │ │
│  │  ┌─────────────────────────────────────────────────┐    │ │
│  │  │                   isolate                       │    │ │
│  │  │  - CPU制限                                      │    │ │
│  │  │  - メモリ制限                                   │    │ │
│  │  │  - 時間制限                                     │    │ │
│  │  │  - ネットワーク隔離                             │    │ │
│  │  │  - ファイルシステム制限                         │    │ │
│  │  └─────────────────────────────────────────────────┘    │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## API 仕様

### POST /api/judge

コードを判定する。

**Request:**

```json
{
  "code": "#include <stdio.h>\nint main() { printf(\"Hello\\n\"); return 0; }",
  "language": "c",
  "testCases": [
    { "input": "", "expected": "Hello\n" },
    { "input": "5", "expected": "25\n" }
  ],
  "timeLimit": 2000,
  "memoryLimit": 256
}
```

**Response (成功):**

```json
{
  "results": [
    {
      "index": 0,
      "verdict": "AC",
      "executionTime": 15,
      "memoryUsed": 1024,
      "actualOutput": "Hello\n"
    },
    {
      "index": 1,
      "verdict": "WA",
      "executionTime": 12,
      "memoryUsed": 1024,
      "actualOutput": "10\n"
    }
  ]
}
```

**Response (コンパイルエラー):**

```json
{
  "results": [
    {
      "index": 0,
      "verdict": "CE",
      "errorMessage": "main.c:1:10: fatal error: stdioo.h: No such file or directory"
    }
  ]
}
```

### GET /api/health

ヘルスチェック。

**Response:**

```json
{
  "status": "healthy",
  "supportedLanguages": ["c"]
}
```

## Verdict（判定結果）

| Verdict | 説明                                 |
| ------- | ------------------------------------ |
| AC      | Accepted - 正解                      |
| WA      | Wrong Answer - 不正解                |
| TLE     | Time Limit Exceeded - 時間超過       |
| MLE     | Memory Limit Exceeded - メモリ超過   |
| RE      | Runtime Error - 実行時エラー         |
| CE      | Compilation Error - コンパイルエラー |

## サンドボックス要件

### セキュリティ

1. **ネットワーク隔離**: 外部通信を完全に遮断
2. **ファイルシステム制限**: 読み書き可能な領域を最小限に
3. **プロセス制限**: fork bomb 対策
4. **システムコール制限**: seccomp で危険な syscall をブロック

### リソース制限

| リソース       | デフォルト | 最大  |
| -------------- | ---------- | ----- |
| CPU 時間       | 2 秒       | 10 秒 |
| メモリ         | 256MB      | 512MB |
| プロセス数     | 1          | 10    |
| ファイルサイズ | 10MB       | 50MB  |

## 実行フロー

```
1. リクエスト受信 (Main Backend から)
   ↓
2. コードをファイルに書き出し
   ↓
3. コンパイル（C言語の場合: gcc -O2 -o main main.c）
   ├─ 失敗 → CE を返す
   ↓
4. 各テストケースに対して:
   a. isolate でサンドボックス作成
   b. 入力をstdinに渡して実行
   c. 時間・メモリを監視
   d. 出力を取得
   e. 期待値と比較
   ↓
5. 結果を集約して返す
```

## ディレクトリ構成

```
judge-service/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── handler.go
│   │   └── router.go
│   ├── compiler/
│   │   ├── compiler.go
│   │   └── gcc.go
│   ├── executor/
│   │   ├── executor.go
│   │   └── isolate.go
│   ├── comparator/
│   │   └── comparator.go
│   └── config/
│       └── config.go
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

## 設定

### 環境変数

| 変数名               | デフォルト | 説明                      |
| -------------------- | ---------- | ------------------------- |
| PORT                 | 8081       | サーバーポート            |
| DEFAULT_TIME_LIMIT   | 2000       | デフォルト時間制限 (ms)   |
| DEFAULT_MEMORY_LIMIT | 256        | デフォルトメモリ制限 (MB) |
| MAX_TIME_LIMIT       | 10000      | 最大時間制限 (ms)         |
| MAX_MEMORY_LIMIT     | 512        | 最大メモリ制限 (MB)       |

### config.yaml

```yaml
server:
  port: 8081
  timeout: 30s

sandbox:
  defaultTimeLimit: 2000
  defaultMemoryLimit: 256
  maxTimeLimit: 10000
  maxMemoryLimit: 512

compiler:
  c:
    command: "gcc"
    args: ["-O2", "-o", "main", "main.c"]
    timeout: 10s
```

## Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o judge-service ./cmd/server

FROM ubuntu:22.04

# isolate と GCC をインストール
RUN apt-get update && apt-get install -y \
    gcc \
    libc6-dev \
    libcap-dev \
    git \
    make \
    && rm -rf /var/lib/apt/lists/*

# isolate をビルド
RUN git clone https://github.com/ioi/isolate.git /tmp/isolate \
    && cd /tmp/isolate \
    && make install \
    && rm -rf /tmp/isolate

COPY --from=builder /app/judge-service /usr/local/bin/

EXPOSE 8081

CMD ["judge-service"]
```

## デプロイ手順

### 1. ECR リポジトリ作成

```bash
aws ecr create-repository \
  --repository-name judge-service-dev \
  --region ap-northeast-1
```

### 2. Docker イメージビルド & プッシュ

```bash
# ログイン
aws ecr get-login-password --region ap-northeast-1 | \
  docker login --username AWS --password-stdin 127214181395.dkr.ecr.ap-northeast-1.amazonaws.com

# ビルド & プッシュ
docker build -t judge-service-dev .
docker tag judge-service-dev:latest 127214181395.dkr.ecr.ap-northeast-1.amazonaws.com/judge-service-dev:latest
docker push 127214181395.dkr.ecr.ap-northeast-1.amazonaws.com/judge-service-dev:latest
```

### 3. ECS サービス作成

```bash
# タスク定義登録
aws ecs register-task-definition --cli-input-json file://task-definition.json

# サービス作成
aws ecs create-service \
  --cluster j15-backend-cluster-dev \
  --service-name judge-service-dev \
  --task-definition judge-service-dev \
  --desired-count 1 \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[subnet-0b92bd0152e3cae6a],securityGroups=[sg-0bcda1559b9fdde75],assignPublicIp=ENABLED}"
```

### 4. セキュリティグループ更新

Main Backend から Judge Service への通信を許可:

```bash
aws ec2 authorize-security-group-ingress \
  --group-id sg-0bcda1559b9fdde75 \
  --protocol tcp \
  --port 8081 \
  --source-group sg-0bcda1559b9fdde75
```

## メインバックエンドとの連携

```
Main Backend (ECS)                    Judge Service (ECS)
     │                                      │
     │  POST /api/judge                     │
     │  {code, language, testCases, ...}    │
     │ ─────────────────────────────────────▶│
     │                                      │
     │                                      │ コンパイル
     │                                      │ 実行 (isolate)
     │                                      │ 比較
     │                                      │
     │  {results: [{verdict, time, ...}]}   │
     │ ◀─────────────────────────────────────│
     │                                      │
     │  結果をDBに保存                       │
     │  スコア計算                           │
     ▼                                      ▼
```

### Main Backend 設定

`application.yml`:

```yaml
judge-service:
  base-url: http://judge-service.local:8081 # Cloud Map 使用時
  # または ECS Service Discovery / 固定IP
  timeout: 30000
```

## 今後の拡張

1. **言語追加**: Python, Java, C++, Rust など
2. **スペシャルジャッジ**: カスタム比較関数対応
3. **インタラクティブ問題**: 双方向通信対応
4. **並列実行**: 複数テストケースの同時実行
5. **キャッシュ**: 同一コードの再判定をスキップ
6. **Auto Scaling**: 負荷に応じた自動スケーリング

## 参考

- [isolate](https://github.com/ioi/isolate) - IOI 公式サンドボックス
- [Judge0](https://github.com/judge0/judge0) - オープンソースオンラインジャッジ
- [AtCoder](https://atcoder.jp/) - 競技プログラミングサイト
