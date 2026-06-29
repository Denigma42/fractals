package repository

import (
    "sync"
    "github.com/hashicorp/golang-lru"
)

type TileCache interface {
    Get(key string) ([]uint16, bool)
    Add(key string, data []uint16)
}

type LRUCache struct {
    cache *lru.Cache
    mu    sync.RWMutex
}

func NewLRUCache(size int) (*LRUCache, error) {
    c, err := lru.New(size)
    if err != nil {
        return nil, err
    }
    return &LRUCache{cache: c}, nil
}

func (c *LRUCache) Get(key string) ([]uint16, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    val, ok := c.cache.Get(key)
    if !ok {
        return nil, false
    }
    // Возвращаем копию, чтобы избежать мутаций
    data := val.([]uint16)
    copyData := make([]uint16, len(data))
    copy(copyData, data)
    return copyData, true
}

func (c *LRUCache) Add(key string, data []uint16) {
    c.mu.Lock()
    defer c.mu.Unlock()
    // Сохраняем копию
    copyData := make([]uint16, len(data))
    copy(copyData, data)
    c.cache.Add(key, copyData)
}