package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"todo-app/internal/models"
	"todo-app/internal/repository"

	"github.com/gorilla/mux"
)

type TodoHandler struct {
	repo *repository.TodoRepository
}

func NewTodoHandler(repo *repository.TodoRepository) *TodoHandler {
	return &TodoHandler{repo: repo}
}

func (h *TodoHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	todos, err := h.repo.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if todos == nil {
		todos = []models.Todo{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

func (h *TodoHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	todo, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if todo == nil {
		http.Error(w, "todo not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

func (h *TodoHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	todo := &models.Todo{
		Title:       req.Title,
		Description: req.Description,
		Completed:   false,
	}

	if err := h.repo.Create(todo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(todo)
}

func (h *TodoHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	existingTodo, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if existingTodo == nil {
		http.Error(w, "todo not found", http.StatusNotFound)
		return
	}

	var req models.UpdateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Title != "" {
		existingTodo.Title = req.Title
	}
	if req.Description != "" {
		existingTodo.Description = req.Description
	}
	if req.Completed != nil {
		existingTodo.Completed = *req.Completed
	}

	if err := h.repo.Update(existingTodo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingTodo)
}

func (h *TodoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	existingTodo, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if existingTodo == nil {
		http.Error(w, "todo not found", http.StatusNotFound)
		return
	}

	if err := h.repo.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TodoHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/todos", h.GetAll).Methods(http.MethodGet)
	router.HandleFunc("/todos/{id}", h.GetByID).Methods(http.MethodGet)
	router.HandleFunc("/todos", h.Create).Methods(http.MethodPost)
	router.HandleFunc("/todos/{id}", h.Update).Methods(http.MethodPut)
	router.HandleFunc("/todos/{id}", h.Delete).Methods(http.MethodDelete)
}

// HealthCheck - проверка здоровья сервиса
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Ping - простой пинг
func Ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "pong"})
}
