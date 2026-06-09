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

type GroupHandler struct {
	repo *repository.GroupRepository
}

func NewGroupHandler(repo *repository.GroupRepository) *GroupHandler {
	return &GroupHandler{repo: repo}
}

func (h *GroupHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	groups, err := h.repo.GetAll()
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}
	if groups == nil {
		groups = []models.GroupWithTeacher{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}

func (h *GroupHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	group, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}
	if group == nil {
		http.Error(w, `{"error":"Not found"}`, http.StatusNotFound)
		return
	}
	members, _ := h.repo.GetMembers(id)
	if members == nil {
		members = []models.User{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"group":   group,
		"members": members,
	})
}

func (h *GroupHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var req models.CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	// Преподаватель создает группу только себе
	g := &models.Group{
		Name:      req.Name,
		TeacherID: userID, // Автоматически подставляем ID преподавателя
		Course:    req.Course,
		Year:      req.Year,
	}
	if err := h.repo.Create(g); err != nil {
		http.Error(w, `{"error":"Could not create group"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(g)
}

func (h *GroupHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	userRole := middleware.GetUserRole(r)

	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	var req models.UpdateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	// Проверка прав: только админ или преподаватель этой группы
	group, err := h.repo.GetByID(id)
	if err != nil || group == nil {
		http.Error(w, `{"error":"Group not found"}`, http.StatusNotFound)
		return
	}

	if userRole != "admin" && group.TeacherID != userID {
		http.Error(w, `{"error":"Only group teacher or admin can update"}`, http.StatusForbidden)
		return
	}

	// Преподаватель не может изменить преподавателя группы
	g := &models.Group{
		ID:        id,
		Name:      req.Name,
		TeacherID: group.TeacherID, // Оставляем старого преподавателя
		Course:    req.Course,
		Year:      req.Year,
	}
	if err := h.repo.Update(g); err != nil {
		http.Error(w, `{"error":"Could not update group"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(g)
}

func (h *GroupHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	if err := h.repo.Delete(id); err != nil {
		http.Error(w, `{"error":"Could not delete group"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *GroupHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	userRole := middleware.GetUserRole(r)

	groupID, _ := strconv.Atoi(mux.Vars(r)["id"])
	var body struct {
		StudentID int `json:"student_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	// Проверка прав доступа: только админ или преподаватель группы
	if userRole != "admin" {
		// Проверяем, является ли пользователь преподавателем этой группы
		group, err := h.repo.GetByID(groupID)
		if err != nil || group == nil {
			http.Error(w, `{"error":"Group not found"}`, http.StatusNotFound)
			return
		}
		if group.TeacherID != userID {
			http.Error(w, `{"error":"Only group teacher or admin can add members"}`, http.StatusForbidden)
			return
		}
	}

	if err := h.repo.AddMember(groupID, body.StudentID); err != nil {
		http.Error(w, `{"error":"Could not add member"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Member added"})
}

func (h *GroupHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	userRole := middleware.GetUserRole(r)

	vars := mux.Vars(r)
	groupID, _ := strconv.Atoi(vars["id"])
	studentID, _ := strconv.Atoi(vars["studentId"])

	// Проверка прав доступа: только админ или преподаватель группы
	if userRole != "admin" {
		group, err := h.repo.GetByID(groupID)
		if err != nil || group == nil {
			http.Error(w, `{"error":"Group not found"}`, http.StatusNotFound)
			return
		}
		if group.TeacherID != userID {
			http.Error(w, `{"error":"Only group teacher or admin can remove members"}`, http.StatusForbidden)
			return
		}
	}

	if err := h.repo.RemoveMember(groupID, studentID); err != nil {
		http.Error(w, `{"error":"Could not remove member"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *GroupHandler) GetMyGroup(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	group, err := h.repo.GetStudentGroup(userID)
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}
	if group == nil {
		http.Error(w, `{"error":"Not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(group)
}

func (h *GroupHandler) JoinGroup(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	role := middleware.GetUserRole(r)

	// Только студенты могут записываться в группы
	if role != "student" {
		http.Error(w, `{"error":"Only students can join groups"}`, http.StatusForbidden)
		return
	}

	groupID, _ := strconv.Atoi(mux.Vars(r)["id"])

	// Проверяем, существует ли группа
	group, err := h.repo.GetByID(groupID)
	if err != nil || group == nil {
		http.Error(w, `{"error":"Group not found"}`, http.StatusNotFound)
		return
	}

	// Проверяем, не состоит ли студент уже в группе
	existing, _ := h.repo.GetStudentGroup(userID)
	if existing != nil {
		http.Error(w, `{"error":"You are already in a group"}`, http.StatusBadRequest)
		return
	}

	if err := h.repo.AddMember(groupID, userID); err != nil {
		http.Error(w, `{"error":"Could not join group"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Successfully joined group"})
}

func (h *GroupHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/groups", h.GetAll).Methods("GET")
	router.HandleFunc("/api/groups", h.Create).Methods("POST")
	router.HandleFunc("/api/groups/{id}", h.GetByID).Methods("GET")
	router.HandleFunc("/api/groups/{id}", h.Update).Methods("PUT")
	router.HandleFunc("/api/groups/{id}", h.Delete).Methods("DELETE")
	router.HandleFunc("/api/groups/{id}/members", h.AddMember).Methods("POST")
	router.HandleFunc("/api/groups/{id}/members/{studentId}", h.RemoveMember).Methods("DELETE")
	router.HandleFunc("/api/groups/my", h.GetMyGroup).Methods("GET")
	router.HandleFunc("/api/groups/{id}/join", h.JoinGroup).Methods("POST")
}
