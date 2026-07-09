package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Connection struct {
	ID int
}

type Pool struct {
	conns chan *Connection
}

func NewPool(maxConn int) *Pool {
	p := &Pool{
		make(chan *Connection, maxConn),
	}
	for i := 1; i <= maxConn; i++ {
		p.conns <- &Connection{i}
	}
	return p
}

func (p *Pool) Acquire(ctx context.Context) (*Connection, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case conn := <-p.conns:
		return conn, nil
	case <-time.After(3 * time.Second):
		return nil, errors.New("timeout error")
	}
}

func (p *Pool) Release(conn *Connection) {
	p.conns <- conn

}

func main() {
	var wg sync.WaitGroup
	pool := NewPool(2)

	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			fmt.Printf("[Воркер %d] Пытаюсь получить коннект...\n", workerID)
			conn, err := pool.Acquire(ctx)
			if err != nil {
				fmt.Printf("[Воркер %d] Не дождался коннекта: %v\n", workerID, err)
				return
			}

			fmt.Printf("[Воркер %d] Получил коннект №%d\n", workerID, conn.ID)
			time.Sleep(600 * time.Millisecond)

			fmt.Printf("[Воркер %d] Возвращаю коннект №%d в пул\n", workerID, conn.ID)
			pool.Release(conn)
		}(i)
	}
	wg.Wait()
	fmt.Println("\n[Main] Все воркеры отработали.")
}
