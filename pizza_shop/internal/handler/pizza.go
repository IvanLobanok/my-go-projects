package handler

import (
	"log"
	"net/http"
	"pizza_shop/internal/model"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) HandleGetPizzas(c *gin.Context) {
	pizzas, err := h.repo.GetPizzas(c.Request.Context())
	if err != nil {
		log.Printf("Ошибка получения пицц из бд: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Внутренняя ошибка сервера"})
		return
	}

	c.JSON(http.StatusOK, pizzas)
}

func (h *Handler) HandleCreatePizza(c *gin.Context) {
	var input model.Pizza

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный JSON-body"})
		return

	}

	if err := h.repo.AddPizza(c.Request.Context(), &input); err != nil {
		log.Printf("Ошибка сохранения пиццы в бд: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сохранить пиццу"})
		return
	}

	c.JSON(http.StatusCreated, input)
}

func (h *Handler) HandleDeletePizza(c *gin.Context) {
	idStr := c.Param("id")

	pizzaID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат ID пиццы"})
		return
	}

	err = h.repo.DeletePizza(c.Request.Context(), pizzaID)
	if err != nil {
		log.Printf("Не удалось удалить пиццу: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки данных пользователя"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Пицца удалена из базы данных",
	})

}
