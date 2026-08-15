# microservices-ecommerce

Проект, выполненный в рамках курса по Go, представляет собой распределенную платформу электронной коммерции, построенную на базе микросервисной архитектуры и принципов Clean Architecture. Система разделена на ключевые сервисы: корзину (cart), систему управления заказами, товарами и складами (loms), а также сервис уведомлений (notifications). Взаимодействие между компонентами реализовано через синхронный gRPC с HTTP/REST-шлюзом (grpc-gateway) и надежную асинхронную доставку событий в Apache Kafka с использованием паттерна Transactional Outbox.
## Технологии

* **Go 1.26.2**
* **gRPC / Protocol Buffers**
* **gRPC-Gateway**
* **PostgreSQL**
* **sqlc**
* **Apache Kafka**
* **Transactional Outbox**
* **GoMock (`mockgen`)**
* **Docker / Docker Compose**
* **Taskfile (`task`)**
* **EasyP**

---

## Структура проекта

Проект организован как монорепозиторий на базе **Go Workspaces (`go.work`)**. Каждый микросервис изолирован и построен по принципам **Clean Architecture**.

```text
├── cart/                   # Сервис корзины 
├── loms/                   # Logistics & Order Management System (Заказы, Товары, Склады, Outbox)
├── notifications/          # Сервис нотификаций (Kafka Consumer)
├── integration-tests/      # Интеграционные тесты 
├── docs/                   # OpenAPI / Swagger 
├── Taskfile.yml            # Скрипты для автоматизации (генерация, линтинг, тесты, миграции)
└── go.work                 # Конфигурация Go Workspace
```

---

## Сервисы

### 1. Сервис корзины (`cart`)

Управление товарами в корзине пользователя, просмотр содержимого, очистка и запуск чекаута.

**API Эндпоинты:**
* `POST /v1/cart/item/add` — добавление товара в корзину
* `DELETE /v1/cart/item/delete` — удаление товара из корзины
* `GET /v1/cart/list` — просмотр содержимого корзины
* `POST /v1/cart/clear` — очистка корзины
* `POST /v1/cart/checkout` — оформление заказа из корзины

#### Добавление товара в корзину
```mermaid
sequenceDiagram
    actor User
    participant Cart as cart
    participant DB as Postgres (Cart)
    participant LOMS as loms (gRPC)

    User->>Cart: POST /v1/cart/item/add (user_id, sku, count)
    activate Cart
    Cart->>LOMS: GetProduct(sku)
    activate LOMS
    LOMS-->>Cart: name, price
    deactivate LOMS

    Cart->>LOMS: GetStock(sku)
    activate LOMS
    LOMS-->>Cart: available_count
    deactivate LOMS

    alt Остатков достаточно
        Cart->>DB: AddItem(user_id, sku, count)
        Cart-->>User: 200 OK
    else Недостаточно остатков
        Cart-->>User: 412 Precondition Failed (ErrInsufficientStock)
    end
    deactivate Cart
```

<details>
<summary><b>Просмотр корзины (List Cart)</b></summary>

```mermaid
sequenceDiagram
    actor User
    participant Cart
    participant CartStorage
    participant ProductService

    User->>Cart: POST /v1/cart/list<br/>- user_id
    activate Cart
    Cart->>CartStorage: GetItemsByUserID(user_id)
    loop for each item in cart
        Cart->>ProductService: GetProduct(sku)<br/>/v1/product/info
        activate ProductService
        ProductService-->>Cart: name, price
        deactivate ProductService
    end
    Cart-->>User: Response:<br/>items[]{ sku, count, name, price }<br/>total_price
    deactivate Cart
```

</details>

<details>
<summary><b>Создание заказа (Checkout)</b></summary>

```mermaid
sequenceDiagram
    actor User
    participant Cart
    participant CartStorage
    participant LOMS

    User->>Cart: POST /v1/cart/checkout<br/>- user_id
    activate Cart
    Cart->>CartStorage: GetItemsByUserID(user_id)
    Cart->>LOMS: CreateOrder(user_id, []item)<br/>/v1/order/create
    activate LOMS
    LOMS-->>Cart: order_id
    deactivate LOMS
    Cart->>CartStorage: DeleteItemsByUserID(user_id)
    Cart-->>User: 200 OK<br/>order_id
    deactivate Cart
```

</details>

<details>
<summary><b>Удаление товара из корзины (Delete Item)</b></summary>

```mermaid
sequenceDiagram
    actor User
    participant Cart
    participant CartStorage

    User->>Cart: DELETE /v1/cart/item/delete<br/>- user_id<br/>- sku
    activate Cart
    Cart->>CartStorage: DeleteItem(user_id, sku)
    activate CartStorage
    CartStorage-->>Cart: OK
    deactivate CartStorage
    Cart-->>User: 200 OK
    deactivate Cart
```

</details>

<details>
<summary><b>Очистка корзины (Clear Cart)</b></summary>

```mermaid
sequenceDiagram
    actor User
    participant Cart
    participant CartStorage

    User->>Cart: POST /v1/cart/clear<br/>- user_id
    activate Cart
    Cart->>CartStorage: ClearCartByUserID(user_id)
    activate CartStorage
    CartStorage-->>Cart: OK
    deactivate CartStorage
    Cart-->>User: 200 OK
    deactivate Cart
```

</details>

---

### 2. Сервис логистики и заказов (`loms`)

Logistics & Order Management System: управление заказами, каталогом товаров, складскими остатками и публикация событий через Transactional Outbox.

**API Эндпоинты:**

* **Orders (`Loms`):**
  * `POST /v1/order/create` — создание заказа и резервация остатков
  * `POST /v1/order/info` — получение информации о заказе
  * `POST /v1/order/pay` — оплата заказа
  * `POST /v1/order/cancel` — отмена заказа и снятие резервов
* **Products (`ProductService`):**
  * `POST /v1/product/create` — добавление товара в каталог
  * `POST /v1/product/info` — получение информации о товаре по SKU
  * `POST /v1/product/list` — батч-получение информации по списку SKU
* **Stocks (`Stocks`):**
  * `POST /v1/stock/set` — установка остатков по SKU
  * `POST /v1/stock/info` — получение доступного остатка по SKU

#### Создание заказа (Create Order & Outbox)
```mermaid
sequenceDiagram
    actor Cart
    participant LOMS
    participant StocksStorage
    participant OrdersStorage
    participant OutboxStorage
    participant OutboxWorker
    participant Kafka

    Cart->>LOMS: POST /v1/order/create<br/>- user_id<br/>- items[]{ sku, count }
    activate LOMS
    LOMS->>StocksStorage: ReserveStocks(items)
    alt reserve success
        Note over LOMS,OutboxStorage: В одной транзакции (Transactor)
        LOMS->>OrdersStorage: CreateOrder(user_id, items)<br/>status=ORDER_STATUS_AWAITING_PAYMENT
        LOMS->>OutboxStorage: SaveOutboxEvent(topic: "order_events", status: "created")
        LOMS-->>Cart: 200 OK<br/>order_id
    else reserve failed
        LOMS-->>Cart: 412 Failed Precondition<br/>ErrInsufficientStock
    end
    deactivate LOMS

    loop Background Worker
        OutboxWorker->>OutboxStorage: FetchUnprocessedEvents()
        OutboxWorker->>Kafka: Produce(order_events, msg)
        OutboxWorker->>OutboxStorage: MarkEventProcessed(event_id)
    end
```

<details>
<summary><b>Получение информации о заказе (Get Order)</b></summary>

```mermaid
sequenceDiagram
    actor User
    participant LOMS
    participant OrdersStorage

    User->>LOMS: POST /v1/order/info<br/>- order_id
    activate LOMS
    LOMS->>OrdersStorage: GetOrderByID(order_id)
    activate OrdersStorage
    OrdersStorage-->>LOMS: order (status, user_id, items, timestamps)
    deactivate OrdersStorage
    LOMS-->>User: 200 OK<br/>status, user_id, items, created_at, updated_at
    deactivate LOMS
```

</details>

<details>
<summary><b>Оплата заказа (Pay Order)</b></summary>

```mermaid
sequenceDiagram
    actor PaymentService as User / Payment
    participant LOMS
    participant OrdersStorage
    participant OutboxStorage

    PaymentService->>LOMS: POST /v1/order/pay<br/>- order_id
    activate LOMS
    Note over LOMS,OutboxStorage: В одной транзакции (Transactor)
    LOMS->>OrdersStorage: UpdateOrderStatus(order_id, status=ORDER_STATUS_PAID)
    LOMS->>OutboxStorage: SaveOutboxEvent(topic: "order_events", status: "paid")
    LOMS-->>PaymentService: 200 OK
    deactivate LOMS
```

</details>

<details>
<summary><b>Отмена заказа (Cancel Order)</b></summary>

```mermaid
sequenceDiagram
    actor User
    participant LOMS
    participant OrdersStorage
    participant StocksStorage
    participant OutboxStorage

    User->>LOMS: POST /v1/order/cancel<br/>- order_id
    activate LOMS
    Note over LOMS,OutboxStorage: В одной транзакции (Transactor)
    LOMS->>OrdersStorage: GetOrderByID(order_id)
    LOMS->>StocksStorage: ReleaseStocks(order.items)
    LOMS->>OrdersStorage: UpdateOrderStatus(order_id, status=ORDER_STATUS_CANCELLED)
    LOMS->>OutboxStorage: SaveOutboxEvent(topic: "order_events", status: "cancelled")
    LOMS-->>User: 200 OK
    deactivate LOMS
```

</details>

<details>
<summary><b>Создание товара (Create Product)</b></summary>

```mermaid
sequenceDiagram
    actor Client
    participant LOMS
    participant ProductStorage

    Client->>LOMS: POST /v1/product/create<br/>- name, price
    activate LOMS
    LOMS->>ProductStorage: CreateProduct(name, price)
    activate ProductStorage
    ProductStorage-->>LOMS: sku
    deactivate ProductStorage
    LOMS-->>Client: 200 OK<br/>sku
    deactivate LOMS
```

</details>

<details>
<summary><b>Получение информации о товаре (Get Product Info)</b></summary>

```mermaid
sequenceDiagram
    actor Client
    participant LOMS
    participant ProductStorage

    Client->>LOMS: POST /v1/product/info<br/>- sku
    activate LOMS
    LOMS->>ProductStorage: GetProductBySKU(sku)
    activate ProductStorage
    ProductStorage-->>LOMS: name, price
    deactivate ProductStorage
    LOMS-->>Client: 200 OK<br/>name, price
    deactivate LOMS
```

</details>

<details>
<summary><b>Список товаров (List Products)</b></summary>

```mermaid
sequenceDiagram
    actor Client
    participant LOMS
    participant ProductStorage

    Client->>LOMS: POST /v1/product/list<br/>- skus[]
    activate LOMS
    LOMS->>ProductStorage: GetProductsBySKUs(skus[])
    activate ProductStorage
    ProductStorage-->>LOMS: products[]{ sku, name, price }
    deactivate ProductStorage
    LOMS-->>Client: 200 OK<br/>products[]
    deactivate LOMS
```

</details>

<details>
<summary><b>Установка остатка товара (Set Stock)</b></summary>

```mermaid
sequenceDiagram
    actor Client
    participant LOMS
    participant StocksStorage

    Client->>LOMS: POST /v1/stock/set<br/>- sku, count
    activate LOMS
    LOMS->>StocksStorage: SetStock(sku, count)
    activate StocksStorage
    StocksStorage-->>LOMS: OK
    deactivate StocksStorage
    LOMS-->>Client: 200 OK
    deactivate LOMS
```

</details>

<details>
<summary><b>Получение остатка товара (Get Stock Info)</b></summary>

```mermaid
sequenceDiagram
    actor Client
    participant LOMS
    participant StocksStorage

    Client->>LOMS: POST /v1/stock/info<br/>- sku
    activate LOMS
    LOMS->>StocksStorage: GetStockBySKU(sku)
    activate StocksStorage
    StocksStorage-->>LOMS: count
    deactivate StocksStorage
    LOMS-->>Client: 200 OK<br/>count
    deactivate LOMS
```

</details>

### 3. Сервис уведомлений (`notifications`)

Асинхронная обработка событий заказов из Apache Kafka с ручным управлением оффсетами (Manual Commit) и поддержкой очереди недоставленных сообщений (**Dead Letter Queue / DLQ**).

**Ключевые особенности консьюмера:**
* **Manual Commit:** фиксация смещения (`CommitMessage`) строго после успешной обработки в `notifier`.
* **DLQ (Dead Letter Queue):** битые сообщения (ошибки парсинга JSON) автоматически перенаправляются в топик `{topic}-dlq` с сохранением контекста ошибки, исходных партиций и оффсетов.
* **Retry Loop:** при сбое обработчика отправки сообщение не коммитится и повторно вычитывается.

#### Обработка событий заказа (Kafka Consumer Flow & DLQ)
```mermaid
sequenceDiagram
    participant Kafka as Apache Kafka (order_events)
    participant Consumer as Notifications Consumer
    participant DLQ as Kafka DLQ (order_events-dlq)
    participant Notifier as Notifier Handler
    actor User

    Kafka->>Consumer: Poll() -> Message
    activate Consumer

    alt Ошибка парсинга JSON (Unmarshal error)
        Consumer->>DLQ: Produce to DLQ (original msg + error reason)
        Consumer->>Kafka: CommitMessage(offset)
    else JSON валидный
        Consumer->>Notifier: SendMessageNotificationsKindHandler(payload)
        alt Успешная обработка
            Notifier->>User: Отправка уведомления
            Notifier-->>Consumer: nil (success)
            Consumer->>Kafka: CommitMessage(offset)
        else Ошибка отправки
            Notifier-->>Consumer: error
            Note over Consumer: Retry: пауза 1s, оффсет НЕ коммитится
        end
    end
    deactivate Consumer
```

## Локальный запуск

```bash
go install [github.com/go-task/task/v3/cmd/task@latest](https://github.com/go-task/task/v3/cmd/task@latest)
```
```
task backend
```

```
docker compose ps
```

###  Порты и доступ к сервисам

| Сервис / Инстанс | Назначение | Протокол | Хост и Порт |
| :--- | :--- | :--- | :--- |
| **`cart` (Node 1)** | Сервис корзины | HTTP (REST / Gateway)<br/>gRPC | `http://localhost:8080`<br/>`localhost:50051` |
| **`cart-2` (Node 2)** | Сервис корзины (реплика) | HTTP (REST / Gateway)<br/>gRPC | `http://localhost:8090`<br/>`localhost:50061` |
| **`loms` (Node 1)** | Заказы, товары, склады | HTTP (REST / Gateway)<br/>gRPC | `http://localhost:8081`<br/>`localhost:50052` |
| **`loms-2` (Node 2)** | Заказы, товары, склады (реплика) | HTTP (REST / Gateway)<br/>gRPC | `http://localhost:8091`<br/>`localhost:50062` |
| **`notifications`** | Сервис нотификаций | gRPC | `localhost:50053` |
| **PostgreSQL** | База данных (`ecommerce`) | TCP | `localhost:5432` |
| **Apache Kafka** | Брокер сообщений | PLAINTEXT | `localhost:9092` / `localhost:9093` |
| **Kafdrop** | Web UI для Apache Kafka | HTTP | `http://localhost:9000` |

###  Переменные окружения (Environment Variables)

**Общие параметры подключения к базе данных (PostgreSQL)**
* `POSTGRES_HOST` — хост для подключения к СУБД (`postgres`)
* `POSTGRES_PORT` — порт сервера PostgreSQL (`5432`)
* `POSTGRES_DB` — имя рабочей базы данных (`ecommerce`)
* `POSTGRES_USER` — имя пользователя БД (`ecommerce_user`)

**Сервис корзины (`cart` / `cart-2`)**
* `GRPC_PORT` — сетевой порт gRPC-сервера (`50051` / `50061`)
* `GRPC_GATEWAY_PORT` — сетевой порт HTTP REST gRPC-Gateway (`8080` / `8090`)
* `LOMS_GRPC_ADDR` — адрес downstream-сервиса LOMS по gRPC (`loms:50052` / `loms-2:50062`)

**Сервис заказов и логистики (`loms` / `loms-2`)**
* `GRPC_PORT` — сетевой порт gRPC-сервера (`50052` / `50062`)
* `GRPC_GATEWAY_PORT` — сетевой порт HTTP REST gRPC-Gateway (`8081` / `8091`)
* `OUTBOX_WORKERS` — количество параллельных фоновых воркеров Outbox (`1`)
* `OUTBOX_BATCH_SIZE` — размер пачки сообщений, вычитываемых из БД за итерацию (`10`)
* `OUTBOX_FETCH_PERIOD` — интервал опроса таблицы Outbox (`200ms`)
* `OUTBOX_IN_PROGRESS_TTL` — время жизни зависшего сообщения перед повторной отправкой (`1s`)
* `KAFKA_BROKERS` — список брокеров Apache Kafka (`kafka-single:9092`)
* `KAFKA_NOTIFICATIONS_TOPIC` — целевой топик для публикации событий заказа (`order_status_notifications`)

**Сервис уведомлений (`notifications`)**
* `GRPC_PORT` — сетевой порт gRPC-сервера (`50053`)
* `CALLBACK_ADDR` — адрес внешнего сервиса для обратных HTTP-уведомлений (`integration-tests:8080`)
* `KAFKA_BROKERS` — список брокеров Apache Kafka (`kafka-single:9092`)
* `KAFKA_NOTIFICATIONS_TOPIC` — топик для вычитывания событий (`order_status_notifications`)
* `KAFKA_CONSUMER_GROUP` — идентификатор consumer-группы Kafka (`notifications`)

## Тестирование
* `task test` - запуск интеграционных тестов
* `task unit-test` - запуск юнит тестов(перед запуском нужно выполнить комаду 
task generate для генерации моков)




## Другие полезные команды

* `task up` — фоновый запуск сервисов
* `task down` — остановка и очистка
* `task lint` — проверка линтером
* `task generate` — генерация всего кода
* `task fast-generate` — быстрая кодогенерация
* `task bin-deps` — установка локальных утилит
* `task frontend` — запуск фронтенда
* `task backend` — запуск бекенда