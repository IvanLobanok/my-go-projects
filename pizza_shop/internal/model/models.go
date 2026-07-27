package model

import "time"

type Pizza struct {
	ID    int     `json:"id"`
	Title string  `json:"title"`
	Price float64 `json:"price"`
}

type OrderItemDTO struct {
	PizzaID    int     `json:"pizza_id"`
	PizzaTitle string  `json:"pizza_title"`
	Price      float64 `json:"price"`
	Quantity   int     `json:"quantity"`
}

type OrderDTO struct {
	OrderID      int            `json:"order_id"`
	CustomerID   int            `json:"customer_id"`
	CustomerName string         `json:"customer_name"`
	Status       string         `json:"status"`
	TotalPrice   float64        `json:"total_price"`
	Items        []OrderItemDTO `json:"items"`
}

type OrderItemInput struct {
	PizzaID  int `json:"pizza_id" binding:"required,gt=0"`
	Quantity int `json:"quantity" binding:"required,gt=0"`
}

type User struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Balance      float64 `json:"balance"`
	Email        string  `json:"email"`
	PasswordHash string  `json:"-"`
	Role         string  `json:"role"`
}

type RegisterInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role"`
}

type CreateOrderInput struct {
	Items []OrderItemInput `json:"items" binding:"required,dive"`
}

type OrderHistoryItem struct {
	PizzaName string  `json:"pizza_name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}

type OrderHistoryResponse struct {
	ID        int                `json:"order_id"`
	TotalCost float64            `json:"total_cost"`
	CreatedAt time.Time          `json:"created_at"`
	Items     []OrderHistoryItem `json:"items"`
}
