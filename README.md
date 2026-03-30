# Todo API
## README сделан с помощью ИИ
REST API для управления задачами на Go с PostgreSQL.
"Просто тест комитта"
## Быстрый старт

```bash
go run main.go
```

Откройте в браузере: **http://localhost:8080**

Всё остальное (создание БД, таблиц) происходит автоматически!

## API Endpoints

| Метод | Путь | Описание |
|-------|------|----------|
| GET | /health | Проверка здоровья |
| GET | /ping | Пинг |
| GET | /todos | Список всех задач |
| GET | /todos/{id} | Получить задачу по ID |
| POST | /todos | Создать задачу |
| PUT | /todos/{id} | Обновить задачу |
| DELETE | /todos/{id} | Удалить задачу |

## Примеры использования

### Создать задачу
```bash
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{"title": "Изучить Go", "description": "Выучить основы Go"}'
```

### Получить все задачи
```bash
curl http://localhost:8080/todos
```

### Обновить задачу
```bash
curl -X PUT http://localhost:8080/todos/1 \
  -H "Content-Type: application/json" \
  -d '{"completed": true}'
```

### Удалить задачу
```bash
curl -X DELETE http://localhost:8080/todos/1
```
