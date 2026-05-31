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

type ThesisHandler struct {
	repo *repository.ThesisRepository
}

func NewThesisHandler(repo *repository.ThesisRepository) *ThesisHandler {
	return &ThesisHandler{repo: repo}
}

func (h *ThesisHandler) canAccess(r *http.Request, thesis *models.ThesisWithStudent) bool {
	if thesis == nil {
		return false
	}
	role := middleware.GetUserRole(r)
	userID := middleware.GetUserID(r)
	if role == "admin" {
		return true
	}
	if role == "teacher" {
		return true // упрощённо: преподаватель видит все курсовые
	}
	return thesis.StudentID == userID
}

func (h *ThesisHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	role := middleware.GetUserRole(r)
	userID := middleware.GetUserID(r)

	var list []models.ThesisWithStudent
	var err error

	if role == "student" {
		list, err = h.repo.GetByStudent(userID)
	} else {
		list, err = h.repo.GetAll()
	}

	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []models.ThesisWithStudent{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *ThesisHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	thesis, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}
	if thesis == nil || !h.canAccess(r, thesis) {
		http.Error(w, `{"error":"Not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(thesis)
}

func (h *ThesisHandler) Create(w http.ResponseWriter, r *http.Request) {
	role := middleware.GetUserRole(r)
	if role == "student" {
		http.Error(w, `{"error":"Forbidden"}`, http.StatusForbidden)
		return
	}
	var req models.CreateThesisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}
	t := &models.Thesis{
		StudentID:   req.StudentID,
		Title:       req.Title,
		Description: req.Description,
		Status:      "planning",
		StartDate:   &req.StartDate,
		Deadline:    &req.Deadline,
	}
	if req.StartDate == "" {
		t.StartDate = nil
	}
	if req.Deadline == "" {
		t.Deadline = nil
	}
	if err := h.repo.Create(t); err != nil {
		http.Error(w, `{"error":"Could not create thesis"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

func (h *ThesisHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	thesis, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}
	if thesis == nil || !h.canAccess(r, thesis) {
		http.Error(w, `{"error":"Not found"}`, http.StatusNotFound)
		return
	}
	var req models.UpdateThesisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}
	t := &models.Thesis{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		StartDate:   &req.StartDate,
		Deadline:    &req.Deadline,
	}
	if req.StartDate == "" {
		t.StartDate = nil
	}
	if req.Deadline == "" {
		t.Deadline = nil
	}
	if err := h.repo.Update(t); err != nil {
		http.Error(w, `{"error":"Could not update thesis"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

func (h *ThesisHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	thesis, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}
	if thesis == nil || !h.canAccess(r, thesis) {
		http.Error(w, `{"error":"Not found"}`, http.StatusNotFound)
		return
	}
	if err := h.repo.Delete(id); err != nil {
		http.Error(w, `{"error":"Could not delete thesis"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ThesisHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/theses", h.GetAll).Methods("GET")
	router.HandleFunc("/api/theses", h.Create).Methods("POST")
	router.HandleFunc("/api/theses/{id}", h.GetByID).Methods("GET")
	router.HandleFunc("/api/theses/{id}", h.Update).Methods("PUT")
	router.HandleFunc("/api/theses/{id}", h.Delete).Methods("DELETE")
}
