# SimpleBank

A Go-based banking backend built with PostgreSQL, SQLC, Gin, and secure token-based authentication. The project models a simple money-transfer system with users, accounts, entries, and transfers, and is designed as a practical example of a layered backend with transactional database logic, API validation, and Docker-ready setup.

## Overview

SimpleBank is a learning-focused backend service for managing user accounts and transferring money between accounts. It includes:

- User registration and login
- JWT/PASETO token-based authentication
- Multi-currency account support
- Money transfer processing with transactional safety
- SQLC-generated database access layer
- PostgreSQL migrations
- Docker and Makefile-based local development workflow
- Unit and API tests

The application is built around a clean server layer that sits on top of PostgreSQL, with SQL queries defined in `db/query/*.sql` and code generation handled by SQLC.

## Features

- User management with username/email uniqueness checks
- Password hashing using bcrypt
- Authenticated account creation and listing
- Per-user authorization checks on account access
- Protected transfers between accounts
- Transfer transaction workflow that updates balances atomically
- Support for a small currency set: USD, EUR, CAD
- Database schema versioning using Go Migrate
- Docker Compose setup for local Postgres + API

## Tech Stack

- Go 1.25+
- PostgreSQL
- Gin HTTP framework
- sqlc for type-safe SQL generation
- golang-migrate for schema migrations
- PASETO as the default access token implementation
- JWT implementation included as an alternative/token library option
- Viper for configuration
- bcrypt for password hashing
- Docker + Docker Compose
- Testify + GoMock for testing

## Architecture

The application is organized into a few core layers:

- `api/` — HTTP handlers, route setup, validation, auth middleware
- `db/sqlc/` — generated database models and query methods
- `db/query/` — SQL query definitions used by sqlc
- `db/migration/` — schema migration files
- `token/` — token generation/verification implementations
- `util/` — common utilities, config, hashing, random helpers, currency validation
- `main.go` — app bootstrap and dependency wiring

### Request flow

1. `main.go` loads configuration from `app.env`.
2. A PostgreSQL connection is opened.
3. A `db.Store` is created using the SQLC-generated store.
4. `api.NewServer()` configures the Gin router and token maker.
5. HTTP requests are authenticated by middleware and processed in handlers.
6. Database operations execute against PostgreSQL with transactional safety for transfers.

## Project Structure

```text
.
├── api/
│   ├── account.go
│   ├── middleware.go
│   ├── server.go
│   ├── transfer.go
│   ├── user.go
│   ├── validator.go
│   ├── account_test.go
│   ├── main_test.go
│   └── ...
├── db/
│   ├── migration/
│   │   ├── 000001_init_schema.up.sql
│   │   ├── 000001_init_schema.down.sql
│   │   ├── 000002_add_users.up.sql
│   │   └── 000002_add_users.down.sql
│   ├── mock/
│   │   └── store.go
│   ├── query/
│   │   ├── account.sql
│   │   ├── entry.sql
│   │   ├── transfer.sql
│   │   └── users.sql
│   └── sqlc/
│       ├── db.go
│       ├── models.go
│       ├── querier.go
│       ├── store.go
│       ├── ... generated files ...
│       └── *_test.go
├── token/
│   ├── jwt_maker.go
│   ├── jwt_maker_test.go
│   ├── maker.go
│   ├── paseto_maker.go
│   ├── pasetp_maker_test.go
│   └── payload.go
├── util/
│   ├── config.go
│   ├── currency.go
│   ├── password.go
│   ├── password_test.go
│   ├── random.go
│   └── ...
├── Dockerfile
├── Makefile
├── README.md
├── app.env
├── docker-compose.yaml
├── go.mod
├── go.sum
├── main.go
├── sqlc.yaml
├── start.sh
├── wait-for.sh
└── ...
```

## Database Model

The schema is defined by migrations in `db/migration` and mapped to structs in `db/sqlc/models.go`.

### users

Stores user identity and authentication data.

Fields:

- `username` (primary key)
- `hashed_password`
- `full_name`
- `email` (unique)
- `password_changed_at`
- `created_at`

### accounts

Represents each user-owned bank account.

Fields:

- `id`
- `owner`
- `balance`
- `currency`
- `created_at`

Notes:

- `owner` references `users.username`
- Each user can only have one account per currency (`owner_currency_key` unique constraint)

### entries

Tracks money movement into or out of an account.

Fields:

- `id`
- `account_id`
- `amount`
- `created_at`

Note: `amount` can be negative or positive.

### transfers

Tracks a transfer between two accounts.

Fields:

- `id`
- `from_account_id`
- `to_account_id`
- `amount`
- `created_at`

## Authentication

Authentication is handled by middleware in `api/middleware.go`.

The middleware expects an Authorization header in the form:

```http
Authorization: Bearer <token>
```

It validates the token, extracts the payload, and stores it in the request context under `authorization_payload`.

The server is configured to use PASETO by default:

```go
tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
```

The project also includes a JWT implementation in `token/jwt_maker.go` and tests for both token systems. In practice, the app uses PASETO for normal runtime behavior.

## API Endpoints

### Public endpoints

#### Create user

- Method: `POST /users`
- Body:

```json
{
  "username": "alice",
  "password": "secret123",
  "full_name": "Alice Example",
  "email": "alice@example.com"
}
```

#### Login

- Method: `POST /users/login`
- Body:

```json
{
  "username": "alice",
  "password": "secret123"
}
```

Response includes:

- `access_token`
- `user` metadata

### Protected endpoints (require Bearer token)

#### Create account

- Method: `POST /accounts`
- Body:

```json
{
  "currency": "USD"
}
```

#### Get account by ID

- Method: `GET /accounts/:id`

The handler verifies the account belongs to the authenticated user before returning it.

#### List accounts

- Method: `GET /accounts?page_id=1&page_size=5`

Returns a paginated list of the authenticated user’s accounts.

#### Transfer money

- Method: `POST /transfers`
- Body:

```json
{
  "from_account_id": 1,
  "to_account_id": 2,
  "amount": 250,
  "currency": "USD"
}
```

Validation includes:

- positive transfer amount
- same-currency accounts
- source account must belong to the authenticated user
- transfer is processed in an atomic database transaction

## Transaction Design

Transfers are implemented as a database transaction in `db/sqlc/store.go`.

The flow is:

1. Insert a `transfers` record
2. Insert a debit entry for the source account
3. Insert a credit entry for the destination account
4. Update account balances atomically
5. Commit or rollback as a single transaction

This ensures the transfer either completes fully or leaves the system unchanged if an error occurs.

## Configuration

Configuration is loaded from `app.env` using Viper in `util/config.go`.

Example values:

```env
DB_DRVIER=postgres
DB_SOURCE=localhost:5432/simple_bank?sslmode=disable
SERVER_ADDRESS=0.0.0.0:8080
TOKEN_SYMMETRIC_KEY=12345678901234567890123456789012
ACCESS_TOKEN_DURATION=15m
```

Notes:

- `DB_DRVIER` is used as the database driver name
- `DB_SOURCE` points to the PostgreSQL connection string
- `SERVER_ADDRESS` controls the HTTP bind address
- `TOKEN_SYMMETRIC_KEY` must match the token library requirement size
- `ACCESS_TOKEN_DURATION` controls access token lifetime

## Local Development

### Prerequisites

- Go 1.25+
- Docker and Docker Compose
- PostgreSQL client tools (optional, but useful for debugging)
- `migrate` CLI if you want to run migrations manually

### Start PostgreSQL locally

```bash
make postgres
make createdb
```

### Run database migrations

```bash
make migrateup1
```

### Run the API server

```bash
make server
```

Or via Go directly:

```bash
go run main.go
```

The service listens on `0.0.0.0:8080` by default.

### Run with Docker Compose

```bash
docker compose up --build
```

Docker Compose starts:

- PostgreSQL container
- API service
- automatic DB migration before app startup via `start.sh`

## Useful Make Commands

```bash
make postgres
make createdb
make dropdb
make migrateup
make migrateup1
make migratedown
make migratedown1
make sqlc
make test
make server
make mock
```

## SQLC Workflow

The project uses SQLC to map SQL files into strongly typed Go code.

Source SQL files live in:

- `db/query/account.sql`
- `db/query/entry.sql`
- `db/query/transfer.sql`
- `db/query/users.sql`

Generated Go code lives under:

- `db/sqlc/*.go`

To regenerate:

```bash
make sqlc
```

## Testing

The project includes database tests and API tests.

Run the test suite with:

```bash
make test
```

Or:

```bash
go test ./...
```

Covered areas include:

- account CRUD and validation
- transfer creation and listing
- user creation and lookup
- password hashing and comparison
- token creation and expiry validation
- API auth and authorization logic

## Security Notes

This project is built as a backend example and includes several sensible security controls:

- Passwords are hashed with bcrypt
- Tokens are signed and verified
- Bearer-token auth middleware protects sensitive routes
- Authorization checks ensure a user can only access their own accounts
- Transfer validation prevents invalid or unauthorized transactions

Important: this is not a production-grade financial platform. Real-world banking systems would also require strict operational controls, audit logs, monitoring, secret management, rate limiting, and additional compliance hardening.

## Development Notes

A few practical observations from the codebase:

- `token.NewPasetoMaker` is the default token implementation used by the server
- A JWT variant exists and is tested, but it is not used in the router configuration by default
- The project uses `gin.Default()` and simple JSON responses for errors
- `app.env` is loaded automatically by Viper in the project root
- The custom currency validator ensures only supported currencies pass API validation

## Common Commands

```bash
# Start database
make postgres

# Create database
make createdb

# Run migrations
make migrateup1

# Launch server
make server

# Run tests
go test ./...

# Regenerate SQLC code
make sqlc
```

## License

This project is intended for learning and demonstration purposes. See the repository for any licensing details or add a formal license if needed.

## Summary

SimpleBank is a compact but complete Go backend example demonstrating a secure, transactional accounting system built around PostgreSQL, SQLC, and Gin. The codebase illustrates common patterns used in backend service development: schema migrations, SQL generation, secure auth, request validation, atomic transactions, and test-driven API development.
