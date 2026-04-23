# Go Clean Architecture Real-Time Chat

A production-grade, highly scalable real-time chat app api built with Go. This project demonstrates how to build a distributed WebSocket backend using Clean Architecture principles, PostgreSQL for persistent storage, and Redis Pub/Sub for horizontal scalability.

## 🚀 Features

- **Real-Time WebSockets:** Low-latency, bidirectional communication using `gorilla/websocket`.
- **Horizontal Scalability:** Redis Pub/Sub acts as a central message broker, allowing users connected to different server instances to chat seamlessly.
- **Clean Architecture:** Strict separation of concerns (Domain, Repository, Usecase, and Delivery layers) making the codebase testable and database-agnostic.
- **Robust Concurrency:** Safe state management using Go channels (no mutex locks) and a highly tuned Ping/Pong heartbeat system to prevent memory leaks.
- **Production-Ready Server:** Includes graceful shutdown, structured JSON logging (`log/slog`), and protection against Slowloris attacks.
- **Configuration Management:** Strongly typed environment variable parsing using `envconfig`.

## 🛠 Tech Stack

- **Language:** Go (1.25+)
- **Routing:** [go-chi/chi](https://github.com/go-chi/chi)
- **WebSockets:** [gorilla/websocket](https://github.com/gorilla/websocket)
- **Database:** PostgreSQL with [jackc/pgx](https://github.com/jackc/pgx) (Connection Pooling)
- **Message Broker:** Redis with [redis/go-redis](https://github.com/redis/go-redis)
- **Configuration:** [godotenv](https://github.com/joho/godotenv) & [envconfig](https://github.com/kelseyhightower/envconfig)

## 📂 Project Structure

```text
.
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point & Dependency Injection
├── internal/
│   ├── config/                     # Environment variable parsing
│   ├── delivery/
│   │   └── ws/                     # WebSocket Handlers, Hub, and Client logic
│   ├── domain/                     # Core business entities and interfaces
│   ├── repository/                 # PostgreSQL implementation of data access
│   ├── storage/                    # Database and Redis connection setups
│   └── usecase/                    # Core business logic and rules
├── .env.example                    # Example environment variables
├── index.html                      # Simple frontend to test the chat
├── go.mod
└── go.sum
```
