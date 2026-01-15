package translator

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Cache 翻译缓存
type Cache struct {
	data map[string]CacheEntry
	dir  string // 缓存目录
	mu   sync.RWMutex
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Value      string    `json:"value"`
	ExpireTime time.Time `json:"expire_time"`
}

// NewCache 创建新的缓存
func NewCache(dir string) (*Cache, error) {
	// 创建缓存目录
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建缓存目录失败: %w", err)
	}

	cache := &Cache{
		data: make(map[string]CacheEntry),
		dir:  dir,
	}
	// 启动清理协程
	go cache.cleanup()
	return cache, nil
}

// DisableCache 禁用缓存读取
func (c *Cache) DisableCache() {
	// 这里可以添加禁用逻辑，比如设置一个标志
	// 目前简化处理
}

// Get 获取缓存值
func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	// 先查内存
	if entry, exists := c.data[key]; exists {
		c.mu.RUnlock()
		if time.Now().After(entry.ExpireTime) {
			return "", false
		}
		return entry.Value, true
	}
	c.mu.RUnlock()

	// 查磁盘
	entry, exists := c.loadFromDisk(key)
	if !exists {
		return "", false
	}

	if time.Now().After(entry.ExpireTime) {
		return "", false
	}

	// 回填内存
	c.mu.Lock()
	c.data[key] = entry
	c.mu.Unlock()

	return entry.Value, true
}

// Set 设置缓存值
func (c *Cache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := CacheEntry{
		Value:      value,
		ExpireTime: time.Now().Add(24 * 7 * time.Hour), // 延长到7天过期，避免太频繁失效
	}

	c.data[key] = entry

	// 异步写入磁盘，避免阻塞
	go c.saveToDisk(key, entry)
}

// loadFromDisk 从磁盘读取缓存
func (c *Cache) loadFromDisk(key string) (CacheEntry, bool) {
	filePath := filepath.Join(c.dir, key)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return CacheEntry{}, false
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return CacheEntry{}, false
	}

	return entry, true
}

// saveToDisk 保存缓存到磁盘
func (c *Cache) saveToDisk(key string, entry CacheEntry) {
	filePath := filepath.Join(c.dir, key)
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	_ = os.WriteFile(filePath, data, 0644)
}

// cleanup 清理过期缓存
func (c *Cache) cleanup() {
	ticker := time.NewTicker(12 * time.Hour) // 降低清理频率
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		// 清理内存
		for key, entry := range c.data {
			if now.After(entry.ExpireTime) {
				delete(c.data, key)
				// 也可以尝试删除磁盘文件，但为了性能可能暂时忽略，或者单独做磁盘清理
			}
		}
		c.mu.Unlock()
	}
}

// CacheKey 生成缓存键
func CacheKey(text, targetLanguage, userPrompt string) string {
	data := fmt.Sprintf("%s|%s|%s", text, targetLanguage, userPrompt)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// SaveProgressMap 保存任务进度映射
func (c *Cache) SaveProgressMap(taskID string, data map[string]string) error {
	filePath := filepath.Join(c.dir, fmt.Sprintf("progress_%s.json", taskID))
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化进度数据失败: %w", err)
	}
	return os.WriteFile(filePath, jsonData, 0644)
}

// LoadProgressMap 加载任务进度映射
func (c *Cache) LoadProgressMap(taskID string) (map[string]string, error) {
	filePath := filepath.Join(c.dir, fmt.Sprintf("progress_%s.json", taskID))
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // 文件不存在，不是错误
		}
		return nil, fmt.Errorf("读取进度文件失败: %w", err)
	}

	var progress map[string]string
	if err := json.Unmarshal(data, &progress); err != nil {
		return nil, fmt.Errorf("解析进度数据失败: %w", err)
	}
	return progress, nil
}
