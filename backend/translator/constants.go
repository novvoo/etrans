package translator

import "errors"

// ProviderType 翻译提供商类型
type ProviderType string

const (
	// ProviderOpenAI OpenAI
	ProviderOpenAI ProviderType = "openai"
	// ProviderClaude Claude
	ProviderClaude ProviderType = "claude"
	// ProviderGemini Gemini
	ProviderGemini ProviderType = "gemini"
	// ProviderDeepSeek DeepSeek
	ProviderDeepSeek ProviderType = "deepseek"
	// ProviderOllama Ollama
	ProviderOllama ProviderType = "ollama"
	// ProviderNLTranslator NLTranslator (MacOS Native)
	ProviderNLTranslator ProviderType = "nltranslator"
	// ProviderCustom Custom
	ProviderCustom ProviderType = "custom"
)

const (
	// TranslationModeBasic 基础翻译模式
	TranslationModeBasic = "basic"
)

// ProviderConfig 提供商配置
type ProviderConfig struct {
	Type        ProviderType           `json:"type"`
	APIKey      string                 `json:"api_key"`
	APIURL      string                 `json:"api_url"`
	Model       string                 `json:"model"`
	Temperature float64                `json:"temperature"`
	MaxTokens   int                    `json:"max_tokens"`
	Extra       map[string]interface{} `json:"extra"`
}

// ErrPaused 任务已暂停
var ErrPaused = errors.New("task paused")
