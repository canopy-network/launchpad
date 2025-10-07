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
├── api/                 # API documentation
├── tests/               # Test files
├── migrations/          # Database migrations
├── configs/             # Configuration files
└── scripts/             # Development scripts
```

## API Endpoints

### Core Chain Operations

- `GET /api/v1/chains` - List chains with filtering and pagination
- `GET /api/v1/chains/{id}` - Get specific chain with relations
- `POST /api/v1/chains` - Create new chain
- `DELETE /api/v1/chains/{id}` - Delete chain (draft only)

### Multi-Step Update Endpoints

- `PUT /api/v1/chains/{id}/development-integration` - Step 3: Development Integration
- `PUT /api/v1/chains/{id}/documentation` - Step 4: Project Documentation
- `PUT /api/v1/chains/{id}/economics` - Step 5: Launch Economics
- `PUT /api/v1/chains/{id}/scheduling` - Step 6: Launch Scheduling
- `POST /api/v1/chains/{id}/launch` - Finalize chain launch

### Additional Endpoints

- `GET /api/v1/chains/{id}/virtual-pool` - Get virtual pool data
- `GET /api/v1/chains/{id}/transactions` - Get trading transactions
- `GET /health` - Health check

## Quick Start

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 12+
- Docker and Docker Compose (optional)

### Local Development

1. **Clone and setup:**
   ```bash
   git clone <repository-url>
   cd launchpad
   make setup
   ```

2. **Configure environment:**
   ```bash
   cp .env.example .env
   # Edit .env with your database and JWT secret
   ```

3. **Run with Docker Compose:**
   ```bash
   make docker-up
   ```

4. **Or run locally:**
   ```bash
   # Ensure PostgreSQL is running with the schema loaded
   make run
   ```

### Using Docker

**Build and start all services:**
```bash
docker-compose up -d
```

**View logs:**
```bash
docker-compose logs -f
```

**Stop services:**
```bash
docker-compose down
```

## Configuration

The application is configured through environment variables:

```bash
# Server
PORT=3000
ENVIRONMENT=development

# Database
DATABASE_URL=postgres://user:pass@localhost:5432/launchpad

# Security
JWT_SECRET=your-secret-key
JWT_EXPIRATION_HOURS=24

# External Services
GITHUB_CLIENT_ID=your-client-id
GITHUB_CLIENT_SECRET=your-client-secret
```

See `.env.example` for all available options.

## Database Schema

The application uses PostgreSQL with Atlas for database migration management. The schema includes:

- **chains**: Main chain entities
- **chain_templates**: Pre-built templates
- **users**: User accounts
- **chain_repositories**: GitHub integrations
- **chain_social_links**: Social media connections
- **chain_assets**: Media and files
- **virtual_pools**: Trading pools
- **virtual_pool_transactions**: Trading history

## Database Migrations

This project uses [Atlas](https://atlasgo.io/) for managing database migrations in a professional, versioned manner.

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

## API Usage Examples

### Create a Chain

```bash
curl -X POST http://localhost:3000/api/v1/chains \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{
    "chain_name": "MyChain",
    "token_symbol": "MYC",
    "template_id": "template-uuid",
    "chain_description": "My awesome blockchain"
  }'
```

### Update Development Integration

```bash
curl -X PUT http://localhost:3000/api/v1/chains/{id}/development-integration \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user-uuid" \
  -d '{
    "repository": {
      "github_url": "https://github.com/owner/repo",
      "repository_name": "repo",
      "repository_owner": "owner"
    }
  }'
```

### Launch Chain

```bash
curl -X POST http://localhost:3000/api/v1/chains/{id}/launch \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user-uuid" \
  -d '{
    "payment_confirmation": {
      "transaction_hash": "0x1234...",
      "amount_cnpy": 100.0
    }
  }'
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

### Key Components

1. **Chi Router**: HTTP routing and middleware
2. **PostgreSQL**: Primary database with sqlx
3. **Validator**: Request validation with struct tags
4. **CORS**: Cross-origin resource sharing
5. **Graceful Shutdown**: Clean server termination

### Security

- JWT authentication (placeholder implementation)
- Input validation on all endpoints
- SQL injection prevention with parameterized queries
- CORS configuration
- Request timeouts

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Run `make fmt lint test`
6. Submit a pull request

## License

[Add your license here]

## Support

For support, please open an issue in the GitHub repository.