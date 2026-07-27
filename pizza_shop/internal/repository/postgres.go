package repository

import (
	"context"
	"fmt"
	"time"

	"pizza_shop/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
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

func (r *PostgresRepository) AddPizza(ctx context.Context, p *model.Pizza) error {
	query := "INSERT INTO pizzas (title,price) VALUES ($1,$2) RETURNING id"

	err := r.db.QueryRow(ctx, query, p.Title, p.Price).Scan(&p.ID)

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
	defer tx.Rollback(context.Background())

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

func (r *PostgresRepository) NewUser(ctx context.Context, name, email, passwordHash string) (int, error) {
	if name == "" || email == "" || passwordHash == "" {
		return 0, fmt.Errorf("имя пользователя, email и пароль не могут быть пустыми")
	}
	defaultRole := "defaultUser"

	var newUserID int
	query := `
			INSERT INTO users (name, email, password_hash, balance, role) 
			VALUES ($1,$2,$3,$4,$5) 
			RETURNING id	
	`
	err := r.db.QueryRow(ctx, query, name, email, passwordHash, 0.00, defaultRole).Scan(&newUserID)
	if err != nil {
		return 0, fmt.Errorf("Вставки нового пользователя: %w", err)
	}
	return newUserID, nil
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, name, balance, email, password_hash, role 
		FROM users 
		WHERE email = $1
	`
	var u model.User
	err := r.db.QueryRow(ctx, query, email).Scan(&u.ID, &u.Name, &u.Balance, &u.Email, &u.PasswordHash, &u.Role)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, userID int) (*model.User, error) {
	query := `
		SELECT id, name, balance, email, password_hash, role
		FROM users 
		WHERE id = $1
	`
	var u model.User
	err := r.db.QueryRow(ctx, query, userID).Scan(&u.ID, &u.Name, &u.Balance, &u.Email, &u.PasswordHash, &u.Role)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *PostgresRepository) GetUserOrders(ctx context.Context, userID int) ([]model.OrderHistoryResponse, error) {
	query := `
		SELECT 
			o.id, o.total_cost, o.created_at,
			p.name, oi.price, oi.quantity
		FROM orders o
		JOIN order_items oi ON o.id = oi.order_id
		JOIN pizzas p ON oi.pizza_id = p.id
		WHERE o.user_id = $1
		ORDER BY o.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("Ошибка получения истории заказов: %v", err)
	}
	defer rows.Close()

	orderMap := make(map[int]*model.OrderHistoryResponse)
	var orderIDs []int

	for rows.Next() {
		var orderID, quantity int
		var totalCost, price float64
		var createdAt time.Time
		var pizzaName string

		err := rows.Scan(&orderID, &totalCost, &createdAt, &pizzaName, &price, &quantity)
		if err != nil {
			return nil, fmt.Errorf("Ошибка сканирования строки: %v", err)

		}
		if _, exists := orderMap[orderID]; !exists {
			orderMap[orderID] = &model.OrderHistoryResponse{
				ID:        orderID,
				TotalCost: totalCost,
				CreatedAt: createdAt,
				Items:     []model.OrderHistoryItem{},
			}
			orderIDs = append(orderIDs, orderID)
		}
		orderMap[orderID].Items = append(orderMap[orderID].Items, model.OrderHistoryItem{
			PizzaName: pizzaName,
			Price:     price,
			Quantity:  quantity,
		})
	}
	orders := make([]model.OrderHistoryResponse, 0, len(orderIDs))
	for _, id := range orderIDs {
		orders = append(orders, *orderMap[id])
	}

	return orders, nil
}
