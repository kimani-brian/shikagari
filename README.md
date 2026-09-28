# ShikaGari 🚗

> Kenya's car marketplace API — built with Go, Gin, and PostgreSQL.

ShikaGari (Swahili: *"get a car"*) is a RESTful backend for a vehicle
marketplace platform supporting public browsing, dealer and private seller
listings, admin approval workflows, a favorites wishlist, and a buyer-seller
inquiry system.

---

## Table of Contents

- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
- [Environment Variables](#environment-variables)
- [Running the Server](#running-the-server)
- [API Reference](#api-reference)
- [Authentication](#authentication)
- [User Roles](#user-roles)
- [Seller Approval Flow](#seller-approval-flow)
- [Kenya-Specific Notes](#kenya-specific-notes)
- [Development Notes](#development-notes)

---

## Tech Stack

| Layer        | Technology              |
|--------------|-------------------------|
| Language     | Go 1.22                 |
| Framework    | Gin                     |
| Database     | PostgreSQL 15+          |
| ORM          | GORM                    |
| Auth         | JWT (HS256)             |
| Passwords    | bcrypt (cost 12)        |
| File storage | Local (swap for S3)     |

---

## Project Structure
```
shikagari/
├── cmd/api/main.go              # Entry point
├── config/                      # Config loader
├── internal/
│   ├── domain/                  # Database models
│   ├── dto/                     # Request / response shapes
│   ├── handler/                 # HTTP controllers (Gin)
│   ├── middleware/              # JWT auth, RBAC, logger, CORS
│   ├── repository/              # Data access layer
│   │   ├── interfaces/          # Repository contracts
│   │   └── postgres/            # PostgreSQL implementations
│   ├── router/                  # Route registration + DI wiring
│   └── service/                 # Business logic
├── migrations/                  # Raw SQL migration files
├── pkg/
│   ├── database/                # DB connection + AutoMigrate
│   ├── hash/                    # bcrypt helpers
│   ├── jwt/                     # JWT generation & parsing
│   ├── response/                # Standardised API response wrapper
│   ├── upload/                  # Image upload service
│   └── validator/               # Struct validation helpers
├── .env.example
├── go.mod
└── README.md
```

---

## Prerequisites

- Go 1.22+
- PostgreSQL 15+
- `make` (optional, for convenience commands)

---

## Getting Started

### 1. Clone the repository
```bash
git clone https://github.com/your-org/shikagari.git
cd shikagari
```

### 2. Install dependencies
```bash
go mod download
```

### 3. Set up the database
```bash
# Connect to PostgreSQL and create the database and user
psql -U postgres

CREATE USER shikagari_user WITH PASSWORD 'your_secure_password';
CREATE DATABASE shikagari_db OWNER shikagari_user;
GRANT ALL PRIVILEGES ON DATABASE shikagari_db TO shikagari_user;

# Enable UUID generation extension
\c shikagari_db
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
\q
```

### 4. Configure environment
```bash
cp .env.example .env
# Edit .env with your values
```

### 5. Run the server
```bash
go run cmd/api/main.go
```

The API will be available at `http://localhost:8080`.

---

## Environment Variables

| Variable            | Required | Default        | Description                          |
|---------------------|----------|----------------|--------------------------------------|
| `APP_ENV`           | No       | `development`  | `development` or `production`        |
| `APP_PORT`          | No       | `8080`         | HTTP server port                     |
| `APP_NAME`          | No       | `ShikaGari`    | Application name                     |
| `DB_HOST`           | No       | `localhost`    | PostgreSQL host                      |
| `DB_PORT`           | No       | `5432`         | PostgreSQL port                      |
| `DB_USER`           | No       | `shikagari_user` | PostgreSQL user                    |
| `DB_PASSWORD`       | **Yes**  | —              | PostgreSQL password                  |
| `DB_NAME`           | No       | `shikagari_db` | PostgreSQL database name             |
| `DB_SSLMODE`        | No       | `disable`      | `disable` (dev) or `require` (prod)  |
| `JWT_SECRET`        | **Yes**  | —              | Min 32-character signing secret      |
| `JWT_EXPIRY_HOURS`  | No       | `72`           | Token validity in hours              |
| `UPLOAD_DIR`        | No       | `./uploads`    | Local image storage path             |
| `MAX_FILE_SIZE_MB`  | No       | `5`            | Max upload size in megabytes         |

---

## Running the Server

### Development
```bash
go run cmd/api/main.go
```

### With live reload (using Air)
```bash
# Install Air
go install github.com/air-verse/air@latest

# Run with hot reload
air
```

### Production build
```bash
go build -o bin/shikagari ./cmd/api
./bin/shikagari
```

### Health check
```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "ok",
  "service": "ShikaGari",
  "env": "development"
}
```

---

## Authentication

ShikaGari uses **JWT Bearer token** authentication.

### Obtaining a token
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password123"}'
```

### Using the token

Include the token in the `Authorization` header on all protected routes:
```
Authorization: Bearer <your_token>
```

### Token payload
```json
{
  "user_id":    "uuid",
  "email":      "user@example.com",
  "role":       "buyer | seller | dealer | admin",
  "is_verified": false,
  "exp":         1234567890
}
```

---

## User Roles

| Role     | Capabilities                                                        |
|----------|---------------------------------------------------------------------|
| `buyer`  | Browse listings, send inquiries, manage favorites                   |
| `seller` | Private (individual) seller: everything a buyer can do + create a private seller profile, then list own cars after admin approval |
| `dealer` | Business dealership: everything a buyer can do + create a dealer profile, then list inventory after admin approval |
| `admin`  | Everything + approve/reject profiles, manage all users              |

> **Note:** The `admin` role is assigned manually by a superadmin directly
> in the database. It cannot be self-assigned via the registration endpoint.

---

## Seller Approval Flow
```
1. User registers with a role: "buyer" | "seller" (private) | "dealer"
   (registration form: Buy card → buyer; Sell card → Dealer / Private-seller sub-cards)
         │
         ▼
2. Seller creates a private seller profile  →  POST /sellers/profile
   OR dealer creates a dealer profile       →  POST /dealers/profile
        │
        ▼
3. Profile status = "pending"
   Admin reviews in:
     GET   /admin/dealers              (list pending dealers)
     PATCH /admin/dealers/:id/review   (approve or reject)
        │
        ▼
4. On approval: status = "approved"
   Seller can now create listings  →  POST /listings
        │
        ▼
5. Admin can assign verified badge →  PATCH /admin/users/:id
   { "is_verified": true }
   Verified badge appears on all seller listings
```

---

## Kenya-Specific Notes

- All prices are stored and returned in **Kenyan Shillings (KES)**
- Location values are validated against the 47 Kenyan counties
  (see `KenyanCounties` in `internal/domain/location.go`).
  Listing search filters still accept the original city list.
- Database timezone is set to `Africa/Nairobi`
- Mileage is stored in **kilometres**
- API responses are designed to be lightweight for mobile and
  low-bandwidth users (card vs detail response shapes)
- KRA PIN field is included on dealer profiles for tax compliance

---

## Development Notes

- GORM AutoMigrate runs on every server start in development
- Soft deletes are used on `users` and `listings` to preserve
  referential integrity on `inquiries` and `favorites`
- The `view_count` increment uses a raw atomic SQL `UPDATE`
  to prevent race conditions under concurrent requests
- Upload storage defaults to local filesystem (`./uploads/`).
  To switch to S3, replace the internals of `UploadService.UploadImage`
  while keeping the same method signature
- All repository methods return `nil, nil` (not an error) when a
  record is not found — callers handle the not-found case explicitly