# Launchpad API

A comprehensive REST API for blockchain chain creation and management, built with Go and Chi router.

## Features

- **Chain Management**: Full CRUD operations for blockchain chains
- **Multi-Step Launch Process**: Step-by-step chain configuration and deployment
- **Template System**: Pre-built blockchain templates
- **Development Integration**: GitHub repository connections
- **Virtual Pools**: Bonding curve mechanics and trading
- **Asset Management**: Media and documentation handling
- **Social Integration**: Social media links and verification

## Project Structure

```
launchpad/
├── cmd/server/           # Application entry points
├── internal/             # Private application code
│   ├── config/          # Configuration management
│   ├── handlers/        # HTTP request handlers
│   ├── models/          # Data models and DTOs
│   ├── repository/      # Data access layer
│   ├── server/          # HTTP server setup
│   ├── services/        # Business logic
│   └── validators/      # Request validation
├── pkg/                 # Public packages
├── tests/               # Test files
├── migrations/          # Database migrations
├── templates/           # Email templates
└── scripts/             # Development scripts
```

## Quick Start

### Local Development

1. **Configure environment:**
   ```bash
   cp .env.example .env
   ```

2. **Run with Docker Compose:**
   ```bash
   make docker-up
   ```

3. **Apply DB Schema**
   ```bash
   make migrate-up
   ```

4. **Populate Database**
   ```bash
   make load-fixtures
   ```

## Database Migrations

This project uses [Atlas](https://atlasgo.io/) for managing database migrations.

### Migration Commands

```bash
# Apply pending migrations to set up the database schema
make migrate-setup

# Check current migration status
make migrate-status

# Generate a new migration file when you modify schema.sql
make migrate-diff

# Apply pending migrations (same as migrate-setup)
make migrate-up

# Validate migration files for safety
make migrate-validate

# Complete database reset with fresh schema and fixture data
make db-reset
```

### Migration Workflow

1. **Initial Setup:**
   ```bash
   # Start PostgreSQL and apply initial schema
   docker-compose up -d postgres
   make migrate-setup
   make load-fixtures
   ```

2. **Making Schema Changes:**
   ```bash
   # Edit schema.sql with your changes
   vim schema.sql

   # Generate migration file
   make migrate-diff
   # Enter migration name when prompted: e.g., "add_user_preferences"

   # Review the generated migration in migrations/
   # Apply the migration
   make migrate-up
   ```

3. **Development Database Reset:**
   ```bash
   # WARNING: This destroys all data
   echo "yes" | make clear-data
   make migrate-setup
   make load-fixtures
   ```

### Migration Files

Atlas generates and manages migration files in the `migrations/` directory:

```
migrations/
├── 20250928195103_initial_schema.sql    # Generated migration
├── atlas.sum                            # Migration integrity checksums
└── [future migrations...]
```

### Configuration

Migration behavior is configured in `atlas.hcl`:

- **Local Environment**: For development with local PostgreSQL
- **Docker Environment**: For Docker Compose setup
- **Production Environment**: Template for production deployments

### Best Practices

1. **Always generate migrations** instead of applying schema changes directly
2. **Review generated migrations** before applying them
3. **Test migrations** on a copy of production data before deploying
4. **Never modify existing migration files** - create new ones instead
5. **Keep schema.sql as the canonical source** of your database schema

### Database Fixtures

The project includes comprehensive fixture data in the `fixtures/` directory with:
- 4 sample users with different verification levels (`users.sql`)
- 5 chain templates covering different blockchain types (`chain_templates.sql`)
- 5 sample chains in various lifecycle states (`chains.sql`)
- Virtual pools, transactions, and related data (`virtual_pools.sql`, `virtual_pool_transactions.sql`)
- Chain assets, social links, and repository integrations

The fixture files are organized by entity type for easier maintenance. Load all fixtures after applying migrations:
```bash
make load-fixtures
```

## Development

### Available Make Commands

```bash
make help              # Show available commands
make build             # Build the application
make run               # Run the application
make test              # Run tests
make test-coverage     # Run tests with coverage
make fmt               # Format code
make lint              # Run linter
make docker-up         # Start with Docker
make docker-down       # Stop Docker services
```

### Code Structure

The application follows Clean Architecture principles:

- **Handlers** handle HTTP requests and responses
- **Services** contain business logic
- **Repositories** handle data persistence
- **Models** define data structures
- **Validators** handle request validation

### Testing

Run tests:
```bash
make test
```

Run with coverage:
```bash
make test-coverage
```

### JSON Schema Validation

Request/response schemas are defined using JSON Schema 2020-12 in `jsonschema.json`. The application uses Go struct tags for validation.

## Production Deployment

### Build for Production

```bash
make build-prod
```

### Docker Production Build

```bash
docker build -t launchpad-api .
```

### Environment Variables

Ensure these are set in production:

- `DATABASE_URL`: PostgreSQL connection string
- `JWT_SECRET`: Secure JWT signing key (32+ characters)
- `ENVIRONMENT`: Set to "production"
- `GITHUB_CLIENT_ID/SECRET`: For GitHub integration

## Architecture

### Request Flow

```
HTTP Request → Middleware → Router → Handler → Service → Repository → Database
```
