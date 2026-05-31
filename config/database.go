package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

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

	var exists bool
	err = dbDefault.QueryRow("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = $1)", cfg.DBName).Scan(&exists)
	if err != nil {
		dbDefault.Close()
		return nil, fmt.Errorf("не могу проверить базу: %w", err)
	}

	if !exists {
		_, err = dbDefault.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.DBName))
		if err != nil {
			dbDefault.Close()
			return nil, fmt.Errorf("не могу создать базу: %w", err)
		}
		log.Printf("База '%s' создана", cfg.DBName)
	}
	dbDefault.Close()

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

	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("ошибка миграций: %w", err)
	}

	log.Println("База данных готова!")
	return db, nil
}

func runMigrations(db *sql.DB) error {
	files, err := os.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("не могу прочитать папку миграций: %w", err)
	}

	var sqlFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			sqlFiles = append(sqlFiles, f.Name())
		}
	}
	sort.Strings(sqlFiles)

	for _, name := range sqlFiles {
		path := filepath.Join("migrations", name)
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("не могу прочитать %s: %w", name, err)
		}
		_, err = db.Exec(string(data))
		if err != nil {
			return fmt.Errorf("ошибка выполнения %s: %w", name, err)
		}
		log.Printf("Миграция применена: %s", name)
	}

	return nil
}
