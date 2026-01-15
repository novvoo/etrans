package translator

import (
	"fmt"
	"path/filepath"
	"strings"
)

// DocumentType 文档类型
type DocumentType string

const (
	DocumentTypeEPUB DocumentType = "epub"
	DocumentTypePDF  DocumentType = "pdf"
)

// Document 文档接口
type Document interface {
	GetTextBlocks() []string
	InsertTranslation(translations map[string]string) error
	InsertMonolingualTranslation(translations map[string]string) error
	Save(outputPath string) error
}

// OpenDocument 打开文档
func OpenDocument(filePath string) (Document, DocumentType, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".epub":
		doc, err := OpenEPUB(filePath)
		if err != nil {
			return nil, "", fmt.Errorf("打开 EPUB 文件失败: %w", err)
		}
		return doc, DocumentTypeEPUB, nil
	case ".pdf":
		// PDF支持暂时不实现
		return nil, "", fmt.Errorf("PDF支持尚未实现")
	default:
		return nil, "", fmt.Errorf("不支持的文件格式: %s，仅支持 .epub 文件", ext)
	}
}

// ValidateDocument 验证文档
func ValidateDocument(filePath string) error {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".epub":
		return ValidateEPUB(filePath)
	case ".pdf":
		return fmt.Errorf("PDF支持尚未实现")
	default:
		return fmt.Errorf("不支持的文件格式: %s，仅支持 .epub 文件", ext)
	}
}

// GetTranslationMode 获取翻译模式
func GetTranslationMode(ext string) string {
	switch ext {
	case ".epub":
		// EPUB使用基础翻译
		return TranslationModeBasic
	case ".pdf":
		return TranslationModeBasic
	default:
		return TranslationModeBasic
	}
}

// GetDocumentInfo 获取文档信息
func GetDocumentInfo(filePath string) (map[string]interface{}, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	info := make(map[string]interface{})

	switch ext {
	case ".epub":
		epub, err := OpenEPUB(filePath)
		if err != nil {
			return nil, err
		}
		info["type"] = "EPUB"
		info["title"] = epub.Metadata.Title
		info["author"] = epub.Metadata.Author
		info["language"] = epub.Metadata.Language
		info["textBlocks"] = len(epub.GetTextBlocks())
	case ".pdf":
		return nil, fmt.Errorf("PDF支持尚未实现")
	default:
		return nil, fmt.Errorf("不支持的文件格式: %s", ext)
	}

	return info, nil
}
