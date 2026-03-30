package main

import (
	"log"
	"net/http"

	"todo-app/config"
	"todo-app/internal/handlers"
	"todo-app/internal/repository"

	"github.com/gorilla/mux"
)

func main() {
	// Загружаем настройки
	cfg := config.Load()

	// Подключаемся к БД (сама создастся если нет)
	db, err := config.InitDB(cfg)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе: %v", err)
	}
	defer db.Close()

	// Создаём обработчики
	todoRepo := repository.NewTodoRepository(db)
	todoHandler := handlers.NewTodoHandler(todoRepo)

	// Настраиваем маршруты
	router := mux.NewRouter()

	// CSS и JS файлы
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// Главная страница
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	// Простые эндпоинты для проверки
	router.HandleFunc("/health", handlers.HealthCheck).Methods(http.MethodGet)
	router.HandleFunc("/ping", handlers.Ping).Methods(http.MethodGet)

	// API для задач
	todoHandler.RegisterRoutes(router)

	// Запускаем сервер
	addr := ":" + cfg.ServerPort
	log.Printf("Сервер запущен: http://localhost:%s", cfg.ServerPort)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Не могу запустить сервер: %v", err)
	}
}
