# Infario Project Documentation

## Overview

**Infario** is a cloud platform API service for building, deploying, and managing projects. It integrates with Git providers (GitHub, GitLab, Bitbucket) to automate CI/CD pipelines through Docker-based deployments.

### Key Responsibilities
- Project management (CRUD operations)
- Deployment orchestration with Docker
- Git provider integration (webhooks, repository access)
- Background task processing via worker pools
- Request validation and error handling
- Rate limiting and authentication

---

## Architecture

Infario follows a **layered architecture** with clear separation of concerns:

```
HTTP Request
    ↓
┌─────────────────────────────────────┐
│  GATEWAY (Security & Middleware)    │ ← Auth, Rate Limit, Logging, Recovery
└─────────────────────────────────────┘
    ↓
┌─────────────────────────────────────┐
│  RESOURCE (API Domain Logic)        │ ← Projects, Deployments (CRUD)
│  • Routes (HTTP Handlers)           │
│  • Services (Business Logic)        │
│  • Repositories (Data Access)       │
└─────────────────────────────────────┘
    ↓
┌─────────────────────────────────────┐
│  PLATFORM (Engine Room)             │ ← Docker, Workers, Scheduler
└─────────────────────────────────────┘
    ↓
┌─────────────────────────────────────┐
│  PROVIDER (External Services)       │ ← GitHub, GitLab, Bitbucket
└─────────────────────────────────────┘
```

### Layer Responsibilities

| Layer | Purpose | Examples |
|-------|---------|----------|
| **Gateway** | Protect routes, enforce rules | Auth middleware, rate limiting, request logging |
| **Resource** | Handle API operations | Project CRUD, validation, HTTP endpoints |
| **Platform** | Execute heavy tasks | Docker builds, background workers, scheduled cleanup |
| **Provider** | Integrate external services | Git API clients, webhooks |

---

## Directory Structure

### `/cmd/app/main.go`
- **Entry point** for the application
- Wires together database, routes, and HTTP server
- Loads environment configuration
- Runs database migrations
- Starts graceful shutdown handling

### `/internal/gateway/`
**Security layer protecting all routes**

- `auth/jwt.go` - JWT token generation and validation
- `middleware/`
  - `auth.go` - Protects routes by validating JWT tokens
  - `logging.go` - Logs HTTP requests/responses with structured logging
  - `recovery.go` - Panic recovery prevents server crashes
- `ratelimit/limiter.go` - Sliding window rate limiting algorithm

### `/internal/resources/`
**API domain orchestrator**

- `resources.go` - Module initializer: calls each domain's `Init()` function
- `project/` - Project management module
- `deployment/` - Deployment orchestration module (placeholder for future work)

### `/internal/resources/project/`
**Complete project CRUD implementation**

- `domain.go` - DTOs and interfaces
  - `Project` - Core entity
  - `CreateProject`, `UpdateProject`, `DeleteProject`, `GetSingleProject` - Request DTOs
  - `ProjectRepository` interface - Data access contract
  - `ProjectService` interface - Business logic contract

- `service.go` - Business logic layer
  - Validates all inputs using `validator.Validate`
  - Delegates CRUD to repository
  - Returns properly formatted domain objects

- `postgres.go` - PostgreSQL implementation
  - SQL queries for Create, Read, Update, Delete
  - Soft deletes using `deleted_at` timestamp
  - Uses PostgreSQL `pgxpool` for connection pooling

- `routes.go` - HTTP handlers
  - `GET /projects/{id}` - Retrieve project
  - `POST /projects` - Create project
  - `PATCH /projects/{id}` - Update project
  - `DELETE /projects/{id}` - Soft delete project
  - Includes Swagger documentation comments

- `module.go` - Dependency injection
  - Creates repository → service → handler chain
  - Registers routes with HTTP mux

### `/internal/platform/`
**Engine room for heavy lifting**

- `engine/docker.go` - Docker SDK integration
  - Builds images from project repositories
  - Pushes built images to registries
  - Container orchestration

- `scheduler/cron.go` - Background task scheduler
  - Cleanup of old deployments
  - Scheduled maintenance tasks

- `worker/pool.go` - Async task worker pool
  - Queues deployment builds
  - Processes builds asynchronously
  - Integrates with Redis for task persistence

### `/internal/provider/`
**External service integrations**

- `provider.go` - Shared interfaces
  - `GitProvider` interface - Git repository access
  - `CloudProvider` interface - Cloud deployment targets

- `github/client.go` - GitHub API client
  - Repository access and metadata
  - Webhook setup and verification
  - Commit information retrieval

- `gitlab/client.go` - GitLab API client (skeleton)
- `bitbucket/client.go` - Bitbucket API client (skeleton)

### `/pkgs/`
**Shared utilities used across modules**

- `database/postgres.go`
  - `NewPostgres()` - Creates pgxpool with connection limits
  - `RunMigrations()` - Executes SQL migrations on startup

- `config/config.go`
  - Loads environment variables using `caarlos0/env`
  - Supports `.env` files via `godotenv`
  - Default values for development

- `request/params.go`
  - `PagingParams` - Standard pagination parameters (pageNumber, pageSize)
  - `Offset()` - Converts page-based params to database offset
  - `ParsePaging()` - Parses pagination from HTTP query params with defaults

- `response/response.go`
  - `JSON()` - Writes JSON with proper headers
  - `Error()` - Standard error response format
  - `MapValidationErrors()` - Converts validator.ValidationErrors to field messages

- `response/page.go`
  - `Collection[T]` - Generic paginated response with Items, TotalCount, PageNumber, PageSize, PageCount
  - `NewCollection()` - Helper to create paginated responses with calculated page count

- `validator/validator.go`
  - Global `validator.Validate` instance
  - Custom tag name function: uses `json` tag instead of struct field names

### `/migrations/`
**SQL schema definitions**

- `000001_init_schema.up.sql` - Creates `projects` and `deployments` tables
- `000001_init_schema.down.sql` - Rollback script

### `/docs/`
**Auto-generated Swagger documentation**
- Generated by `swag init` command
- Includes all route definitions with examples and error responses

---

## Database Schema

### `projects` table
```sql
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    git_url VARCHAR(255) NOT NULL,
    git_provider VARCHAR(20) NOT NULL DEFAULT 'github',
    primary_branch VARCHAR(50) DEFAULT 'main',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);
```

**Key Features:**
- UUID primary key (auto-generated)
- Soft deletes via `deleted_at` (queries always filter where `deleted_at IS NULL`)
- Unique name constraint
- Timestamps for audit trail

### `deployments` table
```sql
CREATE TABLE deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    commit_hash VARCHAR(40) NOT NULL,
    status SMALLINT NOT NULL DEFAULT 0,  -- 0:Queued, 1:Building, 2:Ready, 3:Failed
    preview_url VARCHAR(255) UNIQUE,
    storage_key TEXT,
    metadata_json JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

**Key Features:**
- Foreign key to `projects` with CASCADE delete
- Status machine: Queued → Building → Ready/Failed
- Unique preview URL
- JSONB for flexible git metadata
- Indexes for performance:
  - `idx_deployments_preview_url` - Fast lookup of ready deployments
  - `idx_deployments_project_latest` - Get latest deployment per project

---

## Development Workflow

### Setup
```bash
# Load environment variables
cp .env.example .env

# Start the full stack (database, Redis, API)
make up

# View logs
make logs

# Enter container shell
make shell
```

### Code Changes
```bash
# API auto-reloads via Air (configured in .air.toml)
# Just save your changes and refresh the request
```

### Generate Swagger Docs
```bash
# After adding new routes or updating Swagger comments
make swagger

# Then restart the server to pick up changes
make restart
```

### Database Migrations
```bash
# Migrations run automatically on server startup
# Create new migration:
touch migrations/000002_add_column.up.sql
touch migrations/000002_add_column.down.sql

# Write SQL, restart server to apply
```

---

## Coding Patterns

### Module Pattern (Dependency Injection)
Every domain (project, deployment) follows this structure:

```
domain/
├── domain.go      → Types & Interfaces
├── service.go     → Business Logic
├── postgres.go    → Data Access
├── routes.go      → HTTP Handlers
└── module.go      → Wire it together (Init function)
```

**Initialization Flow:**
```go
// In module.go
func Init(mux *http.ServeMux, pgx *pgxpool.Pool) {
    repo := NewPostgresRepository(pgx)          // Data layer
    service := NewService(repo)                 // Business layer
    RegisterRoutes(mux, *service)               // HTTP layer
}

// In resources.go
func RegisterRoutes(mux *http.ServeMux, db *pgxpool.Pool) {
    project.Init(mux, db)      // Wire project module
    deployment.Init(mux, db)   // Wire deployment module (future)
}
```

### Validation Pattern
```go
// Service validates input
func (s *Service) CreateNewProject(ctx context.Context, p CreateProject) (*Project, error) {
    if err := validator.Validate.Struct(p); err != nil {
        return nil, fmt.Errorf("Validation failed: %w", err)
    }
    // ... rest of logic
}

// Route handler maps validation errors to HTTP response
if fields := response.MapValidationErrors(err); len(fields) > 0 {
    response.Error(w, http.StatusUnprocessableEntity, "Validation failed", fields)
    return
}
```

### Error Handling Pattern
```go
// Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to create project: %w", err)
}

// Respond with appropriate HTTP status and details
response.Error(w, http.StatusInternalServerError, "Operation failed")
```

### Soft Delete Pattern
```sql
-- Query always filters soft-deleted records
WHERE deleted_at IS NULL

-- Delete sets timestamp instead of removing row
UPDATE projects SET deleted_at = NOW() WHERE id = $1
```

### Pagination Patterns

**Offset-Based Pagination**

Use offset-based pagination for **stable datasets** or when total count is essential. Implemented via `request.ParsePaging()` and `response.Collection[T]`.

```go
// domain.go
type GetSingleProject struct {
    ProjectID string `json:"project_id" validate:"required"`
    request.PagingParams  // Embeds PageNumber, PageSize, and Offset() method
}

// postgres.go - Use OFFSET for page-based queries
func (r *PostgresRepository) GetPaged(ctx context.Context, params GetPagedProject) (*response.Collection[*Project], error) {
    offset := params.Offset()  // (PageNumber - 1) * PageSize
    query := `
        WITH projects_cte AS (
            SELECT *, COUNT(*) OVER () AS total_count
            FROM projects
            WHERE deleted_at IS NULL
            ORDER BY created_at DESC
            LIMIT $1 OFFSET $2
        )
        SELECT *, total_count FROM projects_cte
    `
    // ... scan results
    return response.NewCollection(projects, totalCount, params.PageNumber, params.PageSize), nil
}
```

---

## Configuration

### Environment Variables
```bash
APP_ENV=development              # development or production

DATABASE_URL=postgresql://...    # PostgreSQL connection string (required)
DB_MAX_OPEN_CONNS=25            # Connection pool size
DB_CONN_LIFETIME=5m              # Max connection lifetime

REDIS_URL=localhost:6379        # Redis for worker tasks (required)
```

### HTTP Server
- Runs on `:8080`
- Graceful shutdown on SIGTERM/SIGINT (10-second timeout)
- Structured logging via `log/slog`

---

## Validation Rules

### Project Creation (`CreateProject`)
| Field | Rules | Example |
|-------|-------|---------|
| `name` | required, 3-100 chars | "My App" |
| `git_url` | required, valid URL | "https://github.com/user/repo" |
| `git_provider` | required, one of: github, gitlab, bitbucket | "github" |
| `primary_branch` | optional | "main" (defaults to empty) |

### Project Update (`UpdateProject`)
- All fields optional
- Same validation rules apply when provided
- Only non-zero values update the database

### Validation Errors
Error responses include field-level messages:
```json
{
  "status": 422,
  "message": "Validation failed",
  "errors": {
    "name": "Must be at least 3 characters",
    "git_url": "Must be a valid URL"
  }
}
```

---

## Key Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/jackc/pgx/v5` | PostgreSQL driver with connection pooling |
| `github.com/go-playground/validator/v10` | Input validation with tags |
| `github.com/swaggo/swag` | Swagger documentation generation |
| `github.com/golang-migrate/migrate/v4` | Database schema migrations |
| `github.com/caarlos0/env/v10` | Environment variable parsing |
| `github.com/joho/godotenv` | `.env` file loading |

---

## Future Work (Scaffolding in Place)

- **Deployment Module** (`internal/resources/deployment/`) - Routes and service exist, needs full implementation
- **Platform Engine** (`internal/platform/`) - Docker integration, worker pools, scheduler
- **Provider Clients** (`internal/provider/`) - Full GitHub, GitLab, Bitbucket implementations
- **Gateway Security** (`internal/gateway/`) - JWT auth, rate limiting middleware
- **Webhook Handling** - Git provider webhooks for triggering builds

---

## Testing & Documentation

### Swagger UI
Visit `http://localhost:8080/swagger/` to explore the API interactively.

### API Documentation
All routes include Swagger annotations:
```go
// @Summary      Create a project
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param request body CreateProject true "Project Details"
// @Success      201 {object} Project
// @Failure      400 {object} response.ErrorResponse "Bad Request"
// @Router       /projects [post]
```

---

## Common Tasks

### Add a New Route
1. Add handler method in `routes.go`
2. Register in `RegisterRoutes()` with Swagger comments
3. Run `make swagger && make restart`

### Add Validation to a Field
1. Update struct tag in `domain.go`:
   ```go
   type CreateProject struct {
       Name string `json:"name" validate:"required,min=3,max=100"`
   }
   ```
2. Error message auto-maps in `response/response.go`

### Query the Database Directly
```go
// In postgres.go
query := `SELECT ... FROM projects WHERE ...`
err := r.db.QueryRow(ctx, query, params...).Scan(&result)
```

### Create a Background Task
1. Add job to worker pool in `platform/worker/pool.go`
2. Enqueue from service layer
3. Worker processes asynchronously, updates `deployments` status

---

## Quick Reference

| File | Purpose |
|------|---------|
| `cmd/app/main.go` | Application bootstrap |
| `internal/resources/project/domain.go` | Data models and contracts |
| `internal/resources/project/service.go` | Business logic |
| `internal/resources/project/postgres.go` | SQL queries |
| `internal/resources/project/routes.go` | HTTP handlers |
| `pkgs/validator/validator.go` | Global validator instance |
| `pkgs/response/response.go` | HTTP response formatting |
| `migrations/*.sql` | Database schema |

