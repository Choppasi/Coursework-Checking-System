package models

import "time"

type ThesisPoint struct {
	ID          int       `json:"id"`
	ThesisID    int       `json:"thesis_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Order       int       `json:"order"`
	Deadline    *string   `json:"deadline"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreatePointRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Order       int    `json:"order"`
	Deadline    string `json:"deadline"`
}

type UpdatePointRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Order       int    `json:"order"`
	Deadline    string `json:"deadline"`
}

type UpdatePointStatusRequest struct {
	Status string `json:"status"`
}
