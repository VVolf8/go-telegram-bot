package cache

import (
	"errors"
	"sync"

	"gotgbot/core"
)

// Cache – интерфейс кэширования.
type Cache interface {
	Set(key string, value interface{}) error
	Get(key string) (interface{}, error)
	Delete(key string) error
}

// MemoryCache – простая реализация кэша в памяти.
type MemoryCache struct {
	data   map[string]interface{}
	mu     sync.RWMutex
	logger core.Logger
}

// NewMemoryCache создаёт новый in-memory кэш с использованием переданного логгера.
func NewMemoryCache(logger core.Logger) *MemoryCache {
	return &MemoryCache{
		data:   make(map[string]interface{}),
		logger: logger,
	}
}

// Set устанавливает значение для заданного ключа.
func (mc *MemoryCache) Set(key string, value interface{}) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.data[key] = value
	mc.logger.Info("Cache set", core.Field{"key", key})
	return nil
}

// Get возвращает значение по ключу. Если ключ отсутствует, возвращает ошибку.
func (mc *MemoryCache) Get(key string) (interface{}, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	val, exists := mc.data[key]
	if !exists {
		mc.logger.Warn("Cache miss", core.Field{"key", key})
		return nil, errors.New("key not found")
	}
	mc.logger.Info("Cache hit", core.Field{"key", key})
	return val, nil
}

// Delete удаляет значение по ключу. Если ключ не найден, возвращает ошибку.
func (mc *MemoryCache) Delete(key string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	if _, exists := mc.data[key]; !exists {
		mc.logger.Warn("Cache delete: key not found", core.Field{"key", key})
		return errors.New("key not found")
	}
	delete(mc.data, key)
	mc.logger.Info("Cache deleted", core.Field{"key", key})
	return nil
}
