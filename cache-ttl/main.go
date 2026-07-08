package main

import (
	"fmt"
	"sync"
	"time"
)

type CacheItem struct {
	Value      any
	Expiration time.Time
}

type Cache struct {
	mu    sync.RWMutex
	items map[string]CacheItem
}

func (c *Cache) Set(key string, value any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	timeExpiration := time.Now().Add(ttl)
	cItem := CacheItem{
		Value:      value,
		Expiration: timeExpiration,
	}
	c.items[key] = cItem
}

func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cItem, ok := c.items[key]
	if !ok {
		return nil, false
	}

	if time.Now().After(cItem.Expiration) {
		return nil, false
	}

	return cItem.Value, true
}

func main() {
	cache := Cache{
		items: make(map[string]CacheItem),
	}

	fmt.Println("[Main] Сохраняем токен сессии в кэш на 1 секунду...")
	cache.Set("session_token", "SECRET_ABC_123", 1*time.Second)

	val, found := cache.Get("session_token")
	if found {
		fmt.Printf("Сразу после записи: данные найдены! Значение: %v\n", val)
	} else {
		fmt.Println("Сразу после записи: данные не найдены!")
	}

	fmt.Println("\n[Main] Ждем 1.5 секунды (пока кэш протухнет)...")
	time.Sleep(1500 * time.Millisecond)

	val, found = cache.Get("session_token")
	if found {
		fmt.Printf("Через 1.5 сек: данные найдены! Значение: %v\n", val)
	} else {
		fmt.Println("Через 1.5 сек: данные НЕ найдены (время жизни истекло!).")
	}
}
