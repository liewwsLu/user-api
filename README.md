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
docker run --rm --mount "type=bind,source=${PWD}\migrations,target=/migrations,readonly" --network user-api_default migrate/migrate:v4.19.1 -path=/migrations -database "postgres://user:password@postgres:5432/user_api?sslmode=disable" up
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

## Tests

```powershell
go test ./...
```

Команда запускает тесты, доступные без тестовой БД, а PostgreSQL-тесты без `TEST_DATABASE_URL` будут пропущены.

### Integration tests

Для запуска integration-тестов нужны:

- запущенный PostgreSQL;
- применённые миграции;
- переменная окружения `TEST_DATABASE_URL`.

При первом запуске создайте отдельную тестовую базу:

```powershell
docker compose up -d
docker exec user-api-postgres psql -U user -d postgres -c 'CREATE DATABASE user_api_test OWNER "user";'
```

Примените миграции и запустите тесты:

```powershell
docker run --rm --mount "type=bind,source=${PWD}\migrations,target=/migrations,readonly" --network user-api_default migrate/migrate:v4.19.1 -path=/migrations -database "postgres://user:password@postgres:5432/user_api_test?sslmode=disable" up
$env:TEST_DATABASE_URL = "postgres://user:password@localhost:5432/user_api_test?sslmode=disable"
go test ./... -count=1
```

Разбор:

- `docker compose up -d` запускает локальный PostgreSQL;
- `CREATE DATABASE` выполняется один раз, пока существует Docker volume;
- `"user"` взят в кавычки, потому что `user` имеет специальное значение в PostgreSQL;
- миграции направлены в `user_api_test`, а не в рабочую `user_api`;
- `TEST_DATABASE_URL` доступна только текущему PowerShell;
- `-count=1` заставляет integration-тесты реально выполниться;
- `DATABASE_URL` и `SERVER_PORT` здесь не нужны, потому что HTTP-сервер не запускается.
