package repository

import (
	"database/sql"
	"thesis-app/internal/models"
)

type ResultRepository struct {
	db *sql.DB
}

func NewResultRepository(db *sql.DB) *ResultRepository {
	return &ResultRepository{db: db}
}

func (r *ResultRepository) Create(res *models.PointResult) error {
	return r.db.QueryRow(
		`INSERT INTO point_results (point_id, student_id, content, file_url, file_name, review_status)
		VALUES ($1, $2, $3, $4, $5, 'pending')
		RETURNING id, submitted_at`,
		res.PointID, res.StudentID, res.Content, res.FileURL, res.FileName,
	).Scan(&res.ID, &res.SubmittedAt)
}

func (r *ResultRepository) GetByPoint(pointID int) ([]models.PointResult, error) {
	rows, err := r.db.Query(`
		SELECT id, point_id, student_id, content, file_url, file_name, submitted_at, review, review_status, reviewed_at, reviewed_by
		FROM point_results WHERE point_id = $1 ORDER BY submitted_at DESC
	`, pointID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.PointResult
	for rows.Next() {
		var res models.PointResult
		var review sql.NullString
		var reviewedAt sql.NullTime
		var reviewedBy sql.NullInt64
		if err := rows.Scan(&res.ID, &res.PointID, &res.StudentID, &res.Content, &res.FileURL, &res.FileName, &res.SubmittedAt, &review, &res.ReviewStatus, &reviewedAt, &reviewedBy); err != nil {
			return nil, err
		}
		if review.Valid {
			res.Review = &review.String
		}
		if reviewedAt.Valid {
			rt := reviewedAt.Time
			res.ReviewedAt = &rt
		}
		if reviewedBy.Valid {
			rb := int(reviewedBy.Int64)
			res.ReviewedBy = &rb
		}
		list = append(list, res)
	}
	return list, nil
}

func (r *ResultRepository) GetByID(id int) (*models.PointResult, error) {
	var res models.PointResult
	var review sql.NullString
	var reviewedAt sql.NullTime
	var reviewedBy sql.NullInt64
	err := r.db.QueryRow(`
		SELECT id, point_id, student_id, content, file_url, file_name, submitted_at, review, review_status, reviewed_at, reviewed_by
		FROM point_results WHERE id = $1
	`, id).Scan(&res.ID, &res.PointID, &res.StudentID, &res.Content, &res.FileURL, &res.FileName, &res.SubmittedAt, &review, &res.ReviewStatus, &reviewedAt, &reviewedBy)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if review.Valid {
		res.Review = &review.String
	}
	if reviewedAt.Valid {
		rt := reviewedAt.Time
		res.ReviewedAt = &rt
	}
	if reviewedBy.Valid {
		rb := int(reviewedBy.Int64)
		res.ReviewedBy = &rb
	}
	return &res, err
}

func (r *ResultRepository) Update(res *models.PointResult) error {
	_, err := r.db.Exec(
		`UPDATE point_results SET content = $1 WHERE id = $2`,
		res.Content, res.ID,
	)
	return err
}

func (r *ResultRepository) Review(id int, review, status string, reviewedBy int) error {
	_, err := r.db.Exec(
		`UPDATE point_results SET review = $1, review_status = $2, reviewed_at = CURRENT_TIMESTAMP, reviewed_by = $3 WHERE id = $4`,
		review, status, reviewedBy, id,
	)
	return err
}

func (r *ResultRepository) Delete(id int) error {
	_, err := r.db.Exec(`DELETE FROM point_results WHERE id = $1`, id)
	return err
}
