# User API

User API — учебный HTTP API для управления пользователями. Проект написан на Go и хранит данные в PostgreSQL.

## Возможности

- CRUD HTTP API для управления пользователями.
- PostgreSQL storage и SQL-миграции.
- Валидация JSON-запросов и преобразование ошибок в HTTP-статусы.
- Unit-тесты HTTP handlers и integration-тесты PostgreSQL storage.
- Передача `context.Context` от HTTP-запроса до SQL-операций.
- GitHub Actions CI с PostgreSQL и автоматическим запуском тестов.
- Таймауты HTTP-сервера и graceful shutdown.

## Требования

- Go
- Docker
- Docker Compose

## Конфигурация

- `DATABASE_URL` — обязательная строка подключения к PostgreSQL;
- `SERVER_PORT` — необязательный порт сервера, по умолчанию используется `8080`;
- безопасный пример конфигурации находится в `.env.example`.

## Локальный запуск

```powershell
docker compose up -d
docker run --rm --mount "type=bind,source=${PWD}\migrations,target=/migrations,readonly" --network user-api_default migrate/migrate:v4.19.1 -path=/migrations -database "postgres://user:password@postgres:5432/user_api?sslmode=disable" up
$env:DATABASE_URL = "postgres://user:password@localhost:5432/user_api?sslmode=disable"
$env:SERVER_PORT = "8080"
go run ./cmd/user-api
```

## Эндпоинты API

| Method | Path | Назначение |
|---|---|---|
| `GET` | `/health` | Проверка работоспособности |
| `GET` | `/users` | Получение всех пользователей |
| `POST` | `/users` | Создание пользователя |
| `GET` | `/user?id=1` | Получение одного пользователя |
| `PUT` | `/user?id=1` | Изменение пользователя |
| `DELETE` | `/user?id=1` | Удаление пользователя |

## Тесты

```powershell
go test ./...
```

Команда запускает тесты, доступные без тестовой БД, а PostgreSQL-тесты без `TEST_DATABASE_URL` будут пропущены.

### Интеграционные тесты

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

## Непрерывная интеграция (CI)

GitHub Actions запускается при push в `main` и при создании Pull Request в `main`.

CI выполняет следующие шаги:

1. Загружает репозиторий.
2. Устанавливает версию Go из `go.mod`.
3. Запускает PostgreSQL service.
4. Применяет SQL-миграции.
5. Запускает все тесты командой `go test ./... -count=1`.

## Завершение сервера

Сервер обрабатывает `Ctrl+C` и `SIGTERM`. После получения сигнала он перестаёт принимать новые соединения и даёт активным запросам до 5 секунд на завершение.
