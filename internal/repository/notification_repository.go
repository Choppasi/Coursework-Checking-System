package repository

import (
	"database/sql"
	"thesis-app/internal/models"
)

type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(n *models.Notification) error {
	return r.db.QueryRow(
		`INSERT INTO notifications (user_id, title, message) VALUES ($1, $2, $3) RETURNING id, created_at`,
		n.UserID, n.Title, n.Message,
	).Scan(&n.ID, &n.CreatedAt)
}

func (r *NotificationRepository) GetByUser(userID int) ([]models.Notification, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, title, message, is_read, created_at
		FROM notifications WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Notification
	for rows.Next() {
		var n models.Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Message, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, n)
	}
	return list, nil
}

func (r *NotificationRepository) GetUnread(userID int) ([]models.Notification, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, title, message, is_read, created_at
		FROM notifications WHERE user_id = $1 AND is_read = FALSE ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Notification
	for rows.Next() {
		var n models.Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Message, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, n)
	}
	return list, nil
}

func (r *NotificationRepository) MarkRead(id int) error {
	_, err := r.db.Exec(`UPDATE notifications SET is_read = TRUE WHERE id = $1`, id)
	return err
}

func (r *NotificationRepository) MarkAllRead(userID int) error {
	_, err := r.db.Exec(`UPDATE notifications SET is_read = TRUE WHERE user_id = $1`, userID)
	return err
}
