# User API

User API — учебный HTTP API для управления пользователями. Проект написан на Go и хранит данные в PostgreSQL.

## Requirements:
 
 - Go
 - Docker 
 - Docker Compose

## Configuration:
 
 - `DATABASE_URL` - обязательная строка подключения к PostgreSQL;
 - `SERVER_PORT` - необязательный порт, поскольку по умолчанию 8080;
 - `безопасный пример` находится в `.env.example`.

## Running locally:

```powershell
docker compose up -d
$env:DATABASE_URL = "postgres://user:password@localhost:5432/user_api?sslmode=disable"
$env:SERVER_PORT = "8080"
go run ./cmd/user-api
```

## API endpoints

| Method | Path | Назначение |
|---|---|---|
| `GET` | `/health` | Проверка работоспособности |
| `GET` | `/users` | Получение всех пользователей |
| `POST` | `/users` | Создание пользователя |
| `GET` | `/user?id=1` | Получение одного пользователя |
| `PUT` | `/user?id=1` | Изменение пользователя |
| `DELETE` | `/user?id=1` | Удаление пользователя |

## Tests:

```powershell
go test ./...
```