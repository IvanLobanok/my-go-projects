package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"pizza_shop/internal/model"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetPizzas(ctx context.Context) ([]model.Pizza, error) {
	query := "SELECT id, title,price FROM pizzas ORDER BY id"

	rows, err := r.db.Query(ctx, query)

	if err != nil {
		return nil, fmt.Errorf("Ошибка чтения строк:%v", err)
	}
	defer rows.Close()

	pizzas := make([]model.Pizza, 0)

	for rows.Next() {
		var p model.Pizza
		err = rows.Scan(&p.ID, &p.Title, &p.Price)
		if err != nil {
			return nil, fmt.Errorf("Ошибка чтения строк:%v", err)
		}
		pizzas = append(pizzas, p)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("Ошибка чтения строк:%v", err)
	}

	return pizzas, nil
}

func (r *PostgresRepository) AddPizza(ctx context.Context, title string, price float64) error {
	query := "INSERT INTO pizzas (title,price) VALUES ($1,$2)"

	_, err := r.db.Exec(ctx, query, title, price)

	if err != nil {
		return fmt.Errorf("Ошибка добавления строки:%v", err)
	}

	return nil
}

func (r *PostgresRepository) DeletePizza(ctx context.Context, id int) error {
	query := "DELETE FROM pizzas WHERE id = $1"

	_, err := r.db.Exec(ctx, query, id)

	if err != nil {
		return fmt.Errorf("Ошибка удаление строки:%v", err)
	}

	return nil
}

func (r *PostgresRepository) GetOrders(ctx context.Context) ([]model.OrderDTO, error) {
	query := `
		SELECT o.id, o.user_id, u.name, o.status, o.total_price, p.id, p.title, oi.quantity, p.price
		FROM order_items oi
		JOIN orders o ON oi.order_id = o.id
		JOIN users u ON o.user_id = u.id
		JOIN pizzas p ON oi.pizza_id = p.id
		ORDER BY o.id;
	`
	rows, err := r.db.Query(ctx, query)

	if err != nil {
		return nil, fmt.Errorf("Ошибка чтения строк:%v", err)
	}
	defer rows.Close()

	ordersMap := make(map[int]*model.OrderDTO)
	orderIDs := make([]int, 0)

	for rows.Next() {
		var orderID, customerID, pizzaID, quantity int
		var customerName, status, pizzaTitle string
		var totalPrice, pizzaPrice float64
		err = rows.Scan(&orderID, &customerID, &customerName, &status, &totalPrice, &pizzaID, &pizzaTitle, &quantity, &pizzaPrice)
		if err != nil {
			return nil, fmt.Errorf("Ошибка чтения строк:%v", err)
		}

		ord, exist := ordersMap[orderID]
		if !exist {
			ord = &model.OrderDTO{
				OrderID:      orderID,
				CustomerID:   customerID,
				CustomerName: customerName,
				Status:       status,
				TotalPrice:   totalPrice,
				Items:        make([]model.OrderItemDTO, 0),
			}
			ordersMap[orderID] = ord
			orderIDs = append(orderIDs, orderID)
		}
		ord.Items = append(ord.Items, model.OrderItemDTO{
			PizzaID:    pizzaID,
			PizzaTitle: pizzaTitle,
			Price:      pizzaPrice,
			Quantity:   quantity,
		})

	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("Ошибка чтения строк:%v", err)
	}
	result := make([]model.OrderDTO, 0, len(ordersMap))
	for _, id := range orderIDs {
		result = append(result, *ordersMap[id])
	}
	return result, nil
}

func (r *PostgresRepository) NewOrder(ctx context.Context, userID int, items []model.OrderItemInput) error {
	if len(items) == 0 {
		return fmt.Errorf("Ошибка: невозможно добавить пустой заказ")
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("Ошибка создания транзакции:%v", err)
	}
	defer tx.Rollback(ctx)

	var userExist bool

	checkUserQuery := "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)"
	err = tx.QueryRow(ctx, checkUserQuery, userID).Scan(&userExist)
	if err != nil {
		return fmt.Errorf("Ошибка проверки существования пользователя:%v", err)
	}

	if !userExist {
		return fmt.Errorf("Ошибка пользователь не найден:%v", err)
	}
	var orderID int
	orderQuery := "INSERT INTO orders (user_id, status, total_price) VALUES ($1,'pending', 0.0) RETURNING id"
	err = tx.QueryRow(ctx, orderQuery, userID).Scan(&orderID)
	if err != nil {
		return fmt.Errorf("Ошибка создания заказа:%v", err)
	}

	var totalPrice float64

	for _, item := range items {
		var pizzaPrice float64
		pizzaPriceQuery := "SELECT price FROM pizzas WHERE id = $1"
		err = tx.QueryRow(ctx, pizzaPriceQuery, item.PizzaID).Scan(&pizzaPrice)
		if err != nil {
			return fmt.Errorf("Ошибка чтения цены пиццы:%v", err)
		}
		itemQuery := "INSERT INTO order_items (order_id, pizza_id, quantity) VALUES ($1, $2, $3)"
		_, err = tx.Exec(ctx, itemQuery, orderID, item.PizzaID, item.Quantity)
		if err != nil {
			return fmt.Errorf("Ошибка добавления позиции %d в заказ: %w", item.PizzaID, err)
		}

		totalPrice += pizzaPrice * float64(item.Quantity)

	}

	updateQuery := "UPDATE orders SET total_price = $1 WHERE id = $2"

	_, err = tx.Exec(ctx, updateQuery, totalPrice, orderID)
	if err != nil {
		return fmt.Errorf("ошибка обновления общей стоимости заказа: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("ошибка завершения транзакции: %w", err)
	}
	fmt.Println("Успешной добавлен заказ!")
	return nil
}

func (r *PostgresRepository) NewUser(ctx context.Context, name string) (int, error) {
	if name == "" {
		return 0, fmt.Errorf("имя пользователя не может быть пустым")
	}
	var newUserID int
	query := "INSERT INTO users (name) VALUES ($1) RETURNING id	"
	err := r.db.QueryRow(ctx, query, name).Scan(&newUserID)
	if err != nil {
		return 0, fmt.Errorf("Ошибка вставки нового пользователя")
	}
	return newUserID, nil
}
