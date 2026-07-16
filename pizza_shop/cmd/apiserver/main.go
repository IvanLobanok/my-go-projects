package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"pizza_shop/internal/handler"
	"pizza_shop/internal/repository"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Предупреждение: .env файл не найден")
	}

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("Критическая ошибка: переменная DATABASE_URL не задана в .env")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("Ошибка создания пула соединений БД: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("База данных недоступна: %v", err)
	}
	log.Println("Успешное подключение к PostgreSQL!")

	repo := repository.NewPostgresRepository(pool)
	h := handler.NewHandler(repo)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	fmt.Println("Сервер запущен на http://localhost:8080")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
