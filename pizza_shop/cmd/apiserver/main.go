package main

import (
	"context"
	"log"
	"os"
	"pizza_shop/internal/handler"
	"pizza_shop/internal/repository"
	"time"

	"github.com/gin-gonic/gin"
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

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = ":8080"
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

	r := gin.Default()
	h.RegisterRoutes(r)

	log.Printf("Сервер запущен на http://localhost%s\n", port)

	if err := r.Run(port); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
