# Project Architecture & Structure

This document outlines the software architecture and directory layout for the **Hospital Middleware System**, fulfilling **Deliverable 1.a** of the Agnos Candidate Assignment.

---

## 1. Architecture Overview (Clean Architecture / Layered Pattern)

The project adheres to Clean Architecture principles to ensure high cohesion, low coupling, testability, and maintainability:

```text
HTTP Request (Client / Postman)
       │
       ▼
[ Nginx Reverse Proxy (Port 80) ]
       │
       ▼
[ Gin HTTP Router & Middleware (JWT Auth) ]
       │
       ▼
[ Handlers / Controllers ]  <── Handles HTTP serialization, status codes, query parsing
       │
       ▼
[ Services (Business Logic) ]  <── Manages Auth, HIS integration, business rules
       │
       ├───► [ HIS Client ] (External Hospital A API / Mock)
       │
       ▼
[ Repositories (Data Access) ]  <── Encapsulates database queries via GORM
       │
       ▼
[ PostgreSQL Database (Docker) ]
```

---

## 2. Directory Layout

```text
agnos-assessment/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point, DB init, route setup
├── internal/
│   ├── config/
│   │   └── config.go               # Environment configuration loader
│   ├── handler/
│   │   ├── staff_handler.go        # HTTP handler for /staff/create & /staff/login
│   │   ├── patient_handler.go      # HTTP handler for /patient/search
│   │   └── patient_handler_test.go # Handler unit tests
│   ├── middleware/
│   │   ├── auth_middleware.go      # JWT verification middleware
│   │   └── auth_middleware_test.go # Middleware unit tests
│   ├── model/
│   │   └── models.go               # GORM entities (Hospital, Staff, Patient)
│   ├── repository/
│   │   ├── db.go                   # Database connection and AutoMigrate logic
│   │   ├── staff_repository.go     # Staff DB operations
│   │   └── patient_repository.go   # Patient DB operations with hospital scoping
│   ├── router/
│   │   └── router.go               # Gin route definitions & middleware wiring
│   └── service/
│       ├── auth_service.go         # Password hashing & JWT generation logic
│       ├── auth_service_test.go    # Auth unit tests (Positive & Negative)
│       ├── patient_service.go      # Patient search business logic
│       ├── patient_service_test.go # Patient service unit tests
│       └── his_client.go           # HIS Middleware client for Hospital A
├── docs/
│   ├── ER_DIAGRAM.md               # Database schema design & ER diagram
│   ├── API_SPEC.md                 # Complete API specifications
│   └── PROJECT_STRUCTURE.md        # Architecture overview (this file)
├── nginx/
│   └── nginx.conf                  # Nginx reverse proxy configuration
├── Dockerfile                      # Multi-stage Docker build for Go API
├── docker-compose.yml              # Multi-container orchestration (Nginx, Go, Postgres)
├── go.mod                          # Go module dependencies
├── go.sum                          # Go checksums
└── README.md                       # Complete guide to run, test, and evaluate
```

---

## 3. Layer Responsibilities

* **`cmd/server/main.go`**: Loads configs, connects to Postgres, runs AutoMigrate and database seeders, builds dependencies, and starts the HTTP server.
* **`internal/model`**: Defines database entities with GORM tags, validation constraints, and JSON serialization tags.
* **`internal/repository`**: Interfaces and implementations that talk to PostgreSQL. Contains the multi-tenant query filter to enforce hospital boundaries.
* **`internal/service`**: Pure business logic agnostic of HTTP transport. Easy to mock and unit test.
* **`internal/handler`**: Gin route handlers that parse request payloads/parameters, delegate to services, and write JSON responses.
* **`internal/middleware`**: Cross-cutting concerns, specifically JWT token validation and identity extraction.
