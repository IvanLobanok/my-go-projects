package handler

import (
	"fmt"
	"pizza_shop/internal/repository"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo *repository.PostgresRepository
}

func NewHandler(repo *repository.PostgresRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			v1.GET("/pizzas", h.HandleGetPizzas)
			v1.POST("/pizzas", AuthMiddleware(), AdminOnly(), h.HandleCreatePizza)
			v1.DELETE("/pizzas/:id", AuthMiddleware(), AdminOnly(), h.HandleDeletePizza)

			v1.POST("/orders", AuthMiddleware(), h.HandleCreateOrder)
			v1.GET("/orders", AuthMiddleware(), h.HandleGetOrders)
		}

		auth := api.Group("/auth")
		{
			auth.POST("/register", h.HandleRegister)
			auth.POST("/login", h.HandleLogin)
		}

		users := api.Group("/users")
		{
			users.GET("/me", AuthMiddleware(), h.HandleGetUser)
		}
	}
}

func (h *Handler) getUserID(c *gin.Context) (int, error) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		return 0, fmt.Errorf("unauthorized")
	}

	userID, ok := userIDVal.(int)
	if !ok {
		return 0, fmt.Errorf("invalid user id type")
	}

	return userID, nil
}
