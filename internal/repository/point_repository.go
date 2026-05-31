package repository

import (
	"database/sql"
	"thesis-app/internal/models"
)

type PointRepository struct {
	db *sql.DB
}

func NewPointRepository(db *sql.DB) *PointRepository {
	return &PointRepository{db: db}
}

func (r *PointRepository) Create(p *models.ThesisPoint) error {
	return r.db.QueryRow(
		`INSERT INTO thesis_points (thesis_id, title, description, "order", deadline, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`,
		p.ThesisID, p.Title, p.Description, p.Order, p.Deadline, p.Status,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *PointRepository) GetByID(id int) (*models.ThesisPoint, error) {
	var p models.ThesisPoint
	err := r.db.QueryRow(`
		SELECT id, thesis_id, title, description, "order", deadline, status, created_at, updated_at
		FROM thesis_points WHERE id = $1
	`, id).Scan(&p.ID, &p.ThesisID, &p.Title, &p.Description, &p.Order, &p.Deadline, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

func (r *PointRepository) GetByThesis(thesisID int) ([]models.ThesisPoint, error) {
	rows, err := r.db.Query(`
		SELECT id, thesis_id, title, description, "order", deadline, status, created_at, updated_at
		FROM thesis_points WHERE thesis_id = $1 ORDER BY "order", id
	`, thesisID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.ThesisPoint
	for rows.Next() {
		var p models.ThesisPoint
		if err := rows.Scan(&p.ID, &p.ThesisID, &p.Title, &p.Description, &p.Order, &p.Deadline, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func (r *PointRepository) Update(p *models.ThesisPoint) error {
	_, err := r.db.Exec(
		`UPDATE thesis_points SET title = $1, description = $2, "order" = $3, deadline = $4, status = $5, updated_at = CURRENT_TIMESTAMP WHERE id = $6`,
		p.Title, p.Description, p.Order, p.Deadline, p.Status, p.ID,
	)
	return err
}

func (r *PointRepository) UpdateStatus(id int, status string) error {
	_, err := r.db.Exec(
		`UPDATE thesis_points SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`,
		status, id,
	)
	return err
}

func (r *PointRepository) Delete(id int) error {
	_, err := r.db.Exec(`DELETE FROM thesis_points WHERE id = $1`, id)
	return err
}

func (r *PointRepository) DeleteByThesis(thesisID int) error {
	_, err := r.db.Exec(`DELETE FROM thesis_points WHERE thesis_id = $1`, thesisID)
	return err
}
