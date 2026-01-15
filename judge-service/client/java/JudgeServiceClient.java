package com.example.client;

import com.fasterxml.jackson.databind.ObjectMapper;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.util.List;

/**
 * Judge Service Client for Main Backend
 * 
 * Usage:
 *   JudgeServiceClient client = new JudgeServiceClient("http://judge-service:8081");
 *   JudgeResponse response = client.judge(code, "c", testCases, 2000, 256);
 */
public class JudgeServiceClient {
    
    private final String baseUrl;
    private final HttpClient httpClient;
    private final ObjectMapper objectMapper;
    
    public JudgeServiceClient(String baseUrl) {
        this.baseUrl = baseUrl;
        this.httpClient = HttpClient.newBuilder()
            .connectTimeout(Duration.ofSeconds(5))
            .build();
        this.objectMapper = new ObjectMapper();
    }
    
    public JudgeResponse judge(String code, String language, List<TestCase> testCases, 
                                int timeLimit, int memoryLimit) throws Exception {
        JudgeRequest request = new JudgeRequest(code, language, testCases, timeLimit, memoryLimit);
        String json = objectMapper.writeValueAsString(request);
        
        HttpRequest httpRequest = HttpRequest.newBuilder()
            .uri(URI.create(baseUrl + "/api/judge"))
            .header("Content-Type", "application/json")
            .timeout(Duration.ofSeconds(30))
            .POST(HttpRequest.BodyPublishers.ofString(json))
            .build();
        
        HttpResponse<String> response = httpClient.send(httpRequest, 
            HttpResponse.BodyHandlers.ofString());
        
        if (response.statusCode() != 200) {
            throw new RuntimeException("Judge service error: " + response.body());
        }
        
        return objectMapper.readValue(response.body(), JudgeResponse.class);
    }
    
    public boolean healthCheck() {
        try {
            HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(baseUrl + "/api/health"))
                .GET()
                .timeout(Duration.ofSeconds(5))
                .build();
            
            HttpResponse<String> response = httpClient.send(request, 
                HttpResponse.BodyHandlers.ofString());
            return response.statusCode() == 200;
        } catch (Exception e) {
            return false;
        }
    }
    
    // DTOs
    public record TestCase(String input, String expected) {}
    public record JudgeRequest(String code, String language, List<TestCase> testCases, 
                               int timeLimit, int memoryLimit) {}
    public record JudgeResult(int index, String verdict, long executionTime, 
                              long memoryUsed, String actualOutput, String errorMessage) {}
    public record JudgeResponse(List<JudgeResult> results) {}
}
