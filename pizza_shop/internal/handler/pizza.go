package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"pizza_shop/internal/model"
	"pizza_shop/internal/repository"
)

type Handler struct {
	repo *repository.PostgresRepository
}

func NewHandler(repo *repository.PostgresRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/pizzas", h.HandleGetPizzas)
	mux.HandleFunc("POST /api/v1/pizzas", h.HandleCreatePizza)
}

func (h *Handler) HandleGetPizzas(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	pizzas, err := h.repo.GetPizzas(r.Context())
	if err != nil {
		log.Printf("Ошибка получения пицц из бд: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(pizzas)
	if err != nil {
		http.Error(w, "Ошибка кодирования JSON", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleCreatePizza(w http.ResponseWriter, r *http.Request) {
	var input model.Pizza

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Некорректный JSON-body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	err = h.repo.AddPizza(r.Context(), &input)
	if err != nil {
		log.Printf("Ошибка сохранения пиццы в бд: %v", err)
		http.Error(w, "Не удалось сохранить пиццу", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}
