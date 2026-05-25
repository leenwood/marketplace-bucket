# marketplace-bucket

[🇬🇧 English](README.md)

Производственный Go-микросервис для управления корзиной покупок на базе Redis.
Демонстрирует практические паттерны чистой архитектуры, наблюдаемости и распределённых систем.

---

## Что демонстрирует этот репозиторий

- **Чистая архитектура** — четыре строгих слоя (domain → port → usecase → infra/app), контролируемых `go-arch-lint`
- **Трассировка OpenTelemetry** — распространение распределённых трейсов через HTTP-обработчики и операции Redis
- **Метрики Prometheus** — RED-метрики и счётчики операций над корзиной через изолированный реестр
- **Структурированные логи** — JSON-логи с `request_id` и `trace_id` в каждой строке
- **Хранилище на Redis** — корзина хранится как JSON-блоб с настраиваемым TTL (по умолчанию 7 дней)
- **Graceful shutdown** — обработка активных запросов до закрытия Redis и сброса трейсов OTel

---

## Обзор архитектуры

```
cmd/server/main.go
        │
        ▼
internal/app/service/server.go      ← сборка зависимостей, запуск HTTP, обработка завершения
        │
        ▼
internal/app/http/server.go         ← NewServer: маршруты + цепочка middleware
        │
        ├── middleware: otelhttp → Recover → Logger → RequestID → MaxBodySize
        │
        ├── POST   /api/v1/cart/{userID}/items
        ├── GET    /api/v1/cart/{userID}
        ├── PATCH  /api/v1/cart/{userID}/items/{productID}
        ├── DELETE /api/v1/cart/{userID}/items/{productID}
        ├── DELETE /api/v1/cart/{userID}
        │
        ├── GET /health   GET /ready   GET /metrics
        └── GET /swagger/   GET /debug/pprof/*

        │
        ▼
internal/core/usecase/cart.go       ← CartUseCase: бизнес-логика
        │
        ▼
internal/infra/storage/redis/       ← CartRepository: ключ Redis cart:{userID}
```

---

## Ключевые архитектурные решения

| Область | Решение |
|---|---|
| Хранилище | Redis ключ `cart:{user_id}` как JSON-блоб; TTL 7 дней |
| Транспорт | Только HTTP (REST/JSON) |
| Конфигурация | Чистые env-переменные — без Viper и сторонних библиотек |
| Наблюдаемость | Prometheus на `/metrics` (изолированный реестр), OTel OTLP HTTP или stdout |
| Контроль архитектуры | `go-arch-lint` ломает сборку при нарушении границ слоёв |
| Тесты | Unit-тесты мокируют репозиторий через интерфейс; интеграционные тесты используют `miniredis` |

---

## Наблюдаемость

### Логи
Все логи в формате JSON. Каждая строка запроса содержит:
```json
{"time":"...","level":"INFO","msg":"request","method":"POST","path":"/api/v1/cart/u1/items",
 "status":200,"latency":"1.2ms","request_id":"uuid","trace_id":"otel-trace-id"}
```

### Трейсы
Спаны OpenTelemetry покрывают:
- HTTP-обработчик (корневой спан через `otelhttp`)
- Операции Redis: Get / Set / Del

### Метрики
| Метрика | Тип | Лейблы |
|---|---|---|
| `platform_http_requests_total` | Counter | method, path, status |
| `platform_http_request_duration_seconds` | Histogram | method, path |
| `cart_operations_total` | Counter | operation |

---

## Локальный запуск

**Требования:** Docker, Docker Compose, Go 1.22+, [go-task](https://taskfile.dev/installation/)

```bash
# Установить go-task (один раз)
go install github.com/go-task/task/v3/cmd/task@latest

# 1. Клонировать и перейти в директорию
git clone https://github.com/leenwood/marketplace-bucket
cd marketplace-bucket

# 2. Запустить инфраструктуру (Redis + Jaeger)
task docker:up

# 3. Запустить сервер
task run
```

Доступные сервисы:

| Сервис | URL |
|---|---|
| API-сервер | http://localhost:8080 |
| Swagger UI | http://localhost:8080/swagger/index.html |
| Метрики Prometheus | http://localhost:8080/metrics |
| Jaeger UI | http://localhost:16686 |

---

## Переменные окружения

| Переменная | По умолчанию | Описание |
|---|---|---|
| `HTTP_ADDR` | `:8080` | Адрес HTTP-сервера |
| `HTTP_READ_TIMEOUT` | `15s` | Таймаут чтения |
| `HTTP_WRITE_TIMEOUT` | `15s` | Таймаут записи |
| `HTTP_PPROF_ENABLED` | `false` | Включить `/debug/pprof/*` |
| `REDIS_ADDR` | `localhost:6379` | Адрес Redis |
| `REDIS_PASSWORD` | `` | Пароль Redis |
| `REDIS_DB` | `0` | Номер базы данных Redis |
| `CART_TTL` | `168h` | TTL корзины (7 дней) |
| `LOG_LEVEL` | `info` | Уровень логирования (debug/info/warn/error) |
| `LOG_FORMAT` | `json` | Формат логов (json/text) |
| `OTEL_ENABLED` | `false` | Включить OpenTelemetry |
| `OTEL_EXPORTER` | `stdout` | Тип экспортёра (stdout/otlp) |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `localhost:4318` | Endpoint OTLP HTTP |
| `OTEL_SERVICE_NAME` | `marketplace-bucket` | Имя сервиса в трейсах |

---

## Примеры API

```bash
# Проверка работоспособности
curl http://localhost:8080/health

# Проверка готовности (пингует Redis)
curl http://localhost:8080/ready

# Добавить товар в корзину
curl -X POST http://localhost:8080/api/v1/cart/user-123/items \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "prod-42",
    "name": "Беспроводные наушники",
    "price": 59.99,
    "quantity": 2
  }'

# Получить корзину
curl http://localhost:8080/api/v1/cart/user-123

# Обновить количество товара
curl -X PATCH http://localhost:8080/api/v1/cart/user-123/items/prod-42 \
  -H "Content-Type: application/json" \
  -d '{"quantity": 3}'

# Удалить товар
curl -X DELETE http://localhost:8080/api/v1/cart/user-123/items/prod-42

# Очистить всю корзину
curl -X DELETE http://localhost:8080/api/v1/cart/user-123
```

---

## Команды Task

Запустите `task --list` для просмотра всех доступных команд.

| Команда | Описание |
|---|---|
| `task build` | Собрать бинарный файл в `./bin/marketplace-bucket` |
| `task run` | Запустить HTTP-сервер локально |
| `task test` | Запустить все тесты с детектором гонок |
| `task test:cover` | Тесты с отчётом о покрытии |
| `task test:integration` | Интеграционные тесты (требует Docker) |
| `task lint` | Запустить golangci-lint |
| `task arch` | Проверить архитектурные ограничения через go-arch-lint |
| `task vet` | Запустить go vet |
| `task fmt` | Форматировать код с gofmt |
| `task docker:up` | Запустить контейнеры Redis + Jaeger |
| `task docker:down` | Остановить контейнеры |
| `task docker:reset` | Остановить контейнеры и очистить тома |
| `task docs` | Перегенерировать документацию Swagger |

---

## Тестирование

Unit-тесты покрывают `CartUseCase` с мок-репозиторием (на основе интерфейса), а также HTTP-обработчики через `httptest`.

Интеграционные тесты используют `miniredis` для запуска реального Redis в процессе — Docker для `task test` не требуется.

```bash
task test
task test:integration
```

---

## Возможные улучшения для продакшена

- **Rate limiting** — ограничение запросов на пользователя для защиты Redis от пиковых нагрузок
- **Управление TTL** — отдельная задача для очистки истёкших корзин и публикации метрик
- **Оптимистичная блокировка** — Redis `WATCH`/`MULTI`/`EXEC` для конкурентных обновлений корзины
- **Публикация событий** — отправка событий `cart.updated` в Kafka для downstream-сервисов заказов
- **Auth middleware** — валидация JWT для привязки `userID` из claims токена, а не из URL
- **Правила алертинга** — Prometheus-алерты на высокий уровень ошибок и задержки Redis
