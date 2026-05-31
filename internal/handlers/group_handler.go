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
	var req models.CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}
	g := &models.Group{
		Name:      req.Name,
		TeacherID: req.TeacherID,
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
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	var req models.UpdateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}
	g := &models.Group{
		ID:        id,
		Name:      req.Name,
		TeacherID: req.TeacherID,
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
	groupID, _ := strconv.Atoi(mux.Vars(r)["id"])
	var body struct {
		StudentID int `json:"student_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}
	if err := h.repo.AddMember(groupID, body.StudentID); err != nil {
		http.Error(w, `{"error":"Could not add member"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Member added"})
}

func (h *GroupHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID, _ := strconv.Atoi(vars["id"])
	studentID, _ := strconv.Atoi(vars["studentId"])
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

func (h *GroupHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/groups", h.GetAll).Methods("GET")
	router.HandleFunc("/api/groups", h.Create).Methods("POST")
	router.HandleFunc("/api/groups/{id}", h.GetByID).Methods("GET")
	router.HandleFunc("/api/groups/{id}", h.Update).Methods("PUT")
	router.HandleFunc("/api/groups/{id}", h.Delete).Methods("DELETE")
	router.HandleFunc("/api/groups/{id}/members", h.AddMember).Methods("POST")
	router.HandleFunc("/api/groups/{id}/members/{studentId}", h.RemoveMember).Methods("DELETE")
	router.HandleFunc("/api/groups/my", h.GetMyGroup).Methods("GET")
}
