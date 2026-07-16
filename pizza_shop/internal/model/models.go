package model

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
	PizzaID  int
	Quantity int
}
