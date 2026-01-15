# Spring Boot での Judge Service 連携

## 1. application.yml 設定

```yaml
judge-service:
  base-url: http://judge-service.local:8081
  timeout: 30000
```

## 2. Configuration

```java
@Configuration
@ConfigurationProperties(prefix = "judge-service")
public class JudgeServiceConfig {
    private String baseUrl;
    private int timeout = 30000;

    // getters, setters
}
```

## 3. Service 実装

```java
@Service
@RequiredArgsConstructor
public class JudgeService {

    private final RestTemplate restTemplate;
    private final JudgeServiceConfig config;

    public JudgeResponse judge(Submission submission) {
        JudgeRequest request = JudgeRequest.builder()
            .code(submission.getCode())
            .language(submission.getLanguage())
            .testCases(submission.getProblem().getTestCases())
            .timeLimit(submission.getProblem().getTimeLimit())
            .memoryLimit(submission.getProblem().getMemoryLimit())
            .build();

        ResponseEntity<JudgeResponse> response = restTemplate.postForEntity(
            config.getBaseUrl() + "/api/judge",
            request,
            JudgeResponse.class
        );

        return response.getBody();
    }
}
```

## 4. ECS Service Discovery 設定

Main Backend と Judge Service が同じ VPC 内にある場合、
プライベート IP で直接通信可能です。

```yaml
# ECS Service Discovery を使う場合
judge-service:
  base-url: http://judge-service.local:8081

# 固定 IP を使う場合（タスクの Private IP）
judge-service:
  base-url: http://10.0.1.xxx:8081
```
