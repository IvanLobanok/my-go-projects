package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Request struct {
	UserID int
	Path   string
}

type RateLimiter struct {
	mu          sync.Mutex
	counters    map[int]int
	maxRequests int
}

type AuditStorage struct {
	mu       sync.Mutex
	Approved []string
	Rejected []string
}

func (rl *RateLimiter) Allow(userID int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	requests := rl.counters[userID]
	if requests >= rl.maxRequests {
		return false
	}
	rl.counters[userID] = requests + 1
	return true
}

func StartCleaner(ctx context.Context, rl *RateLimiter) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rl.mu.Lock()
			rl.counters = make(map[int]int)
			rl.mu.Unlock()
		}
	}
}

func RequestProcessor(id int, ctx context.Context, requests <-chan Request, limiter *RateLimiter, storage *AuditStorage, wg *sync.WaitGroup) {
	defer wg.Done()

	for req := range requests {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if limiter.Allow(req.UserID) {
			requestResult := fmt.Sprintf("Воркер %d: Пользователь %d допущен к %s", id, req.UserID, req.Path)
			storage.mu.Lock()
			storage.Approved = append(storage.Approved, requestResult)
			storage.mu.Unlock()
		} else {
			requestResult := fmt.Sprintf("Воркер %d: Пользователь %d НЕ допущен к %s (429)", id, req.UserID, req.Path)
			storage.mu.Lock()
			storage.Rejected = append(storage.Rejected, requestResult)
			storage.mu.Unlock()
		}
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	limiter := RateLimiter{
		counters:    make(map[int]int),
		maxRequests: 3,
	}
	storage := AuditStorage{}

	requestsChan := make(chan Request, 50)
	var wg sync.WaitGroup

	go StartCleaner(ctx, &limiter)

	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go RequestProcessor(i, ctx, requestsChan, &limiter, &storage, &wg)
	}

	for i := 1; i <= 10; i++ {
		requestsChan <- Request{UserID: 77, Path: fmt.Sprintf("/api/profile/get-%d", i)}
	}
	requestsChan <- Request{UserID: 99, Path: "/api/dashboard"}
	requestsChan <- Request{UserID: 99, Path: "/api/settings"}

	time.Sleep(1 * time.Second)

	requestsChan <- Request{UserID: 77, Path: "/api/profile/new-wave-1"}
	requestsChan <- Request{UserID: 77, Path: "/api/profile/new-wave-2"}

	close(requestsChan)
	wg.Wait()

	fmt.Println("\n==========================================")
	fmt.Printf("УСПЕШНЫЕ ЗАПРОСЫ (%d):\n", len(storage.Approved))
	for _, res := range storage.Approved {
		fmt.Println("  ", res)
	}

	fmt.Printf("\nЗАБЛОКИРОВАННЫЕ ЗАПРОСЫ (%d):\n", len(storage.Rejected))
	for _, res := range storage.Rejected {
		fmt.Println("  ", res)
	}
}
