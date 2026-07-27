package handler

import (
	"log"
	"net/http"
	"pizza_shop/internal/model"

	"github.com/gin-gonic/gin"
)

func (h *Handler) HandleCreateOrder(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		log.Printf("Не удалось получить userID из контекста: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки данных пользователя"})
		return
	}

	var input model.CreateOrderInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Некорректный состав заказа",
			"details": err.Error(),
		})
		return
	}

	err = h.repo.NewOrder(c.Request.Context(), userID, input.Items)
	if err != nil {
		log.Printf("Ошибка при создании заказа: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Не удалось оформить заказ",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Заказ успешно оформлен и уже готовится!",
	})
}

func (h *Handler) HandleGetOrders(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		log.Printf("Не удалось получить userID из контекста: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки данных пользователя"})
		return
	}

	orders, err := h.repo.GetUserOrders(c.Request.Context(), userID)
	if err != nil {
		log.Printf("Ошибка получения истории заказов для пользователя : %d", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения списка заказов"})
		return
	}

	c.JSON(http.StatusOK, orders)
}
