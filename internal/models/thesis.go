package models

import "time"

type Thesis struct {
	ID          int       `json:"id"`
	StudentID   int       `json:"student_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	StartDate   *string   `json:"start_date"`
	Deadline    *string   `json:"deadline"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ThesisWithStudent struct {
	Thesis
	StudentName string `json:"student_name,omitempty"`
}

type CreateThesisRequest struct {
	StudentID   int    `json:"student_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	StartDate   string `json:"start_date"`
	Deadline    string `json:"deadline"`
}

type UpdateThesisRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	StartDate   string `json:"start_date"`
	Deadline    string `json:"deadline"`
}
