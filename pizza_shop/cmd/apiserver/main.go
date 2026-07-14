package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"pizza_shop/internal/repository"

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

	pizzas, err := repo.GetPizzas(ctx)

	if err != nil {
		log.Printf("Не удалось получить пиццы: %v", err)
		return
	}
	fmt.Println("--- Список доступных пицц ---")

	for _, p := range pizzas {

		fmt.Printf("[%d] %s — %.2f руб.\n", p.ID, p.Title, p.Price)

	}
}
