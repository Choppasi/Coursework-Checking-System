package models

import "time"

type Group struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	TeacherID int       `json:"teacher_id"`
	Course    int       `json:"course"`
	Year      int       `json:"year"`
	CreatedAt time.Time `json:"created_at"`
}

type GroupWithTeacher struct {
	Group
	TeacherName string `json:"teacher_name,omitempty"`
}

type GroupMember struct {
	ID        int       `json:"id"`
	GroupID   int       `json:"group_id"`
	StudentID int       `json:"student_id"`
	JoinedAt  time.Time `json:"joined_at"`
}

type CreateGroupRequest struct {
	Name      string `json:"name"`
	TeacherID int    `json:"teacher_id"`
	Course    int    `json:"course"`
	Year      int    `json:"year"`
}

type UpdateGroupRequest struct {
	Name      string `json:"name"`
	TeacherID int    `json:"teacher_id"`
	Course    int    `json:"course"`
	Year      int    `json:"year"`
}
