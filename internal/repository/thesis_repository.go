package repository

import (
	"database/sql"
	"thesis-app/internal/models"
)

type ThesisRepository struct {
	db *sql.DB
}

func NewThesisRepository(db *sql.DB) *ThesisRepository {
	return &ThesisRepository{db: db}
}

func (r *ThesisRepository) Create(t *models.Thesis) error {
	return r.db.QueryRow(
		`INSERT INTO theses (student_id, title, description, status, start_date, deadline)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`,
		t.StudentID, t.Title, t.Description, t.Status, t.StartDate, t.Deadline,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (r *ThesisRepository) GetByID(id int) (*models.ThesisWithStudent, error) {
	var t models.ThesisWithStudent
	err := r.db.QueryRow(`
		SELECT th.id, th.student_id, th.title, th.description, th.status, th.start_date, th.deadline, th.created_at, th.updated_at, u.full_name
		FROM theses th
		LEFT JOIN users u ON u.id = th.student_id
		WHERE th.id = $1
	`, id).Scan(&t.ID, &t.StudentID, &t.Title, &t.Description, &t.Status, &t.StartDate, &t.Deadline, &t.CreatedAt, &t.UpdatedAt, &t.StudentName)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &t, err
}

func (r *ThesisRepository) GetByStudent(studentID int) ([]models.ThesisWithStudent, error) {
	rows, err := r.db.Query(`
		SELECT th.id, th.student_id, th.title, th.description, th.status, th.start_date, th.deadline, th.created_at, th.updated_at, u.full_name
		FROM theses th
		LEFT JOIN users u ON u.id = th.student_id
		WHERE th.student_id = $1
		ORDER BY th.created_at DESC
	`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.ThesisWithStudent
	for rows.Next() {
		var t models.ThesisWithStudent
		if err := rows.Scan(&t.ID, &t.StudentID, &t.Title, &t.Description, &t.Status, &t.StartDate, &t.Deadline, &t.CreatedAt, &t.UpdatedAt, &t.StudentName); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, nil
}

func (r *ThesisRepository) GetAll() ([]models.ThesisWithStudent, error) {
	rows, err := r.db.Query(`
		SELECT th.id, th.student_id, th.title, th.description, th.status, th.start_date, th.deadline, th.created_at, th.updated_at, u.full_name
		FROM theses th
		LEFT JOIN users u ON u.id = th.student_id
		ORDER BY th.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.ThesisWithStudent
	for rows.Next() {
		var t models.ThesisWithStudent
		if err := rows.Scan(&t.ID, &t.StudentID, &t.Title, &t.Description, &t.Status, &t.StartDate, &t.Deadline, &t.CreatedAt, &t.UpdatedAt, &t.StudentName); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, nil
}

func (r *ThesisRepository) Update(t *models.Thesis) error {
	_, err := r.db.Exec(
		`UPDATE theses SET title = $1, description = $2, status = $3, start_date = $4, deadline = $5, updated_at = CURRENT_TIMESTAMP WHERE id = $6`,
		t.Title, t.Description, t.Status, t.StartDate, t.Deadline, t.ID,
	)
	return err
}

func (r *ThesisRepository) Delete(id int) error {
	_, err := r.db.Exec(`DELETE FROM theses WHERE id = $1`, id)
	return err
}
