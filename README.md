# etrans - EPUB翻译工具

etrans是从etrans项目中分离出来的专门用于EPUB文档翻译的独立工具。

## 项目结构

```
etrans/
├── backend/
│   ├── go.mod              # Go模块定义
│   ├── test.go             # 测试和示例代码
│   └── translator/         # 翻译核心模块
│       ├── cache.go        # 翻译缓存
│       ├── client.go       # 翻译客户端接口
│       ├── constants.go    # 常量定义
│       ├── document.go     # 文档接口
│       ├── epub.go         # EPUB文件处理
│       ├── metadata.go     # 元数据翻译
│       ├── toc.go          # 目录翻译
│       └── translator.go   # 主翻译器
└── README.md               # 项目说明
```

## 功能特性

### 核心功能
- **EPUB文件解析**: 支持打开和解析EPUB文件
- **文本提取**: 智能提取HTML/XHTML中的文本内容
- **批量翻译**: 支持批量翻译文本块
- **双语显示**: 支持原文+译文的双语显示模式
- **单语翻译**: 支持替换原文的单语翻译模式
- **元数据翻译**: 翻译书籍标题、作者等信息
- **目录翻译**: 翻译书籍目录结构
- **翻译缓存**: 避免重复翻译，提高效率

### 技术特性
- **XML解析**: 使用XML解析器处理HTML内容
- **备用方案**: 当XML解析失败时使用正则表达式备用方案
- **进度回调**: 支持翻译进度回调
- **错误处理**: 完善的错误处理和日志记录

## 快速开始

### 准备工作

确保您的系统已安装以下环境：
- [Go](https://go.dev/dl/) 1.24+
- [Node.js](https://nodejs.org/) (用于前端构建)

### 开发环境

在开发模式下，后端会自动代理前端请求，无需单独运行前端服务。

```bash
# Windows (使用批处理脚本)
.\start.bat

# Linux/macOS
make dev
```

启动后访问: http://localhost:8080

### 编译构建

构建生产版本（包含前端静态资源和后端可执行文件）：

```bash
# Windows
go run build.go

# Linux/macOS
make build
```

构建完成后，运行生成的 `etrans` 或 `etrans.exe` 即可启动服务。

### Docker 部署

使用 Docker 快速部署服务：

```bash
# 构建镜像
make docker-build
# 或
docker build -t etrans .

# 启动服务
make docker-run
# 或
docker-compose up -d
```

## 使用方法

### 基本用法

```go
package main

import (
    "log"
    "os"
    
    "./translator"
)

func main() {
    // 创建翻译客户端
    client := &YourTranslatorClient{}
    
    // 创建翻译器
    translator := translator.NewDocumentTranslator(client)
    
    // 翻译EPUB文件
    result, err := translator.TranslateEPUB(
        "input.epub",           // 输入文件
        "output.epub",          // 输出文件
        "zh-CN",               // 目标语言
        "",                    // 用户提示
        "bilingual",           // 翻译模式：bilingual或monolingual
        func(progress float64) {
            log.Printf("进度: %.1f%%", progress*100)
        },
    )
    
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("翻译完成: %s", result)
}
```

### 翻译模式

1. **bilingual（双语模式）**: 在原文后添加翻译，保持双语显示
2. **monolingual（单语模式）**: 用翻译替换原文

### 支持的翻译客户端

etrans使用接口设计，支持多种翻译客户端：

```go
type TranslatorClientInterface interface {
    Translate(text, targetLanguage, userPrompt string) (string, error)
}
```

## API文档

### 主要结构体

#### EPUBFile
表示EPUB文件的结构体，包含文件路径、内容和元数据。

#### DocumentTranslator
主要的翻译器类，负责协调整个翻译流程。

#### Cache
翻译缓存，用于避免重复翻译。

### 主要方法

#### OpenEPUB(path string) (*EPUBFile, error)
打开并解析EPUB文件。

#### TranslateEPUB(inputPath, outputPath, targetLanguage, userPrompt string, generateMode string, progressCallback func(float64)) (string, error)
翻译EPUB文档的主要方法。

#### ExtractTextBlocks(html string) []string
从HTML内容中提取文本块。

#### InsertTranslation(html string, translations map[string]string) string
在HTML中插入翻译（双语模式）。

#### InsertMonolingualTranslation(html string, translations map[string]string) string
在HTML中插入翻译（单语模式）。

## 依赖关系

- Go 1.24.1+
- github.com/google/uuid v1.5.0

## 开发说明

### 添加新的翻译客户端

1. 实现`TranslatorClientInterface`接口
2. 在`client.go`中添加新的客户端类型
3. 在主程序中创建客户端实例

### 扩展文档格式支持

1. 在`document.go`中添加新的文档类型
2. 实现`Document`接口
3. 更新`OpenDocument`函数

## 许可证

本项目继承原etrans项目的许可证。

## 贡献

欢迎提交Issue和Pull Request来改进这个项目。
