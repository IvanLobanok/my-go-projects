package handler

import (
	"log"
	"net/http"
	"pizza_shop/internal/model"

	"github.com/gin-gonic/gin"
)

func (h *Handler) HandleGetUser(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		log.Printf("Не удалось получить userID из контекста: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки данных пользователя"})
		return
	}

	var u *model.User

	u, err = h.repo.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки данных пользователя"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"userID":  userID,
		"name":    u.Name,
		"email":   u.Email,
		"balance": u.Balance,
	})
}
