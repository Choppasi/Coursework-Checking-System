package repository

import (
	"database/sql"
	"time"

	"todo-app/internal/models"
)

type TodoRepository struct {
	db *sql.DB
}

func NewTodoRepository(db *sql.DB) *TodoRepository {
	return &TodoRepository{db: db}
}

func (r *TodoRepository) GetAll() ([]models.Todo, error) {
	rows, err := r.db.Query(`
		SELECT id, title, description, completed, created_at, updated_at 
		FROM todos 
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []models.Todo
	for rows.Next() {
		var todo models.Todo
		err := rows.Scan(&todo.ID, &todo.Title, &todo.Description, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	return todos, nil
}

func (r *TodoRepository) GetByID(id int) (*models.Todo, error) {
	var todo models.Todo
	err := r.db.QueryRow(`
		SELECT id, title, description, completed, created_at, updated_at 
		FROM todos 
		WHERE id = $1
	`, id).Scan(&todo.ID, &todo.Title, &todo.Description, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &todo, nil
}

func (r *TodoRepository) Create(todo *models.Todo) error {
	err := r.db.QueryRow(`
		INSERT INTO todos (title, description, completed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, todo.Title, todo.Description, todo.Completed, todo.CreatedAt, todo.UpdatedAt).Scan(&todo.ID)

	return err
}

func (r *TodoRepository) Update(todo *models.Todo) error {
	todo.UpdatedAt = time.Now()
	_, err := r.db.Exec(`
		UPDATE todos 
		SET title = $1, description = $2, completed = $3, updated_at = $4
		WHERE id = $5
	`, todo.Title, todo.Description, todo.Completed, todo.UpdatedAt, todo.ID)

	return err
}

func (r *TodoRepository) Delete(id int) error {
	_, err := r.db.Exec(`DELETE FROM todos WHERE id = $1`, id)
	return err
}
