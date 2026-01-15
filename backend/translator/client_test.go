package translator

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLLMClient_Translate(t *testing.T) {
	// 1. 创建 Mock 服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求方法
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// 验证请求头
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("Expected Authorization header 'Bearer test-api-key', got %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got %s", r.Header.Get("Content-Type"))
		}

		// 验证请求体
		var req ChatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		if req.Model != "gpt-3.5-turbo" {
			t.Errorf("Expected model 'gpt-3.5-turbo', got %s", req.Model)
		}
		if len(req.Messages) != 2 {
			t.Errorf("Expected 2 messages, got %d", len(req.Messages))
		}
		if req.Messages[1].Content != "Hello" {
			t.Errorf("Expected user message 'Hello', got %s", req.Messages[1].Content)
		}

		// 返回模拟响应
		resp := ChatCompletionResponse{
			Choices: []struct {
				Index        int     `json:"index"`
				Message      Message `json:"message"`
				FinishReason string  `json:"finish_reason"`
			}{
				{
					Message: Message{
						Role:    "assistant",
						Content: "你好",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// 2. 创建客户端
	client := NewLLMClient("test-api-key", server.URL, "gpt-3.5-turbo")

	// 3. 执行翻译
	result, err := client.Translate("Hello", "Chinese", "")
	if err != nil {
		t.Fatalf("Translate failed: %v", err)
	}

	// 4. 验证结果
	if result != "你好" {
		t.Errorf("Expected translation '你好', got '%s'", result)
	}
}

func TestLLMClient_Translate_Error(t *testing.T) {
	// 创建返回错误的 Mock 服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": {"message": "Invalid API Key"}}`))
	}))
	defer server.Close()

	client := NewLLMClient("invalid-key", server.URL, "gpt-3.5-turbo")

	_, err := client.Translate("Hello", "Chinese", "")
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "Invalid API Key") && !strings.Contains(err.Error(), "400") {
		t.Errorf("Expected error message containing 'Invalid API Key' or '400', got '%v'", err)
	}
}
