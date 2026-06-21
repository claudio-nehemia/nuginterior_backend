# BACKEND_STACK.md

## FULL VERSION

````md
# Backend Stack Documentation

## Core Stack

- Language: Golang
- Framework: Gin
- ORM: GORM
- Database: PostgreSQL
- Cache: Redis
- Authentication: JWT
- API Style: REST API
- Architecture: Modular Feature-Based + Layered Architecture

---

# Backend Libraries

## HTTP Framework
- Gin

Purpose:
- Routing
- Middleware
- REST API
- HTTP handling

---

## ORM
- GORM

Purpose:
- Query builder
- Relation handling
- Transactions
- ORM abstraction

---

## Database
- PostgreSQL

Purpose:
- Main relational database

---

## Cache & Queue
- Redis
- asynq

Purpose:
- Cache
- Queue system
- OTP
- Session
- Background jobs

---

## Environment Configuration
- godotenv
- viper

Purpose:
- Environment variable management
- Config management

---

## Validation
- validator/v10

Purpose:
- Request validation

---

## Authentication
- golang-jwt/jwt/v5

Purpose:
- JWT authentication

---

## Logging
- uber-go/zap

Purpose:
- Structured logging
- Production logging

---

## UUID
- google/uuid

Purpose:
- UUID generation

---

## HTTP Client
- resty/v2

Purpose:
- External API requests

---

## Swagger Documentation
- swaggo/gin-swagger
- swaggo/files
- swaggo/swag

Purpose:
- API documentation

---

## Migration
- golang-migrate

Purpose:
- Database migration
- Rollback
- Versioning

---

## CORS
- gin-contrib/cors

Purpose:
- Cross-origin configuration

---

## Retry System
- cenkalti/backoff

Purpose:
- Retry external requests

---

## Cron Jobs
- robfig/cron

Purpose:
- Scheduled jobs

---

## Security Headers
- unrolled/secure

Purpose:
- HTTP security middleware

---

# Backend Folder Structure

```bash
cmd/
internal/
pkg/
migrations/
docs/
scripts/
````

---

# Backend Module Structure

```bash
modules/
└── auth/
    ├── dto/
    ├── entity/
    ├── repository/
    ├── service/
    ├── handler/
    ├── routes/
    └── middleware/
```

---

# Backend Architecture Flow

```text
Request
  ↓
Middleware
  ↓
Handler
  ↓
Service
  ↓
Repository
  ↓
Database
```

---

# Backend Standards

## Response Format

Success:

```json
{
  "success": true,
  "message": "Success",
  "data": {}
}
```

Error:

```json
{
  "success": false,
  "message": "Error message"
}
```

---

# Backend Environment Variables

```env
APP_NAME=MyApp
APP_PORT=8080

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASS=password
DB_NAME=myapp

JWT_SECRET=secret

REDIS_HOST=localhost
REDIS_PORT=6379
```

---

# Backend Middleware

* Logger middleware
* Recovery middleware
* CORS middleware
* JWT middleware
* Request ID middleware
* Rate limiter middleware
* Secure headers middleware
* Timeout middleware

---

# Backend Coding Standards

## File Naming

* kebab-case

Example:

```text
user-service.go
auth-handler.go
jwt-helper.go
```

---

## REST API Naming

```text
GET    /users
GET    /users/:id
POST   /users
PUT    /users/:id
DELETE /users/:id
```

---

# Backend Security Standards

* JWT validation
* Request validation
* Rate limiting
* Secure headers
* SQL injection prevention via ORM
* Environment variable isolation

````