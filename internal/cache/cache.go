package cache

import (
	"bytes"
	"encoding/gob"
	"github.com/coocood/freecache"
	"github.com/gotd/td/tg"
	"go-winx-api/internal/models"
	"go.uber.org/zap"
	"sync"
)

var cache *Cache

type Cache struct {
	cache *freecache.Cache
	mu    sync.RWMutex
	log   *zap.Logger
}

func InitCache(log *zap.Logger) {
	log = log.Named("cache")

	gob.Register(models.File{})
	gob.Register(tg.InputDocumentFileLocation{})
	defer log.Sugar().Info("initialized")

	cache = &Cache{cache: freecache.NewCache(1024 * 1024 * 1024), log: log} // 1GB
}

func GetCache() *Cache {
	return cache
}

func (c *Cache) GetFile(key string, value *models.File) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	data, err := cache.cache.Get([]byte(key))
	if err != nil {
		return err
	}
	dec := gob.NewDecoder(bytes.NewReader(data))
	err = dec.Decode(&value)
	if err != nil {
		return err
	}
	return nil
}

func (c *Cache) SetFile(key string, value *models.File, expireSeconds int) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(value)
	if err != nil {
		return err
	}
	cache.cache.Set([]byte(key), buf.Bytes(), expireSeconds)
	return nil
}

func (c *Cache) GetPost(key string, value *models.Post) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	data, err := cache.cache.Get([]byte(key))
	if err != nil {
		return err
	}
	dec := gob.NewDecoder(bytes.NewReader(data))
	err = dec.Decode(&value)
	if err != nil {
		return err
	}
	return nil
}

func (c *Cache) SetPost(key string, value *models.Post, expireSeconds int) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(value)
	if err != nil {
		return err
	}
	cache.cache.Set([]byte(key), buf.Bytes(), expireSeconds)
	return nil
}

func (c *Cache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	cache.cache.Del([]byte(key))
	return nil
}
