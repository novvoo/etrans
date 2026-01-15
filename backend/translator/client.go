package translator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// TranslatorClient 翻译客户端
type TranslatorClient struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
	Timeout time.Duration
}

// NewTranslatorClient 创建新的翻译客户端
func NewTranslatorClient(apiKey, baseURL string) *TranslatorClient {
	return &TranslatorClient{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
		Timeout: 30 * time.Second,
	}
}

// Translate 翻译文本
func (c *TranslatorClient) Translate(text, targetLanguage, userPrompt string) (string, error) {
	// 这里应该实现具体的翻译逻辑
	// 简化实现，返回原文本
	return text, fmt.Errorf("翻译客户端未实现具体逻辑")
}

// OpenAI-compatible API Request/Response structures
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
}

type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Choices []struct {
		Index        int     `json:"index"`
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
	Error struct {
		Message string      `json:"message"`
		Type    string      `json:"type"`
		Code    interface{} `json:"code"`
	} `json:"error"`
}

// LLMClient LLM翻译客户端
type LLMClient struct {
	APIKey  string
	BaseURL string
	Model   string
	Client  *http.Client
	Timeout time.Duration
}

// NewLLMClient 创建新的LLM客户端
func NewLLMClient(apiKey, baseURL, model string) *LLMClient {
	return &LLMClient{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   model,
		Client: &http.Client{
			Timeout: 60 * time.Second,
		},
		Timeout: 60 * time.Second,
	}
}

// Translate 使用LLM翻译文本
func (c *LLMClient) Translate(text, targetLanguage, userPrompt string) (string, error) {
	// 构造系统提示词
	systemPrompt := "You are a professional translator. Translate the following text into " + targetLanguage + "."
	if userPrompt != "" {
		systemPrompt += "\n" + userPrompt
	} else {
		systemPrompt += "\nOnly return the translated text without any explanations or extra quotes."
	}

	// 构造请求体
	reqBody := ChatCompletionRequest{
		Model: c.Model,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: text},
		},
		Temperature: 0.3,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request failed: %w", err)
	}

	// 构造API地址
	// 如果 BaseURL 已经包含 http，直接使用；否则假设是 OpenAI 格式，追加路径
	// 这里做一个简单的处理：如果用户给的是完整的 /chat/completions 结尾，就直接用
	// 否则追加 /v1/chat/completions (如果是 openai) 或者 /chat/completions (通用)
	// 简单起见，我们优先使用用户提供的 URL，如果它看起来像个 Base URL，我们追加标准路径
	apiURL := c.BaseURL
	if !strings.HasSuffix(apiURL, "/chat/completions") {
		if strings.HasSuffix(apiURL, "/") {
			apiURL += "chat/completions"
		} else {
			apiURL += "/chat/completions"
		}
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	// 发送请求
	resp, err := c.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// 解析响应
	var completionResp ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&completionResp); err != nil {
		return "", fmt.Errorf("decode response failed: %w", err)
	}

	if completionResp.Error.Message != "" {
		return "", fmt.Errorf("API error: %s", completionResp.Error.Message)
	}

	if len(completionResp.Choices) == 0 {
		return "", fmt.Errorf("no translation returned")
	}

	return strings.TrimSpace(completionResp.Choices[0].Message.Content), nil
}
