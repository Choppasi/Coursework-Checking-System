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
	if t.GroupID != nil {
		return r.db.QueryRow(
			`INSERT INTO theses (student_id, group_id, title, description, status, start_date, deadline)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, created_at, updated_at`,
			t.StudentID, t.GroupID, t.Title, t.Description, t.Status, t.StartDate, t.Deadline,
		).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	}
	return r.db.QueryRow(
		`INSERT INTO theses (student_id, title, description, status, start_date, deadline)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`,
		t.StudentID, t.Title, t.Description, t.Status, t.StartDate, t.Deadline,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (r *ThesisRepository) GetByID(id int) (*models.ThesisWithStudent, error) {
	var t models.ThesisWithStudent
	var groupID sql.NullInt64
	err := r.db.QueryRow(`
		SELECT th.id, th.student_id, th.group_id, th.title, th.description, th.status, th.start_date, th.deadline, th.created_at, th.updated_at, u.full_name
		FROM theses th
		LEFT JOIN users u ON u.id = th.student_id
		WHERE th.id = $1
	`, id).Scan(&t.ID, &t.StudentID, &groupID, &t.Title, &t.Description, &t.Status, &t.StartDate, &t.Deadline, &t.CreatedAt, &t.UpdatedAt, &t.StudentName)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if groupID.Valid {
		gid := int(groupID.Int64)
		t.GroupID = &gid
	}
	return &t, err
}

func (r *ThesisRepository) GetByStudent(studentID int) ([]models.ThesisWithStudent, error) {
	rows, err := r.db.Query(`
		SELECT th.id, th.student_id, th.group_id, th.title, th.description, th.status, th.start_date, th.deadline, th.created_at, th.updated_at, u.full_name
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
		var groupID sql.NullInt64
		if err := rows.Scan(&t.ID, &t.StudentID, &groupID, &t.Title, &t.Description, &t.Status, &t.StartDate, &t.Deadline, &t.CreatedAt, &t.UpdatedAt, &t.StudentName); err != nil {
			return nil, err
		}
		if groupID.Valid {
			gid := int(groupID.Int64)
			t.GroupID = &gid
		}
		list = append(list, t)
	}
	return list, nil
}

func (r *ThesisRepository) GetAll() ([]models.ThesisWithStudent, error) {
	rows, err := r.db.Query(`
		SELECT th.id, th.student_id, th.group_id, th.title, th.description, th.status, th.start_date, th.deadline, th.created_at, th.updated_at, u.full_name
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
		var groupID sql.NullInt64
		if err := rows.Scan(&t.ID, &t.StudentID, &groupID, &t.Title, &t.Description, &t.Status, &t.StartDate, &t.Deadline, &t.CreatedAt, &t.UpdatedAt, &t.StudentName); err != nil {
			return nil, err
		}
		if groupID.Valid {
			gid := int(groupID.Int64)
			t.GroupID = &gid
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

func (r *ThesisRepository) UpdateStatus(id int, status string) error {
	_, err := r.db.Exec(
		`UPDATE theses SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`,
		status, id,
	)
	return err
}

func (r *ThesisRepository) UpdateStatusFromPoints(thesisID int) error {
	// Получаем все пункты для курсовой
	rows, err := r.db.Query(`
		SELECT status FROM thesis_points WHERE thesis_id = $1 ORDER BY "order"
	`, thesisID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var statuses []string
	for rows.Next() {
		var status string
		if err := rows.Scan(&status); err != nil {
			return err
		}
		statuses = append(statuses, status)
	}

	if len(statuses) == 0 {
		// Нет пунктов - возвращаемся
		return nil
	}

	// Определяем новый статус курсовой
	newStatus := "planning"

	// Если есть rejected - курсовая rejected
	for _, s := range statuses {
		if s == "rejected" {
			newStatus = "rejected"
			break
		}
	}

	// Если нет rejected, но есть in_progress - курсовая in_progress
	if newStatus == "planning" {
		for _, s := range statuses {
			if s == "in_progress" {
				newStatus = "in_progress"
				break
			}
		}
	}

	// Если все пункты done - курсовая completed
	if newStatus == "planning" {
		allDone := true
		for _, s := range statuses {
			if s != "done" {
				allDone = false
				break
			}
		}
		if allDone {
			newStatus = "completed"
		}
	}

	return r.UpdateStatus(thesisID, newStatus)
}

func (r *ThesisRepository) Delete(id int) error {
	_, err := r.db.Exec(`DELETE FROM theses WHERE id = $1`, id)
	return err
}

func (r *ThesisRepository) GetStudentsByGroup(groupID int) ([]int, error) {
	rows, err := r.db.Query(`
		SELECT student_id FROM group_members WHERE group_id = $1
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var studentIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		studentIDs = append(studentIDs, id)
	}
	return studentIDs, nil
}

func (r *ThesisRepository) CreateForGroup(groupID int, title, description, status, startDate, deadline string) ([]int, error) {
	studentIDs, err := r.GetStudentsByGroup(groupID)
	if err != nil {
		return nil, err
	}

	var createdIDs []int
	var startVal, deadlineVal *string

	if startDate != "" {
		startVal = &startDate
	}
	if deadline != "" {
		deadlineVal = &deadline
	}

	for _, sid := range studentIDs {
		t := &models.Thesis{
			StudentID:   sid,
			GroupID:     &groupID,
			Title:       title,
			Description: description,
			Status:      status,
			StartDate:   startVal,
			Deadline:    deadlineVal,
		}
		if err := r.Create(t); err != nil {
			return nil, err
		}
		createdIDs = append(createdIDs, t.ID)
	}

	return createdIDs, nil
}
