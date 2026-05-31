package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"thesis-app/internal/middleware"
	"thesis-app/internal/models"
	"thesis-app/internal/repository"

	"github.com/gorilla/mux"
)

type NotificationHandler struct {
	repo *repository.NotificationRepository
}

func NewNotificationHandler(repo *repository.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{repo: repo}
}

func (h *NotificationHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	list, err := h.repo.GetByUser(userID)
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []models.Notification{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *NotificationHandler) GetUnread(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	list, err := h.repo.GetUnread(userID)
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []models.Notification{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	// Проверяем, что уведомление принадлежит текущему пользователю
	list, err := h.repo.GetByUser(userID)
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}
	found := false
	for _, n := range list {
		if n.ID == id {
			found = true
			break
		}
	}
	if !found {
		http.Error(w, `{"error":"Not found"}`, http.StatusNotFound)
		return
	}
	if err := h.repo.MarkRead(id); err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *NotificationHandler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if err := h.repo.MarkAllRead(userID); err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *NotificationHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/notifications", h.GetAll).Methods("GET")
	router.HandleFunc("/api/notifications/unread", h.GetUnread).Methods("GET")
	router.HandleFunc("/api/notifications/{id}/read", h.MarkRead).Methods("PUT")
	router.HandleFunc("/api/notifications/read-all", h.MarkAllRead).Methods("PUT")
}
