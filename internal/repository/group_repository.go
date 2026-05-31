package repository

import (
	"database/sql"
	"thesis-app/internal/models"
)

type GroupRepository struct {
	db *sql.DB
}

func NewGroupRepository(db *sql.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

func (r *GroupRepository) Create(g *models.Group) error {
	return r.db.QueryRow(
		`INSERT INTO groups (name, teacher_id, course, year) VALUES ($1, $2, $3, $4) RETURNING id, created_at`,
		g.Name, g.TeacherID, g.Course, g.Year,
	).Scan(&g.ID, &g.CreatedAt)
}

func (r *GroupRepository) GetAll() ([]models.GroupWithTeacher, error) {
	rows, err := r.db.Query(`
		SELECT g.id, g.name, g.teacher_id, g.course, g.year, g.created_at, u.full_name
		FROM groups g
		LEFT JOIN users u ON u.id = g.teacher_id
		ORDER BY g.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.GroupWithTeacher
	for rows.Next() {
		var g models.GroupWithTeacher
		if err := rows.Scan(&g.ID, &g.Name, &g.TeacherID, &g.Course, &g.Year, &g.CreatedAt, &g.TeacherName); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, nil
}

func (r *GroupRepository) GetByID(id int) (*models.GroupWithTeacher, error) {
	var g models.GroupWithTeacher
	err := r.db.QueryRow(`
		SELECT g.id, g.name, g.teacher_id, g.course, g.year, g.created_at, u.full_name
		FROM groups g
		LEFT JOIN users u ON u.id = g.teacher_id
		WHERE g.id = $1
	`, id).Scan(&g.ID, &g.Name, &g.TeacherID, &g.Course, &g.Year, &g.CreatedAt, &g.TeacherName)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &g, err
}

func (r *GroupRepository) Update(g *models.Group) error {
	_, err := r.db.Exec(
		`UPDATE groups SET name = $1, teacher_id = $2, course = $3, year = $4 WHERE id = $5`,
		g.Name, g.TeacherID, g.Course, g.Year, g.ID,
	)
	return err
}

func (r *GroupRepository) Delete(id int) error {
	_, err := r.db.Exec(`DELETE FROM groups WHERE id = $1`, id)
	return err
}

func (r *GroupRepository) AddMember(groupID, studentID int) error {
	_, err := r.db.Exec(
		`INSERT INTO group_members (group_id, student_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		groupID, studentID,
	)
	return err
}

func (r *GroupRepository) RemoveMember(groupID, studentID int) error {
	_, err := r.db.Exec(
		`DELETE FROM group_members WHERE group_id = $1 AND student_id = $2`,
		groupID, studentID,
	)
	return err
}

func (r *GroupRepository) GetMembers(groupID int) ([]models.User, error) {
	rows, err := r.db.Query(`
		SELECT u.id, u.email, u.role, u.full_name, u.created_at, u.updated_at
		FROM users u
		JOIN group_members gm ON gm.student_id = u.id
		WHERE gm.group_id = $1
		ORDER BY u.full_name
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Role, &u.FullName, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *GroupRepository) GetStudentGroup(studentID int) (*models.Group, error) {
	var g models.Group
	err := r.db.QueryRow(`
		SELECT g.id, g.name, g.teacher_id, g.course, g.year, g.created_at
		FROM groups g
		JOIN group_members gm ON gm.group_id = g.id
		WHERE gm.student_id = $1
	`, studentID).Scan(&g.ID, &g.Name, &g.TeacherID, &g.Course, &g.Year, &g.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &g, err
}
