package translator

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
)

// DocumentTranslator 文档翻译器
type DocumentTranslator struct {
	Client TranslatorClientInterface
	Cache  *Cache
}

// TranslatorClientInterface 翻译客户端接口
type TranslatorClientInterface interface {
	Translate(text, targetLanguage, userPrompt string) (string, error)
}

// NewDocumentTranslator 创建新的文档翻译器
func NewDocumentTranslator(config ProviderConfig, cache *Cache) (*DocumentTranslator, error) {
	// 创建对应的翻译客户端
	var client TranslatorClientInterface

	// 根据提供商类型创建客户端
	// 这里简化实现，使用基础的翻译客户端
	// 实际应该根据 config.Type 创建不同的客户端
	switch config.Type {
	case ProviderOpenAI, ProviderClaude, ProviderGemini, ProviderDeepSeek, ProviderOllama, ProviderNLTranslator, ProviderCustom:
		// 使用通用的LLM客户端
		client = NewLLMClient(config.APIKey, config.APIURL, config.Model)
	default:
		return nil, fmt.Errorf("不支持的翻译提供商: %s", config.Type)
	}

	return &DocumentTranslator{
		Client: client,
		Cache:  cache,
	}, nil
}

// TranslateDocument 翻译文档（统一入口）
func (dt *DocumentTranslator) TranslateDocument(taskID string, inputPath, outputPath, targetLanguage, userPrompt string, forceRetranslate bool, generateMode string, progressCallback func(float64), checkStatus func() string) (string, error) {
	// 根据文件扩展名判断文档类型
	ext := strings.ToLower(filepath.Ext(inputPath))

	switch ext {
	case ".epub":
		return dt.TranslateEPUB(taskID, inputPath, outputPath, targetLanguage, userPrompt, generateMode, progressCallback, checkStatus)
	case ".pdf":
		return "", fmt.Errorf("PDF翻译支持尚未实现")
	default:
		return "", fmt.Errorf("不支持的文件格式: %s", ext)
	}
}

// TranslateEPUB 翻译EPUB文档
func (dt *DocumentTranslator) TranslateEPUB(taskID, inputPath, outputPath, targetLanguage, userPrompt string, generateMode string, progressCallback func(float64), checkStatus func() string) (string, error) {
	log.Printf("开始翻译EPUB: %s", inputPath)

	// 打开EPUB文档
	doc, _, err := OpenDocument(inputPath)
	if err != nil {
		return "", fmt.Errorf("打开EPUB文档失败: %w", err)
	}

	// 获取文本块
	textBlocks := doc.GetTextBlocks()
	if len(textBlocks) == 0 {
		return "", fmt.Errorf("EPUB中没有可翻译的文本内容")
	}

	log.Printf("找到 %d 个文本块", len(textBlocks))

	// 批量翻译
	translations := make(map[string]string)

	// 尝试加载之前的进度
	if savedProgress, err := dt.Cache.LoadProgressMap(taskID); err == nil && savedProgress != nil {
		log.Printf("加载已保存的进度: %d 个条目", len(savedProgress))
		translations = savedProgress
	}

	for i, block := range textBlocks {
		// 检查任务状态
		if checkStatus != nil {
			status := checkStatus()
			if status == "paused" {
				// 保存当前进度
				if err := dt.Cache.SaveProgressMap(taskID, translations); err != nil {
					log.Printf("保存进度失败: %v", err)
				}
				return "", ErrPaused
			}
		}

		if block == "" {
			continue
		}

		// 如果已经翻译过（从进度加载或刚刚翻译），跳过
		if _, ok := translations[block]; ok {
			continue
		}

		// 检查缓存
		cacheKey := CacheKey(block, targetLanguage, userPrompt)
		if cached, ok := dt.Cache.Get(cacheKey); ok {
			translations[block] = cached
			continue
		}

		// 翻译文本
		translated, err := dt.Client.Translate(block, targetLanguage, userPrompt)
		if err != nil {
			log.Printf("翻译文本失败: %s, 错误: %v", block, err)
			// 失败时不保存到 translations map，也不写入缓存
			// 下次重试时会重新翻译
			continue
		}

		translations[block] = translated

		// 保存到缓存
		dt.Cache.Set(cacheKey, translated)

		// 更新进度
		if progressCallback != nil {
			progress := float64(i+1) / float64(len(textBlocks))
			progressCallback(progress)
		}
	}

	// 插入翻译到EPUB
	if generateMode == "monolingual" {
		if err := doc.InsertMonolingualTranslation(translations); err != nil {
			return "", fmt.Errorf("插入单语翻译失败: %w", err)
		}
	} else {
		if err := doc.InsertTranslation(translations); err != nil {
			return "", fmt.Errorf("插入双语翻译失败: %w", err)
		}
	}

	// 翻译元数据
	if epub, ok := doc.(*EPUBFile); ok {
		if err := TranslateMetadata(epub, dt.Client, targetLanguage, userPrompt, dt.Cache); err != nil {
			log.Printf("翻译元数据失败: %v", err)
		}
	}

	// 翻译目录
	if epub, ok := doc.(*EPUBFile); ok {
		tocItems, err := ParseTOC(epub)
		if err == nil && len(tocItems) > 0 {
			if err := TranslateTOC(tocItems, dt.Client, targetLanguage, userPrompt, dt.Cache); err != nil {
				log.Printf("翻译目录失败: %v", err)
			} else {
				WriteTOC(epub, tocItems)
			}
		}
	}

	// 保存EPUB文档
	if err := doc.Save(outputPath); err != nil {
		return "", fmt.Errorf("保存EPUB文档失败: %w", err)
	}

	log.Printf("EPUB翻译完成: %s", outputPath)
	return outputPath, nil
}

// TranslateText 翻译文本
func (dt *DocumentTranslator) TranslateText(text, targetLanguage, userPrompt string) (string, error) {
	// 检查缓存
	cacheKey := CacheKey(text, targetLanguage, userPrompt)
	if cached, ok := dt.Cache.Get(cacheKey); ok {
		return cached, nil
	}

	// 翻译文本
	translated, err := dt.Client.Translate(text, targetLanguage, userPrompt)
	if err != nil {
		return "", fmt.Errorf("翻译文本失败: %w", err)
	}

	// 保存到缓存
	dt.Cache.Set(cacheKey, translated)

	return translated, nil
}

// BatchTranslate 批量翻译
func (dt *DocumentTranslator) BatchTranslate(texts []string, targetLanguage, userPrompt string) ([]string, error) {
	results := make([]string, len(texts))

	for i, text := range texts {
		if text == "" {
			results[i] = ""
			continue
		}

		translated, err := dt.TranslateText(text, targetLanguage, userPrompt)
		if err != nil {
			log.Printf("批量翻译第 %d 个文本失败: %v", i, err)
			results[i] = text // 保留原文
			continue
		}

		results[i] = translated
	}

	return results, nil
}

// ValidateInput 验证输入参数
func ValidateInput(inputPath, outputPath, targetLanguage string) error {
	if inputPath == "" {
		return fmt.Errorf("输入文件路径不能为空")
	}

	if outputPath == "" {
		return fmt.Errorf("输出文件路径不能为空")
	}

	if targetLanguage == "" {
		return fmt.Errorf("目标语言不能为空")
	}

	// 验证文件格式
	if !strings.HasSuffix(strings.ToLower(inputPath), ".epub") {
		return fmt.Errorf("仅支持EPUB文件格式")
	}

	return nil
}
