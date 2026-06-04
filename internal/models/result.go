package models

import "time"

type PointResult struct {
	ID           int        `json:"id"`
	PointID      int        `json:"point_id"`
	StudentID    int        `json:"student_id"`
	Content      string     `json:"content"`
	FileURL      string     `json:"file_url"`
	FileName     string     `json:"file_name"`
	SubmittedAt  time.Time  `json:"submitted_at"`
	Review       *string    `json:"review,omitempty"`
	ReviewStatus string     `json:"review_status"`
	ReviewedAt   *time.Time `json:"reviewed_at,omitempty"`
	ReviewedBy   *int       `json:"reviewed_by,omitempty"`
}

type CreateResultRequest struct {
	Content string `json:"content"`
}

type UpdateResultRequest struct {
	Content string `json:"content"`
}

type ReviewResultRequest struct {
	Review       string `json:"review"`
	ReviewStatus string `json:"review_status"`
}
