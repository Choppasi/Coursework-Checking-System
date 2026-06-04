package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"thesis-app/internal/middleware"
	"thesis-app/internal/models"
	"thesis-app/internal/repository"
	"thesis-app/internal/services"

	"github.com/gorilla/mux"
)

type PointHandler struct {
	repo         *repository.PointRepository
	thesisRepo   *repository.ThesisRepository
	notifService *services.NotificationService
}

func NewPointHandler(repo *repository.PointRepository, thesisRepo *repository.ThesisRepository, notif *services.NotificationService) *PointHandler {
	return &PointHandler{repo: repo, thesisRepo: thesisRepo, notifService: notif}
}

func (h *PointHandler) canAccessThesis(r *http.Request, thesisID int) bool {
	role := middleware.GetUserRole(r)
	userID := middleware.GetUserID(r)
	if role == "admin" || role == "teacher" {
		return true
	}
	thesis, err := h.thesisRepo.GetByID(thesisID)
	if err != nil || thesis == nil {
		return false
	}
	return thesis.StudentID == userID
}

func (h *PointHandler) GetByThesis(w http.ResponseWriter, r *http.Request) {
	thesisID, _ := strconv.Atoi(mux.Vars(r)["id"])
	if !h.canAccessThesis(r, thesisID) {
		http.Error(w, `{"error":"Not found"}`, http.StatusNotFound)
		return
	}
	points, err := h.repo.GetByThesis(thesisID)
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}
	if points == nil {
		points = []models.ThesisPoint{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(points)
}

func (h *PointHandler) canAccessPoint(r *http.Request, pointID int) bool {
	role := middleware.GetUserRole(r)
	if role == "admin" || role == "teacher" {
		return true
	}
	point, err := h.repo.GetByID(pointID)
	if err != nil || point == nil {
		return false
	}
	return h.canAccessThesis(r, point.ThesisID)
}

func (h *PointHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	if !h.canAccessPoint(r, id) {
		http.Error(w, `{"error":"Not found"}`, http.StatusNotFound)
		return
	}
	point, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}
	if point == nil {
		http.Error(w, `{"error":"Not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(point)
}

func (h *PointHandler) Create(w http.ResponseWriter, r *http.Request) {
	role := middleware.GetUserRole(r)
	if role == "student" {
		http.Error(w, `{"error":"Forbidden"}`, http.StatusForbidden)
		return
	}
	thesisID, _ := strconv.Atoi(mux.Vars(r)["id"])
	var req models.CreatePointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}
	p := &models.ThesisPoint{
		ThesisID:    thesisID,
		Title:       req.Title,
		Description: req.Description,
		Order:       req.Order,
		Status:      "pending",
	}
	if req.Deadline != "" {
		p.Deadline = &req.Deadline
	}
	if err := h.repo.Create(p); err != nil {
		http.Error(w, `{"error":"Could not create point"}`, http.StatusInternalServerError)
		return
	}

	// Уведомляем студента
	if thesis, _ := h.thesisRepo.GetByID(thesisID); thesis != nil {
		h.notifService.NotifyNewPoint(thesis.StudentID, p.Title)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func (h *PointHandler) Update(w http.ResponseWriter, r *http.Request) {
	role := middleware.GetUserRole(r)
	if role == "student" {
		http.Error(w, `{"error":"Forbidden"}`, http.StatusForbidden)
		return
	}
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	var req models.UpdatePointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}
	existing, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}
	if existing == nil {
		http.Error(w, `{"error":"Not found"}`, http.StatusNotFound)
		return
	}
	existing.Title = req.Title
	existing.Description = req.Description
	existing.Order = req.Order
	if req.Deadline != "" {
		existing.Deadline = &req.Deadline
	} else {
		existing.Deadline = nil
	}
	if err := h.repo.Update(existing); err != nil {
		http.Error(w, `{"error":"Could not update point"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existing)
}

func (h *PointHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	if !h.canAccessPoint(r, id) {
		http.Error(w, `{"error":"Not found"}`, http.StatusNotFound)
		return
	}
	var req models.UpdatePointStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}
	if err := h.repo.UpdateStatus(id, req.Status); err != nil {
		http.Error(w, `{"error":"Could not update status"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": req.Status})
}

func (h *PointHandler) Delete(w http.ResponseWriter, r *http.Request) {
	role := middleware.GetUserRole(r)
	if role == "student" {
		http.Error(w, `{"error":"Forbidden"}`, http.StatusForbidden)
		return
	}
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	if err := h.repo.Delete(id); err != nil {
		http.Error(w, `{"error":"Could not delete point"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *PointHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/theses/{id}/points", h.GetByThesis).Methods("GET")
	router.HandleFunc("/theses/{id}/points", h.Create).Methods("POST")
	router.HandleFunc("/points/{id}", h.GetByID).Methods("GET")
	router.HandleFunc("/points/{id}", h.Update).Methods("PUT")
	router.HandleFunc("/points/{id}", h.Delete).Methods("DELETE")
	router.HandleFunc("/points/{id}/status", h.UpdateStatus).Methods("PUT")
}
