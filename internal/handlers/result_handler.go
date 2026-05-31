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

type ResultHandler struct {
	repo         *repository.ResultRepository
	pointRepo    *repository.PointRepository
	thesisRepo   *repository.ThesisRepository
	notifService *services.NotificationService
	fileService  *services.FileService
}

func NewResultHandler(
	repo *repository.ResultRepository,
	pointRepo *repository.PointRepository,
	thesisRepo *repository.ThesisRepository,
	notif *services.NotificationService,
	file *services.FileService,
) *ResultHandler {
	return &ResultHandler{repo: repo, pointRepo: pointRepo, thesisRepo: thesisRepo, notifService: notif, fileService: file}
}

func (h *ResultHandler) canAccessPoint(r *http.Request, pointID int) bool {
	role := middleware.GetUserRole(r)
	userID := middleware.GetUserID(r)
	if role == "admin" || role == "teacher" {
		return true
	}
	point, err := h.pointRepo.GetByID(pointID)
	if err != nil || point == nil {
		return false
	}
	thesis, err := h.thesisRepo.GetByID(point.ThesisID)
	if err != nil || thesis == nil {
		return false
	}
	return thesis.StudentID == userID
}

func (h *ResultHandler) canAccessResult(r *http.Request, resultID int) bool {
	role := middleware.GetUserRole(r)
	userID := middleware.GetUserID(r)
	if role == "admin" || role == "teacher" {
		return true
	}
	res, err := h.repo.GetByID(resultID)
	if err != nil || res == nil {
		return false
	}
	return res.StudentID == userID
}

func (h *ResultHandler) GetByPoint(w http.ResponseWriter, r *http.Request) {
	pointID, _ := strconv.Atoi(mux.Vars(r)["id"])
	if !h.canAccessPoint(r, pointID) {
		http.Error(w, `{"error":"Not found"}`, http.StatusNotFound)
		return
	}
	results, err := h.repo.GetByPoint(pointID)
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}
	if results == nil {
		results = []models.PointResult{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (h *ResultHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	if !h.canAccessResult(r, id) {
		http.Error(w, `{"error":"Not found"}`, http.StatusNotFound)
		return
	}
	res, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}
	if res == nil {
		http.Error(w, `{"error":"Not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *ResultHandler) Create(w http.ResponseWriter, r *http.Request) {
	pointID, _ := strconv.Atoi(mux.Vars(r)["id"])
	userID := middleware.GetUserID(r)
	if !h.canAccessPoint(r, pointID) {
		http.Error(w, `{"error":"Not found"}`, http.StatusNotFound)
		return
	}

	content := r.FormValue("content")
	fileURL, fileName, err := h.fileService.SaveFile(r, "file")
	if err != nil {
		http.Error(w, `{"error":"File upload failed"}`, http.StatusBadRequest)
		return
	}

	res := &models.PointResult{
		PointID:   pointID,
		StudentID: userID,
		Content:   content,
		FileURL:   fileURL,
		FileName:  fileName,
	}
	if err := h.repo.Create(res); err != nil {
		http.Error(w, `{"error":"Could not create result"}`, http.StatusInternalServerError)
		return
	}

	// Обновляем статус пункта на in_progress
	h.pointRepo.UpdateStatus(pointID, "in_progress")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *ResultHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	if !h.canAccessResult(r, id) {
		http.Error(w, `{"error":"Not found"}`, http.StatusNotFound)
		return
	}
	var req models.UpdateResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}
	res, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}
	if res == nil {
		http.Error(w, `{"error":"Not found"}`, http.StatusNotFound)
		return
	}
	res.Content = req.Content
	if err := h.repo.Update(res); err != nil {
		http.Error(w, `{"error":"Could not update result"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *ResultHandler) Review(w http.ResponseWriter, r *http.Request) {
	role := middleware.GetUserRole(r)
	if role == "student" {
		http.Error(w, `{"error":"Forbidden"}`, http.StatusForbidden)
		return
	}
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	reviewerID := middleware.GetUserID(r)
	var req models.ReviewResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}
	res, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}
	if res == nil {
		http.Error(w, `{"error":"Not found"}`, http.StatusNotFound)
		return
	}
	if err := h.repo.Review(id, req.Review, req.ReviewStatus, reviewerID); err != nil {
		http.Error(w, `{"error":"Could not review result"}`, http.StatusInternalServerError)
		return
	}

	// Обновляем статус пункта
	if req.ReviewStatus == "approved" {
		h.pointRepo.UpdateStatus(res.PointID, "done")
	} else if req.ReviewStatus == "rejected" {
		h.pointRepo.UpdateStatus(res.PointID, "rejected")
	}

	// Уведомляем студента
	if point, _ := h.pointRepo.GetByID(res.PointID); point != nil {
		if thesis, _ := h.thesisRepo.GetByID(point.ThesisID); thesis != nil {
			h.notifService.NotifyReviewed(thesis.StudentID, point.Title, req.ReviewStatus)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Reviewed"})
}

func (h *ResultHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/points/{id}/results", h.GetByPoint).Methods("GET")
	router.HandleFunc("/api/points/{id}/results", h.Create).Methods("POST")
	router.HandleFunc("/api/results/{id}", h.GetByID).Methods("GET")
	router.HandleFunc("/api/results/{id}", h.Update).Methods("PUT")
	router.HandleFunc("/api/results/{id}/review", h.Review).Methods("PUT")
}
