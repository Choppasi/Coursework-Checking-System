package services

import (
	"thesis-app/internal/models"
	"thesis-app/internal/repository"
)

type NotificationService struct {
	repo *repository.NotificationRepository
}

func NewNotificationService(repo *repository.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) Notify(userID int, title, message string) error {
	n := &models.Notification{
		UserID:  userID,
		Title:   title,
		Message: message,
	}
	return s.repo.Create(n)
}

func (s *NotificationService) NotifyNewPoint(userID int, pointTitle string) error {
	return s.Notify(userID, "Новый пункт в плане", "Добавлен пункт: "+pointTitle)
}

func (s *NotificationService) NotifyReviewed(userID int, pointTitle string, status string) error {
	msg := "Результат по пункту \"" + pointTitle + "\" "
	if status == "approved" {
		msg += "одобрен"
	} else {
		msg += "отклонен"
	}
	return s.Notify(userID, "Результат проверен", msg)
}

func (s *NotificationService) NotifyDeadline(userID int, pointTitle string, days int) error {
	msg := "До дедлайна по пункту \"" + pointTitle + "\" осталось "
	if days == 1 {
		msg += "1 день"
	} else {
		msg += "3 дня"
	}
	return s.Notify(userID, "Приближается дедлайн", msg)
}
