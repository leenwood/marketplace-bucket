Ты опытный Go-разработчик. Напиши production-ready микросервис для управления корзиной покупок.

## Задача
Реализуй marketplace-bucket на Go, который управляет корзинами пользователей через кеш.

## Транспорт
- REST (HTTP/JSON)
- gRPC

## Функциональность
- Добавление товара в корзину
- Удаление товара из корзины
- Изменение количества товара
- Просмотр содержимого корзины
- Очистка корзины
- TTL / авто-истечение корзины

## Архитектура и инфраструктура
- Redis как основное хранилище (Hash + JSON)
- Чистая архитектура (handler → service → repository)
- Docker + docker-compose
- Конфигурация через env / Viper
- Graceful shutdown
- Метрики Prometheus + эндпоинт /metrics
- Трейсинг OpenTelemetry

## Качество кода
- Unit-тесты с моками (testify)
- Swagger / OpenAPI документация
- Makefile с командами build, test, lint, run
- golangci-lint конфигурация

## Структура проекта
Используй стандартную Go-структуру:marketplace-bucket/
├── cmd/server/main.go
├── internal/
│   ├── handler/   # транспортный слой
│   ├── service/   # бизнес-логика
│   └── repository/ # работа с Redis
├── pkg/           # переиспользуемый код
└── proto/         # .proto файлы (если gRPC)
## Модель данных в Redis
- Ключ корзины: `cart:{user_id}`
- Структура: Hash или JSON с полями товара (product_id, name, price, quantity)
- TTL: 7 дней (если не указано иное)

## Требования к коду
- Go 1.25+
- Используй интерфейсы для всех зависимостей (DI)
- Возвращай осмысленные HTTP-статусы и gRPC-коды ошибок
- Не используй global state
- Комментарии на английском языке
- Весь код должен компилироваться без ошибок

Начни с `go.mod`, затем реализуй слои снизу вверх: repository → service → handler → main.
