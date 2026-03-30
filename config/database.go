package config

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func InitDB(cfg *Config) (*sql.DB, error) {
	// Подключаемся к системной БД postgres, чтобы создать нашу
	connStrDefault := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword,
	)

	dbDefault, err := sql.Open("postgres", connStrDefault)
	if err != nil {
		return nil, fmt.Errorf("не могу подключиться к postgres: %w", err)
	}

	// Проверяем, есть ли наша база данных
	var exists bool
	err = dbDefault.QueryRow("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = $1)", cfg.DBName).Scan(&exists)
	if err != nil {
		dbDefault.Close()
		return nil, fmt.Errorf("не могу проверить базу: %w", err)
	}

	// Создаём базу, если её нет
	if !exists {
		_, err = dbDefault.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.DBName))
		if err != nil {
			dbDefault.Close()
			return nil, fmt.Errorf("не могу создать базу: %w", err)
		}
		log.Printf("База '%s' создана", cfg.DBName)
	}
	dbDefault.Close()

	// Подключаемся к нашей базе
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("не могу подключиться к базе: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("не могу пингануть базу: %w", err)
	}

	// Создаём таблицу задач
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS todos (
			id SERIAL PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			completed BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("не могу создать таблицу: %w", err)
	}

	log.Println("База данных готова!")
	return db, nil
}
