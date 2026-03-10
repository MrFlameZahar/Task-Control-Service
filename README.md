# Task Control Service

REST API сервис для управления задачами внутри команд.

Сервис позволяет создавать команды, приглашать пользователей, управлять
задачами, отслеживать историю изменений задач и получать аналитическую
информацию.

Проект реализован на **Go** с использованием **MySQL**, **Redis**,
**Docker Compose** и **Prometheus**.

------------------------------------------------------------------------

# Stack

-   Go
-   MySQL
-   Redis
-   Docker Compose
-   Prometheus
-   Chi router
-   sqlx
-   Zap logger
-   JWT authentication

------------------------------------------------------------------------

# Features

### Authentication

-   регистрация пользователей
-   login
-   JWT аутентификация
-   получение текущего пользователя

### Teams

-   создание команды
-   список команд пользователя
-   приглашение пользователей в команду
-   просмотр участников команды

### Tasks

-   создание задач
-   обновление задач
-   фильтрация задач
-   назначение исполнителя
-   история изменений задач

### Infrastructure

-   Redis cache для списка задач
-   Rate limiting
-   Prometheus metrics
-   Circuit breaker для mock email service

### Analytics

-   агрегированная статистика по командам
-   топ пользователей по созданию задач
-   проверка целостности данных

------------------------------------------------------------------------

# Architecture

Проект построен по принципам **layered architecture / ports &
adapters**.

### Слои

**domain** - доменные сущности - интерфейсы репозиториев - бизнес типы

**app** - бизнес логика - сервисы

**adapters** - MySQL repository - Redis cache - email adapter

**ports** - HTTP handlers - router - middleware

**cmd** - точка входа приложения

------------------------------------------------------------------------

# Project structure

    cmd/server

    internal/

      domain/
        user/
        team/
        task/
        analytics/

      app/
        auth/
        team/
        task/
        analytics/

      adapters/
        mysql/
        redis/
        email/

      ports/http/
        handlers/
        middleware/

      metrics/

    migrations/

    docker-compose.yml

------------------------------------------------------------------------

# Roles

  Role     Description
  -------- --------------------------------
  owner    создатель команды
  admin    может приглашать пользователей
  member   обычный участник

Право приглашать пользователей имеют **owner** и **admin**.

------------------------------------------------------------------------

# Run locally

## 1. Запуск инфраструктуры

``` bash
docker compose up -d --build
```

Это поднимет:

-   MySQL
-   Redis
-   Prometheus
-   Приложение

------------------------------------------------------------------------

## 2. API

После запуска сервис доступен по адресу:

    http://localhost:8081

------------------------------------------------------------------------

## 3. Prometheus

Prometheus UI:

    http://localhost:9090

Метрики приложения:

    http://localhost:8081/metrics

------------------------------------------------------------------------

# Configuration

Основные переменные окружения:

  Variable         Description
  ---------------- --------------------
  APP_PORT         порт приложения
  MYSQL_HOST       адрес MySQL
  MYSQL_PORT       порт MySQL
  MYSQL_USER       пользователь MySQL
  MYSQL_PASSWORD   пароль MySQL
  MYSQL_DB         база данных
  REDIS_HOST       Redis host
  REDIS_PORT       Redis port
  JWT_SECRET       секрет для JWT

------------------------------------------------------------------------

# Database schema

Основные таблицы:

    users
    teams
    team_members
    tasks
    task_history
    task_comments

Основные связи:

    teams.owner_id -> users.id

    team_members.user_id -> users.id
    team_members.team_id -> teams.id

    tasks.team_id -> teams.id
    tasks.assignee_id -> users.id
    tasks.created_by -> users.id

    task_history.task_id -> tasks.id

В качестве primary key используется **UUID**.

------------------------------------------------------------------------

# API

## Auth

    POST /api/v1/register
    POST /api/v1/login
    GET /api/v1/me

## Teams

    POST /api/v1/teams
    GET /api/v1/teams
    POST /api/v1/teams/{id}/invite
    GET /api/v1/teams/{id}/members

## Tasks

    POST /api/v1/tasks
    GET /api/v1/tasks
    PUT /api/v1/tasks/{id}
    GET /api/v1/tasks/{id}/history

## Analytics

    GET /api/v1/analytics/team-stats
    GET /api/v1/analytics/top-creators
    GET /api/v1/analytics/integrity

------------------------------------------------------------------------

# Examples

## Register

``` bash
curl -X POST http://localhost:8081/api/v1/register   -H "Content-Type: application/json"   -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

## Login

``` bash
curl -X POST http://localhost:8081/api/v1/login   -H "Content-Type: application/json"   -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

------------------------------------------------------------------------

# Caching strategy

Endpoint

    GET /api/v1/tasks

кешируется в Redis.

TTL кеша:

    5 minutes

Cache key зависит от:

-   team_id
-   status
-   assignee_id
-   limit
-   offset

При создании или обновлении задачи кеш команды инвалидируется.

------------------------------------------------------------------------

# Rate limiting

Для авторизованных пользователей реализован rate limiting:

    100 requests per minute per user

Используется **token bucket algorithm**.

Текущая реализация использует **in-memory limiter**.

Для multi-instance deployment limiter должен быть вынесен в Redis.

------------------------------------------------------------------------

# Metrics

Prometheus endpoint:

    /metrics

Экспортируются метрики:

    http_requests_total
    http_errors_total
    http_request_duration_seconds

------------------------------------------------------------------------

# Circuit breaker

Для mock email service используется **circuit breaker**.

Он предотвращает деградацию системы при отказе внешнего сервиса отправки
email.

Если email сервис недоступен:

-   пользователь всё равно добавляется в команду
-   ошибка отправки email логируется

------------------------------------------------------------------------

# Analytical SQL queries

### Team statistics

    GET /api/v1/analytics/team-stats

JOIN нескольких таблиц и агрегирование.

Возвращает:

-   название команды
-   количество участников
-   количество задач со статусом done за последние 7 дней

### Top task creators

    GET /api/v1/analytics/top-creators

Используется оконная функция:

    ROW_NUMBER() OVER (PARTITION BY team_id)

### Data integrity check

    GET /api/v1/analytics/integrity

Находит задачи, где:

    assignee не является участником команды задачи

------------------------------------------------------------------------

# Architecture decisions

### Layered architecture

Зависимости направлены следующим образом:

    HTTP -> app -> domain
               ^
               |
           adapters

Domain слой не зависит от инфраструктуры.

Это позволяет:

-   тестировать бизнес-логику без базы данных
-   легко заменять адаптеры
-   писать unit tests

------------------------------------------------------------------------

# Trade-offs

### In-memory rate limiting

Плюсы:

-   простота реализации
-   высокая производительность

Минусы:

-   не подходит для multi-instance deployment

Для production limiter должен использоваться Redis.

### Mock email service

Используется mock email сервис для демонстрации circuit breaker.

В production должен использоваться реальный email provider.

------------------------------------------------------------------------

# Tests

Запуск тестов:

``` bash
go test ./... -v
```

Покрыты тестами:

-   AuthService
-   TeamService
-   TaskService

------------------------------------------------------------------------

# Possible improvements

-   Redis-backed rate limiting
-   Swagger / OpenAPI documentation
-   integration tests с MySQL через testcontainers
-   distributed tracing (OpenTelemetry)
