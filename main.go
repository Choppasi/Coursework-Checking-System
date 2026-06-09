package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"thesis-app/config"
	"thesis-app/internal/handlers"
	"thesis-app/internal/middleware"
	"thesis-app/internal/models"
	"thesis-app/internal/repository"
	"thesis-app/internal/services"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	// Настройка логирования в файл
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Не удалось открыть файл логов: %v", err)
	}
	defer logFile.Close()

	// Пишем логи и в файл, и в консоль
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)

	cfg := config.Load()

	db, err := config.InitDB(cfg)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе: %v", err)
	}
	defer db.Close()

	_ = os.MkdirAll("./uploads", 0755)

	// Репозитории
	userRepo := repository.NewUserRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	thesisRepo := repository.NewThesisRepository(db)
	pointRepo := repository.NewPointRepository(db)
	resultRepo := repository.NewResultRepository(db)
	notifRepo := repository.NewNotificationRepository(db)

	// Сервисы
	notifService := services.NewNotificationService(notifRepo)
	fileService := services.NewFileService("./uploads")

	// Хендлеры
	authHandler := handlers.NewAuthHandler(cfg, userRepo)
	groupHandler := handlers.NewGroupHandler(groupRepo)
	thesisHandler := handlers.NewThesisHandler(thesisRepo)
	pointHandler := handlers.NewPointHandler(pointRepo, thesisRepo, notifService)
	resultHandler := handlers.NewResultHandler(resultRepo, pointRepo, thesisRepo, notifService, fileService)
	notifHandler := handlers.NewNotificationHandler(notifRepo)

	// Проверяем JWT secret
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET не задан! Установите переменную окружения JWT_SECRET")
	}

	// Middleware
	authMW := middleware.AuthMiddleware(cfg)
	requireTeacher := middleware.RequireRole("teacher", "admin")
	requireStudent := middleware.RequireRole("student", "teacher", "admin")

	router := mux.NewRouter()

	// Статика (публичная)
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// Загруженные файлы — только для авторизованных
	uploads := router.PathPrefix("/uploads/").Subrouter()
	uploads.Use(authMW)
	uploads.PathPrefix("/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	// Главная
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	// Health
	router.HandleFunc("/health", handlers.HealthCheck).Methods("GET")

	// API публичное
	authHandler.RegisterRoutes(router)

	// API защищенное
	api := router.PathPrefix("/api").Subrouter()
	api.Use(authMW)

	// Группы
	api.HandleFunc("/groups", groupHandler.GetAll).Methods("GET")
	api.Handle("/groups", requireTeacher(http.HandlerFunc(groupHandler.Create))).Methods("POST")
	api.HandleFunc("/groups/{id}", groupHandler.GetByID).Methods("GET")
	api.Handle("/groups/{id}", requireTeacher(http.HandlerFunc(groupHandler.Update))).Methods("PUT")
	api.Handle("/groups/{id}", requireTeacher(http.HandlerFunc(groupHandler.Delete))).Methods("DELETE")
	api.Handle("/groups/{id}/members", http.HandlerFunc(groupHandler.AddMember)).Methods("POST")
	api.Handle("/groups/{id}/members/{studentId}", http.HandlerFunc(groupHandler.RemoveMember)).Methods("DELETE")
	api.Handle("/groups/my", requireStudent(http.HandlerFunc(groupHandler.GetMyGroup))).Methods("GET")
	api.Handle("/groups/{id}/join", requireStudent(http.HandlerFunc(groupHandler.JoinGroup))).Methods("POST")

	// Курсовые
	thesisHandler.RegisterRoutes(api)

	// Пункты
	pointHandler.RegisterRoutes(api)

	// Результаты
	resultHandler.RegisterRoutes(api)

	// Уведомления
	notifHandler.RegisterRoutes(api)

	// Дополнительно: список пользователей для админов и преподавателей
	api.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		role := middleware.GetUserRole(r)
		if role != "admin" && role != "teacher" {
			http.Error(w, `{"error":"Forbidden"}`, http.StatusForbidden)
			return
		}
		users, err := userRepo.GetAll()
		if err != nil {
			http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
			return
		}
		if users == nil {
			users = []models.User{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}).Methods("GET")

	api.HandleFunc("/users/teachers", func(w http.ResponseWriter, r *http.Request) {
		users, err := userRepo.GetByRole("teacher")
		if err != nil {
			http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
			return
		}
		if users == nil {
			users = []models.User{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}).Methods("GET")

	api.HandleFunc("/users/students", func(w http.ResponseWriter, r *http.Request) {
		role := middleware.GetUserRole(r)
		if role != "admin" && role != "teacher" {
			http.Error(w, `{"error":"Forbidden"}`, http.StatusForbidden)
			return
		}
		users, err := userRepo.GetByRole("student")
		if err != nil {
			http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
			return
		}
		if users == nil {
			users = []models.User{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}).Methods("GET")

	api.HandleFunc("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		role := middleware.GetUserRole(r)
		if role != "admin" {
			http.Error(w, `{"error":"Forbidden"}`, http.StatusForbidden)
			return
		}
		id, _ := strconv.Atoi(mux.Vars(r)["id"])
		if id == middleware.GetUserID(r) {
			http.Error(w, `{"error":"Cannot delete yourself"}`, http.StatusBadRequest)
			return
		}
		if err := userRepo.Delete(id); err != nil {
			http.Error(w, `{"error":"Could not delete user"}`, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}).Methods("DELETE")

	addr := ":" + cfg.ServerPort
	log.Printf("Сервер запущен: http://localhost:%s", cfg.ServerPort)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Не могу запустить сервер: %v", err)
	}
}
