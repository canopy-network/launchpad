# Launchpad API Endpoints

Complete API reference for the Launchpad blockchain creation and management platform.

**Base URL:** `http://localhost:3001`
**API Version:** `v1`
**API Prefix:** `/api/v1`
**JSON Schema:** All requests/responses are validated against JSON Schema 2020-12 (see `jsonschema.json`)

## API Endpoints

### Health & Status

> namespace for general health / liveness

- `GET /health` - Health check endpoint

### Authentication

> namespace for email authentication

- `POST /api/v1/auth/email` - Request email authentication
- `POST /api/v1/auth/verify` - Verify email authentication

### Templates

> namespace for VMless templates for smart contracts

- `GET /api/v1/templates` - Get chain templates

### Chains

> subspace for virtual and graduated blockchain information

- `GET /api/v1/chains` - Get chains list
- `GET /api/v1/chains/{id}` - Get specific chain
- `POST /api/v1/chains` - Create new chain
- `DELETE /api/v1/chains/{id}` - Delete chain
- `GET /api/v1/chains/{id}/transactions` - Get chain transactions

### Virtual Pools

> namespace for pre-graduation (in-database) trading pairs and respective metadata

- `GET /api/v1/virtual-pools` - Get trading information for all pre-graduation chains
- `GET /api/v1/virtual-pools/{id}` - Get trading information for a specific pre-graduation chain

### Graduated Pools

> namespace for graduated (live blockchain) trading pairs and respective metadata

- `GET /api/v1/graduated-pools` - Get a trading information for all graduated chains
- `GET /api/v1/graduated-pools/{id}` - Get a trading information for a specific graduated chains
- `POST /api/v1/graduated-pools/liquidity-deposit` - Deposit liquidity from a graduated chain
- `POST /api/v1/graduated-pools/liquidity-withdrawal` - Remove liquidity from a graduated chain
- `POST /api/v1/graduated-pools/swap` - Execute or simulate an AMM dex swap

### Wallets

> namespace for encrypted keypair management

- `GET /api/v1/wallets` - Get list of wallets with filtering
- `POST /api/v1/wallets` - Create new encrypted wallet
- `GET /api/v1/wallets/{id}` - Get specific wallet details
- `PUT /api/v1/wallets/{id}` - Update wallet metadata
- `DELETE /api/v1/wallets/{id}` - Delete wallet
- `POST /api/v1/wallets/{id}/decrypt` - Decrypt wallet private key
- `POST /api/v1/wallets/{id}/unlock` - Unlock locked wallet

### Bridge 

> namespace for 1-way order book swapping (on-ramp to CNPY)

- `POST /api/v1/bridge/swap` - Execute a 1-way-order book swap

## Table of Contents

- [Authentication](#authentication)
- [Response Format](#response-format)
- [Error Codes](#error-codes)
- [Pagination](#pagination)
- [Endpoints](#endpoints)
  - [Health & Status](#health--status)
  - [Authentication Endpoints](#authentication-endpoints)
  - [User Management](#user-management)
  - [Templates](#templates)
  - [Chains](#chains)
  - [Virtual Pools](#virtual-pools)
  - [Wallets](#wallets)

---

## Authentication

**Current Implementation:** Mock authentication for development

All API endpoints (except `/health`) require authentication via the `X-User-ID` header.

```bash
X-User-ID: 550e8400-e29b-41d4-a716-446655440000
```

**Production:** JWT-based authentication will replace the mock user ID system. The infrastructure is in place but currently uses the X-User-ID header fallback.

---

## Response Format

### Success Response

```json
{
  "data": {
    // Response data
  }
}
```

### Success Response with Pagination

```json
{
  "data": [
    // Array of items
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 42,
    "pages": 3
  }
}
```

### Error Response

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": {}  // Optional additional context
  }
}
```

---

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `VALIDATION_ERROR` | 400 | Request validation failed |
| `BAD_REQUEST` | 400 | Malformed request |
| `UNAUTHORIZED` | 401 | Authentication required |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `CONFLICT` | 409 | Resource conflict (e.g., duplicate name) |
| `UNPROCESSABLE_ENTITY` | 422 | Business rule violation |
| `INTERNAL_ERROR` | 500 | Internal server error |

---

## Pagination

All list endpoints support pagination via query parameters:

- `page` (integer, optional) - Page number (default: 1, min: 1)
- `limit` (integer, optional) - Items per page (default: 20, min: 1, max: 100)

---

## Endpoints

### Health & Status

#### `GET /health`

**Description:** Health check endpoint to verify API availability

**Authentication:** Not Required

**Request Parameters:** None

**Response:**
- **Success (200):**
  ```json
  {
    "data": {
      "status": "healthy",
      "timestamp": "2024-01-15T10:30:00Z",
      "version": "1.0.0"
    }
  }
  ```

**Example Request:**
```bash
curl -X GET http://localhost:3001/health
```

**Notes:**
- No authentication required
- Useful for monitoring and load balancer health checks

---

### Authentication Endpoints

#### `POST /api/v1/auth/email`

**Description:** Sends a 6-digit verification code to the provided email address for authentication

**Authentication:** Not Required

**Request Body:**
```json
{
  "email": "string - valid email address (required)"
}
```

**Validation Constraints (JSON Schema):**
- `email`: Required, valid email format

**Response:**
- **Success (200):**
  ```json
  {
    "data": {
      "message": "Verification code sent successfully",
      "email": "user@example.com",
      "code": "123456"
    }
  }
  ```
  *Note: The `code` field is only present in development mode and will be removed in production.*

- **Error (400):**
  ```json
  {
    "error": {
      "code": "VALIDATION_ERROR",
      "message": "Validation failed",
      "details": [
        {
          "field": "email",
          "message": "must be a valid email address"
        }
      ]
    }
  }
  ```

- **Error (500):**
  ```json
  {
    "error": {
      "code": "INTERNAL_ERROR",
      "message": "Failed to send verification code"
    }
  }
  ```

**Example Request:**
```bash
curl -X POST http://localhost:3001/api/v1/auth/email \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com"
  }'
```

**Notes:**
- Verification code is valid for a limited time (configured server-side)
- In development, the code is returned in the response for testing
- In production, code is only sent via email
- No authentication required (public endpoint)

---

#### `POST /api/v1/auth/verify`

**Description:** Verifies the 6-digit code sent to the user's email address

**Authentication:** Not Required

**Request Body:**
```json
{
  "email": "string - valid email address (required)",
  "code": "string - 6-digit numeric code (required)"
}
```

**Validation Constraints (JSON Schema):**
- `email`: Required, valid email format
- `code`: Required, exactly 6 characters, numeric only

**Response:**
- **Success (200):**
  ```json
  {
    "data": {
      "message": "Email verified successfully",
      "email": "user@example.com"
    }
  }
  ```

- **Error (400) - Invalid Code:**
  ```json
  {
    "error": {
      "code": "BAD_REQUEST",
      "message": "Invalid verification code"
    }
  }
  ```

- **Error (400) - Expired Code:**
  ```json
  {
    "error": {
      "code": "BAD_REQUEST",
      "message": "Verification code has expired"
    }
  }
  ```

- **Error (400) - Validation:**
  ```json
  {
    "error": {
      "code": "VALIDATION_ERROR",
      "message": "Validation failed",
      "details": [
        {
          "field": "code",
          "message": "must be exactly 6 characters"
        }
      ]
    }
  }
  ```

**Example Request:**
```bash
curl -X POST http://localhost:3001/api/v1/auth/verify \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "code": "123456"
  }'
```

**Notes:**
- After successful verification, the user should be created/updated in the system
- In production, this endpoint will return a JWT token for subsequent authenticated requests
- Current implementation returns only a success message (JWT implementation pending)
- No authentication required (public endpoint)

---

### User Management

#### `PUT /api/v1/users/profile`

**Description:** Updates the authenticated user's profile information. All fields are optional, allowing partial updates.

**Authentication:** Required (Session token via cookie or X-User-ID header)

**Request Body:**
```json
{
  "username": "string (optional, 3-50 chars, alphanumeric only)",
  "display_name": "string (optional, max 100 chars)",
  "bio": "string (optional, max 500 chars)",
  "avatar_url": "string (optional, valid URL, max 500 chars)",
  "website_url": "string (optional, valid URL, max 500 chars)",
  "twitter_handle": "string (optional, max 50 chars)",
  "github_username": "string (optional, max 100 chars, alphanumeric only)",
  "telegram_handle": "string (optional, max 50 chars)"
}
```

**Validation Constraints:**
- `username`: Optional, 3-50 characters, alphanumeric only (letters and numbers)
- `display_name`: Optional, maximum 100 characters
- `bio`: Optional, maximum 500 characters
- `avatar_url`: Optional, must be valid URL format, maximum 500 characters
- `website_url`: Optional, must be valid URL format, maximum 500 characters
- `twitter_handle`: Optional, maximum 50 characters
- `github_username`: Optional, maximum 100 characters, alphanumeric only
- `telegram_handle`: Optional, maximum 50 characters

**Response:**
- **Success (200):**
  ```json
  {
    "data": {
      "user": {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "wallet_address": "0x1234567890abcdef...",
        "email": "user@example.com",
        "username": "johndoe",
        "display_name": "John Doe",
        "bio": "Blockchain enthusiast and developer",
        "avatar_url": "https://example.com/avatar.jpg",
        "website_url": "https://johndoe.com",
        "twitter_handle": "@johndoe",
        "github_username": "johndoe",
        "telegram_handle": "@johndoe",
        "is_verified": true,
        "verification_tier": "verified",
        "total_chains_created": 5,
        "total_cnpy_invested": 12500.50,
        "reputation_score": 850,
        "created_at": "2024-01-01T00:00:00Z"
      },
      "message": "Profile updated successfully"
    }
  }
  ```

- **Error (400) - No fields provided:**
  ```json
  {
    "error": {
      "code": "BAD_REQUEST",
      "message": "No fields provided to update"
    }
  }
  ```

- **Error (400) - Validation failed:**
  ```json
  {
    "error": {
      "code": "VALIDATION_ERROR",
      "message": "Validation failed",
      "details": [
        {
          "field": "username",
          "message": "must be between 3 and 50 characters"
        },
        {
          "field": "avatar_url",
          "message": "must be a valid URL"
        }
      ]
    }
  }
  ```

- **Error (401) - Not authenticated:**
  ```json
  {
    "error": {
      "code": "UNAUTHORIZED",
      "message": "User not authenticated"
    }
  }
  ```

- **Error (404) - User not found:**
  ```json
  {
    "error": {
      "code": "NOT_FOUND",
      "message": "User not found"
    }
  }
  ```

- **Error (409) - Username conflict:**
  ```json
  {
    "error": {
      "code": "CONFLICT",
      "message": "Username is already taken"
    }
  }
  ```

- **Error (500) - Server error:**
  ```json
  {
    "error": {
      "code": "INTERNAL_ERROR",
      "message": "Failed to update profile"
    }
  }
  ```

**Example Request:**
```bash
# Update username and bio
curl -X PUT http://localhost:3001/api/v1/users/profile \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{
    "username": "johndoe",
    "bio": "Blockchain enthusiast and developer"
  }'

# Update social links
curl -X PUT http://localhost:3001/api/v1/users/profile \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{
    "twitter_handle": "@johndoe",
    "github_username": "johndoe",
    "website_url": "https://johndoe.com"
  }'

# Update display name and avatar
curl -X PUT http://localhost:3001/api/v1/users/profile \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{
    "display_name": "John Doe",
    "avatar_url": "https://example.com/avatar.jpg"
  }'
```

**User Response Fields:**
- `id` - UUID identifier for the user
- `wallet_address` - User's blockchain wallet address (required, immutable)
- `email` - User's email address (nullable)
- `username` - Unique username (nullable, alphanumeric)
- `display_name` - Display name shown in UI (nullable)
- `bio` - User biography/description (nullable)
- `avatar_url` - URL to user's avatar image (nullable)
- `website_url` - User's website URL (nullable)
- `twitter_handle` - Twitter/X handle (nullable)
- `github_username` - GitHub username (nullable)
- `telegram_handle` - Telegram handle (nullable)
- `is_verified` - Whether user has completed verification
- `verification_tier` - User verification level: `basic`, `verified`, `premium`
- `total_chains_created` - Count of chains created by user
- `total_cnpy_invested` - Total CNPY invested across all chains
- `reputation_score` - User reputation score based on activity
- `created_at` - Account creation timestamp (ISO 8601)

**Notes:**
- User ID is extracted from authentication context (session token or X-User-ID header)
- All fields are optional; only provided fields will be updated
- Username must be unique across all users
- Username and github_username must be alphanumeric (no special characters)
- URLs are validated for proper format
- Empty request body will return "No fields provided to update" error
- The response includes the complete updated user object
- Email field cannot be updated through this endpoint (managed via auth flow)
- Wallet address is immutable and cannot be changed
- Related to User schema in jsonschema.json (`$defs/User`)

---

### Templates

#### `GET /api/v1/templates`

**Description:** Retrieves a paginated list of blockchain templates available for chain creation. Templates provide language-specific starter configurations with pre-configured defaults for token economics.

**Authentication:** Required (X-User-ID header)

**Request Parameters:**
- **Query Parameters:**
  - `category` (string, optional) - Filter by template category
    - Examples: `defi`, `gaming`, `enterprise`, `social`, `general`
    - Max length: 50 characters
    - Valid values: `beginner`, `intermediate`, `advanced`, `expert`
    - Note: This filter still works for backward compatibility but the field is no longer in responses
  - `is_active` (boolean, optional) - Filter by active status
    - Default: all templates (active and inactive)
  - `page` (integer, optional) - Page number
    - Default: 1
    - Min: 1
  - `limit` (integer, optional) - Items per page
    - Default: 20
    - Min: 1
    - Max: 100

**Validation Constraints (JSON Schema):**
- `category`: Optional string, max 50 characters
- `is_active`: Optional boolean
- `page`: Optional integer, minimum 1
- `limit`: Optional integer, minimum 1, maximum 100

**Response:**
- **Success (200):**
  ```json
  {
    "data": [
      {
        "id": "550e8400-e29b-41d4-a716-446655441001",
        "template_name": "Python Example Template",
        "template_description": "High-level blockchain development with Python for rapid prototyping and data science integration",
        "template_category": "general",
        "supported_language": "python",
        "default_token_supply": 1000000000,
        "documentation_url": "https://docs.scanopy.io/templates/python",
        "example_chains": ["Example Chain 1", "Example Chain 2"],
        "version": "1.0.0",
        "is_active": true,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      },
      {
        "id": "550e8400-e29b-41d4-a716-446655441002",
        "template_name": "Golang Example Template",
        "template_description": "High-performance blockchain in Go with excellent concurrency and networking capabilities",
        "template_category": "defi",
        "supported_language": "golang",
        "default_token_supply": 1000000000,
        "documentation_url": "https://docs.scanopy.io/templates/golang",
        "example_chains": ["Example Chain 1", "Example Chain 2"],
        "version": "1.2.0",
        "is_active": true,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 5,
      "pages": 1
    }
  }
  ```

- **Error (400) - Invalid page:**
  ```json
  {
    "error": {
      "code": "VALIDATION_ERROR",
      "message": "Validation failed",
      "details": [
        {
          "field": "page",
          "message": "must be greater than or equal to 1"
        }
      ]
    }
  }
  ```

- **Error (400) - Invalid limit:**
  ```json
  {
    "error": {
      "code": "VALIDATION_ERROR",
      "message": "Validation failed",
      "details": [
        {
          "field": "limit",
          "message": "must be less than or equal to 100"
        }
      ]
    }
  }
  ```

- **Error (500):**
  ```json
  {
    "error": {
      "code": "INTERNAL_ERROR",
      "message": "Failed to retrieve templates"
    }
  }
  ```

**Example Request:**
```bash
# Get all templates with default pagination
curl -X GET "http://localhost:3001/api/v1/templates" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000"

# Filter by category with custom pagination
curl -X GET "http://localhost:3001/api/v1/templates?category=defi&page=1&limit=10" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000"

# Get only active templates
curl -X GET "http://localhost:3001/api/v1/templates?is_active=true&page=1&limit=20" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000"
```

**Template Response Fields:**
- `id` - UUID identifier for the template
- `template_name` - Human-readable template name
- `template_description` - Detailed description of template features
- `template_category` - Category classification (e.g., defi, gaming, enterprise)
- `supported_language` - Programming language (python, golang, typescript, kotlin, csharp)
- `default_token_supply` - Default total token supply for chains using this template
- `documentation_url` - Link to template documentation (nullable)
- `example_chains` - Array of example chain names built with this template
- `version` - Template version string (semantic versioning)
- `is_active` - Whether template is currently available for use
- `created_at` - Template creation timestamp (ISO 8601)
- `updated_at` - Last update timestamp (ISO 8601)

**Notes:**
- Templates are language-specific blockchain starter configurations
- Each template provides a different programming language (Python, Golang, TypeScript, Kotlin, C#)
- Use template ID when creating chains to inherit `default_token_supply` and other defaults
- Templates are ordered by creation date (most recent first)
- All numeric fields are validated according to JSON Schema constraints
- Returns empty array if no templates match the filters

---

### Chains

#### `GET /api/v1/chains`

**Description:** Retrieves a paginated list of chains with optional filtering

**Authentication:** Required (X-User-ID header)

**Request Parameters:**
- **Query Parameters:**
  - `status` (string, optional) - Filter by status: `draft`, `pending_launch`, `virtual_active`, `graduated`, `failed`
  - `created_by` (string, optional) - Filter by creator user ID (UUID format)
  - `template_id` (string, optional) - Filter by template ID (UUID format)
  - `include` (string, optional) - Include related data: `template`, `creator`, `repository`, `social_links`, `assets`, `virtual_pool`
  - `page` (integer, optional) - Page number (default: 1, min: 1)
  - `limit` (integer, optional) - Items per page (default: 20, min: 1, max: 100)

**Response:**
- **Success (200):**
  ```json
  {
    "data": [
      {
        "id": "650e8400-e29b-41d4-a716-446655440001",
        "chain_name": "MyDeFiChain",
        "token_symbol": "MDFC",
        "chain_description": "A decentralized finance protocol",
        "template_id": "550e8400-e29b-41d4-a716-446655440000",
        "consensus_mechanism": "tendermint",
        "token_total_supply": 1000000000,
        "graduation_threshold": 50000.0,
        "creation_fee_cnpy": 100.0,
        "initial_cnpy_reserve": 10000.0,
        "initial_token_supply": 1000000,
        "bonding_curve_slope": 0.0001,
        "scheduled_launch_time": "2024-02-01T00:00:00Z",
        "actual_launch_time": null,
        "creator_initial_purchase_cnpy": 1000.0,
        "status": "draft",
        "is_graduated": false,
        "graduation_time": null,
        "chain_id": null,
        "genesis_hash": null,
        "validator_min_stake": 1000.0,
        "created_by": "550e8400-e29b-41d4-a716-446655440000",
        "created_at": "2024-01-15T10:00:00Z",
        "updated_at": "2024-01-15T10:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 1,
      "pages": 1
    }
  }
  ```

**Example Request:**
```bash
curl -X GET "http://localhost:3001/api/v1/chains?status=draft&include=template&page=1&limit=20" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000"
```

**Notes:**
- Use `include` parameter to eagerly load related entities
- Multiple relationships can be included by repeating the parameter

---

#### `GET /api/v1/chains/{id}`

**Description:** Retrieves a single chain by ID with optional related data

**Authentication:** Required (X-User-ID header)

**Request Parameters:**
- **Path Parameters:**
  - `id` (UUID) - Chain ID

- **Query Parameters:**
  - `include` (string, optional) - Include related data: `template`, `creator`, `repository`, `social_links`, `assets`, `virtual_pool`

**Response:**
- **Success (200):**
  ```json
  {
    "data": {
      "id": "650e8400-e29b-41d4-a716-446655440001",
      "chain_name": "MyDeFiChain",
      "token_symbol": "MDFC",
      "chain_description": "A decentralized finance protocol",
      "template_id": "550e8400-e29b-41d4-a716-446655440000",
      "consensus_mechanism": "tendermint",
      "token_total_supply": 1000000000,
      "graduation_threshold": 50000.0,
      "creation_fee_cnpy": 100.0,
      "initial_cnpy_reserve": 10000.0,
      "initial_token_supply": 1000000,
      "bonding_curve_slope": 0.0001,
      "scheduled_launch_time": "2024-02-01T00:00:00Z",
      "actual_launch_time": null,
      "creator_initial_purchase_cnpy": 1000.0,
      "status": "draft",
      "is_graduated": false,
      "graduation_time": null,
      "chain_id": null,
      "genesis_hash": null,
      "validator_min_stake": 1000.0,
      "created_by": "550e8400-e29b-41d4-a716-446655440000",
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z"
    }
  }
  ```

- **Error (404):**
  ```json
  {
    "error": {
      "code": "NOT_FOUND",
      "message": "Chain not found"
    }
  }
  ```

**Example Request:**
```bash
curl -X GET "http://localhost:3001/api/v1/chains/650e8400-e29b-41d4-a716-446655440001?include=template" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000"
```

**Notes:**
- Returns 404 if chain doesn't exist or user doesn't have access

---

#### `POST /api/v1/chains`

**Description:** Creates a new blockchain chain in draft status

**Authentication:** Required (X-User-ID header)

**Request Body:**
```json
{
  "chain_name": "string (required, 1-100 chars)",
  "token_symbol": "string (required, 1-20 chars, uppercase, pattern: ^[A-Z][A-Z0-9]*$)",
  "chain_description": "string (optional, max 5000 chars)",
  "template_id": "UUID (optional, recommended for pre-configured defaults)",
  "consensus_mechanism": "string (optional, max 50 chars, default: 'nestbft')",
  "token_total_supply": "integer (optional, 1M-1T, default: 1000000000)",
  "graduation_threshold": "float (optional, 1K-10M CNPY, default: 50000.00)",
  "creation_fee_cnpy": "float (optional, min 0, default: 100.00)",
  "initial_cnpy_reserve": "float (optional, min 1000, default: 10000.00)",
  "initial_token_supply": "integer (optional, min 100K, default: 800000000)",
  "bonding_curve_slope": "float (optional, min 0.000000001, default: 0.00000001)",
  "validator_min_stake": "float (optional, min 100, default: 1000.00)",
  "creator_initial_purchase_cnpy": "float (optional, min 0, default: 0)"
}
```

**Example Request Body:**
```json
{
  "chain_name": "MyDeFiChain",
  "token_symbol": "MDFC",
  "chain_description": "A revolutionary DeFi protocol",
  "template_id": "550e8400-e29b-41d4-a716-446655440000",
  "token_total_supply": 1000000000,
  "graduation_threshold": 50000.0,
  "initial_cnpy_reserve": 10000.0,
  "creator_initial_purchase_cnpy": 1000.0
}
```

**Response:**
- **Success (201):**
  ```json
  {
    "data": {
      "id": "650e8400-e29b-41d4-a716-446655440001",
      "chain_name": "MyDeFiChain",
      "token_symbol": "MDFC",
      "status": "draft",
      "created_at": "2024-01-15T10:00:00Z",
      // ... full chain object
    }
  }
  ```

- **Error (400):**
  ```json
  {
    "error": {
      "code": "VALIDATION_ERROR",
      "message": "Validation failed",
      "details": [
        {
          "field": "token_symbol",
          "message": "must contain only uppercase letters"
        }
      ]
    }
  }
  ```

- **Error (409):**
  ```json
  {
    "error": {
      "code": "CONFLICT",
      "message": "Chain name already exists"
    }
  }
  ```

**Example Request:**
```bash
curl -X POST http://localhost:3001/api/v1/chains \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{
    "chain_name": "MyDeFiChain",
    "token_symbol": "MDFC",
    "template_id": "550e8400-e29b-41d4-a716-446655440000"
  }'
```

**Notes:**
- Chain is created in `draft` status
- Template ID is optional but recommended for pre-configured defaults
- Token symbol must be uppercase
- All chain configuration is done in a single request
- An encrypted keypair is automatically generated for the chain

---

#### `DELETE /api/v1/chains/{id}`

**Description:** Deletes a chain (only allowed in draft status)

**Authentication:** Required (X-User-ID header)

**Request Parameters:**
- **Path Parameters:**
  - `id` (UUID) - Chain ID

**Response:**
- **Success (200):**
  ```json
  {
    "data": {
      "message": "Chain deleted successfully"
    }
  }
  ```

- **Error (403):**
  ```json
  {
    "error": {
      "code": "FORBIDDEN",
      "message": "Access denied"
    }
  }
  ```

- **Error (404):**
  ```json
  {
    "error": {
      "code": "NOT_FOUND",
      "message": "Chain not found"
    }
  }
  ```

**Example Request:**
```bash
curl -X DELETE http://localhost:3001/api/v1/chains/650e8400-e29b-41d4-a716-446655440001 \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000"
```

**Notes:**
- Only chain creator can delete
- Chain must be in `draft` status
- Cannot delete launched or active chains

---

### Virtual Pools

#### `GET /api/v1/virtual-pools`

**Description:** Retrieves a paginated list of all virtual pools across all chains

**Authentication:** Required (X-User-ID header)

**Request Parameters:**
- **Query Parameters:**
  - `page` (integer, optional) - Page number
    - Default: 1
    - Min: 1
  - `limit` (integer, optional) - Items per page
    - Default: 20
    - Min: 1
    - Max: 100

**Response:**
- **Success (200):**
  ```json
  {
    "data": [
      {
        "id": "750e8400-e29b-41d4-a716-446655440002",
        "chain_id": "650e8400-e29b-41d4-a716-446655440001",
        "cnpy_reserve": 15000.0,
        "token_reserve": 950000,
        "current_price_cnpy": 0.015789,
        "market_cap_usd": 15789.47,
        "total_volume_cnpy": 5000.0,
        "total_transactions": 42,
        "unique_traders": 15,
        "is_active": true,
        "price_24h_change_percent": 5.23,
        "volume_24h_cnpy": 1200.0,
        "high_24h_cnpy": 0.016500,
        "low_24h_cnpy": 0.015000,
        "created_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-01-15T14:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 5,
      "pages": 1
    }
  }
  ```

- **Error (400):**
  ```json
  {
    "error": {
      "code": "VALIDATION_ERROR",
      "message": "Validation failed",
      "details": [
        {
          "field": "limit",
          "message": "must be between 1 and 100"
        }
      ]
    }
  }
  ```

- **Error (500):**
  ```json
  {
    "error": {
      "code": "INTERNAL_ERROR",
      "message": "Failed to retrieve virtual pools"
    }
  }
  ```

**Example Request:**
```bash
curl -X GET "http://localhost:3001/api/v1/virtual-pools?page=1&limit=20" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000"
```

**Notes:**
- Returns all virtual pools across all chains
- Virtual pools are created when chains launch
- Ordered by most recently updated first
- Use for market overview and discovery
- For chain-specific pool data, use `GET /api/v1/chains/{id}/virtual-pool`

**Response Schema (JSON Schema):**
- Each pool object conforms to `VirtualPool` schema in jsonschema.json
- All fields are required as defined in the schema
- Numeric fields use appropriate precision for financial calculations

---

#### `GET /api/v1/chains/{id}/virtual-pool`

**Description:** Retrieves virtual pool state and trading metrics for a chain

**Authentication:** Required (X-User-ID header)

**Request Parameters:**
- **Path Parameters:**
  - `id` (UUID) - Chain ID

**Response:**
- **Success (200):**
  ```json
  {
    "data": {
      "id": "750e8400-e29b-41d4-a716-446655440002",
      "chain_id": "650e8400-e29b-41d4-a716-446655440001",
      "cnpy_reserve": 15000.0,
      "token_reserve": 950000,
      "current_price_cnpy": 0.015789,
      "market_cap_usd": 15789.47,
      "total_volume_cnpy": 5000.0,
      "total_transactions": 42,
      "unique_traders": 15,
      "is_active": true,
      "price_24h_change_percent": 5.23,
      "volume_24h_cnpy": 1200.0,
      "high_24h_cnpy": 0.016500,
      "low_24h_cnpy": 0.015000,
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T14:30:00Z"
    }
  }
  ```

- **Error (404):**
  ```json
  {
    "error": {
      "code": "NOT_FOUND",
      "message": "Chain not found"
    }
  }
  ```

**Example Request:**
```bash
curl -X GET http://localhost:3001/api/v1/chains/650e8400-e29b-41d4-a716-446655440001/virtual-pool \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000"
```

**Notes:**
- Virtual pool is created when chain is launched
- Implements bonding curve pricing mechanism
- Price updates with each trade
- 24-hour metrics refresh periodically

---

#### `GET /api/v1/chains/{id}/transactions`

**Description:** Retrieves paginated transaction history for a chain's virtual pool

**Authentication:** Required (X-User-ID header)

**Request Parameters:**
- **Path Parameters:**
  - `id` (UUID) - Chain ID

- **Query Parameters:**
  - `user_id` (string, optional) - Filter by user ID (UUID format)
  - `transaction_type` (string, optional) - Filter by type: `buy`, `sell`
  - `page` (integer, optional) - Page number (default: 1, min: 1)
  - `limit` (integer, optional) - Items per page (default: 20, min: 1, max: 100)

**Response:**
- **Success (200):**
  ```json
  {
    "data": [
      {
        "id": "850e8400-e29b-41d4-a716-446655440003",
        "virtual_pool_id": "750e8400-e29b-41d4-a716-446655440002",
        "chain_id": "650e8400-e29b-41d4-a716-446655440001",
        "user_id": "550e8400-e29b-41d4-a716-446655440000",
        "transaction_type": "buy",
        "cnpy_amount": 1000.0,
        "token_amount": 62500,
        "price_per_token_cnpy": 0.016000,
        "trading_fee_cnpy": 10.0,
        "slippage_percent": 0.5,
        "transaction_hash": "0xabcd...",
        "block_height": 12345,
        "gas_used": 21000,
        "pool_cnpy_reserve_after": 11000.0,
        "pool_token_reserve_after": 937500,
        "market_cap_after_usd": 11734.38,
        "created_at": "2024-01-15T12:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 42,
      "pages": 3
    }
  }
  ```

**Example Request:**
```bash
curl -X GET "http://localhost:3001/api/v1/chains/650e8400-e29b-41d4-a716-446655440001/transactions?transaction_type=buy&page=1&limit=20" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000"
```

**Notes:**
- Returns all transactions for the chain's virtual pool
- Filter by user_id to see a specific user's trades
- Transactions ordered by most recent first
- Includes pool state snapshot after each transaction

---

### Wallets

#### `GET /api/v1/wallets`

**Description:** Retrieves a paginated list of wallets with optional filtering by user, chain, lock status, and active status

**Authentication:** Required (Session token or X-User-ID header)

**Request Parameters:**
- **Query Parameters:**
  - `user_id` (string, optional) - Filter by user ID (UUID format)
  - `chain_id` (string, optional) - Filter by chain ID (UUID format)
  - `created_by` (string, optional) - Filter by creator user ID (UUID format)
  - `is_active` (boolean, optional) - Filter by active status (true/false)
  - `is_locked` (boolean, optional) - Filter by lock status (true/false)
  - `page` (integer, optional) - Page number
    - Default: 1
    - Min: 1
  - `limit` (integer, optional) - Items per page
    - Default: 20
    - Min: 1
    - Max: 100

**Validation Constraints:**
- `user_id`: Optional, must be valid UUID format
- `chain_id`: Optional, must be valid UUID format
- `created_by`: Optional, must be valid UUID format
- `is_active`: Optional boolean
- `is_locked`: Optional boolean
- `page`: Optional integer, minimum 1
- `limit`: Optional integer, minimum 1, maximum 100

**Response:**
- **Success (200):**
  ```json
  {
    "data": [
      {
        "id": "850e8400-e29b-41d4-a716-446655440001",
        "user_id": "550e8400-e29b-41d4-a716-446655440000",
        "chain_id": "650e8400-e29b-41d4-a716-446655440001",
        "address": "0x1234567890abcdef1234567890abcdef12345678",
        "public_key": "04abcdef...",
        "wallet_name": "My Main Wallet",
        "wallet_description": "Primary trading wallet for DeFi operations",
        "is_active": true,
        "is_locked": false,
        "last_used_at": "2024-01-15T14:30:00Z",
        "password_changed_at": "2024-01-10T10:00:00Z",
        "locked_until": null,
        "created_by": "550e8400-e29b-41d4-a716-446655440000",
        "created_at": "2024-01-10T10:00:00Z",
        "updated_at": "2024-01-15T14:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 5,
      "pages": 1
    }
  }
  ```

- **Error (400) - Invalid UUID:**
  ```json
  {
    "error": {
      "code": "BAD_REQUEST",
      "message": "Invalid user_id format"
    }
  }
  ```

- **Error (400) - Validation:**
  ```json
  {
    "error": {
      "code": "VALIDATION_ERROR",
      "message": "Validation failed",
      "details": [
        {
          "field": "limit",
          "message": "must be between 1 and 100"
        }
      ]
    }
  }
  ```

- **Error (500):**
  ```json
  {
    "error": {
      "code": "INTERNAL_ERROR",
      "message": "Failed to retrieve wallets"
    }
  }
  ```

**Example Request:**
```bash
# Get all wallets with default pagination
curl -X GET "http://localhost:3001/api/v1/wallets" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000"

# Filter by user and active status
curl -X GET "http://localhost:3001/api/v1/wallets?user_id=550e8400-e29b-41d4-a716-446655440000&is_active=true&page=1&limit=10" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000"

# Filter by chain
curl -X GET "http://localhost:3001/api/v1/wallets?chain_id=650e8400-e29b-41d4-a716-446655440001" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000"

# Find locked wallets
curl -X GET "http://localhost:3001/api/v1/wallets?is_locked=true" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000"
```

**Wallet Response Fields:**
- `id` - UUID identifier for the wallet
- `user_id` - Associated user ID (nullable)
- `chain_id` - Associated chain ID (nullable)
- `address` - Blockchain wallet address derived from public key
- `public_key` - Public key in hex format
- `wallet_name` - Human-readable wallet name (nullable)
- `wallet_description` - Wallet description or notes (nullable)
- `is_active` - Whether wallet is currently active
- `is_locked` - Whether wallet is locked due to failed decrypt attempts
- `last_used_at` - Timestamp of last wallet usage (nullable)
- `password_changed_at` - Timestamp of last password change (nullable)
- `locked_until` - Timestamp when wallet will automatically unlock (nullable)
- `created_by` - User ID of wallet creator (nullable)
- `created_at` - Wallet creation timestamp
- `updated_at` - Last update timestamp

**Notes:**
- Encrypted private key, salt, and failed_decrypt_attempts are never exposed in API responses
- Wallets can be associated with a user, a chain, both, or neither
- Multiple filters can be combined for precise queries
- `locked_until` indicates temporary lock duration; null means unlocked or permanently locked
- Ordered by most recently updated first
- Returns empty array if no wallets match the filters

---

#### `POST /api/v1/wallets`

**Description:** Creates a new encrypted wallet with a generated keypair. The private key is encrypted using the provided password.

**Authentication:** Required (Session token or X-User-ID header)

**Request Body:**
```json
{
  "password": "string (required, min 8 characters)",
  "user_id": "UUID string (optional)",
  "chain_id": "UUID string (optional)",
  "wallet_name": "string (optional, 1-100 chars)",
  "wallet_description": "string (optional, max 500 chars)"
}
```

**Validation Constraints:**
- `password`: Required, minimum 8 characters
- `user_id`: Optional, must be valid UUID format if provided
- `chain_id`: Optional, must be valid UUID format if provided
- `wallet_name`: Optional, 1-100 characters if provided
- `wallet_description`: Optional, maximum 500 characters if provided

**Response:**
- **Success (201):**
  ```json
  {
    "data": {
      "id": "850e8400-e29b-41d4-a716-446655440001",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "chain_id": "650e8400-e29b-41d4-a716-446655440001",
      "address": "0x1234567890abcdef1234567890abcdef12345678",
      "public_key": "04abcdef...",
      "wallet_name": "My Main Wallet",
      "wallet_description": "Primary trading wallet",
      "is_active": true,
      "is_locked": false,
      "last_used_at": null,
      "password_changed_at": "2024-01-15T10:00:00Z",
      "locked_until": null,
      "created_by": "550e8400-e29b-41d4-a716-446655440000",
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z"
    }
  }
  ```

- **Error (400) - Invalid JSON:**
  ```json
  {
    "error": {
      "code": "BAD_REQUEST",
      "message": "Invalid JSON payload"
    }
  }
  ```

- **Error (400) - Validation:**
  ```json
  {
    "error": {
      "code": "VALIDATION_ERROR",
      "message": "Validation failed",
      "details": [
        {
          "field": "password",
          "message": "must be at least 8 characters"
        },
        {
          "field": "user_id",
          "message": "must be a valid UUID"
        }
      ]
    }
  }
  ```

- **Error (401):**
  ```json
  {
    "error": {
      "code": "UNAUTHORIZED",
      "message": "Authentication required"
    }
  }
  ```

- **Error (500):**
  ```json
  {
    "error": {
      "code": "INTERNAL_ERROR",
      "message": "Failed to create wallet"
    }
  }
  ```

**Example Request:**
```bash
# Create wallet for a user
curl -X POST http://localhost:3001/api/v1/wallets \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{
    "password": "SecurePass123!",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "wallet_name": "My Main Wallet",
    "wallet_description": "Primary trading wallet for DeFi operations"
  }'

# Create wallet for a chain
curl -X POST http://localhost:3001/api/v1/wallets \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{
    "password": "ChainWallet123!",
    "chain_id": "650e8400-e29b-41d4-a716-446655440001",
    "wallet_name": "Chain Validator Wallet"
  }'

# Create standalone wallet
curl -X POST http://localhost:3001/api/v1/wallets \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{
    "password": "StandalonePass123!"
  }'
```

**Notes:**
- A new Ed25519 keypair is automatically generated
- Private key is encrypted using PBKDF2 with user's password and a random salt
- Public key and address are derived from the generated keypair
- The encrypted private key and salt are stored but never returned in API responses
- `created_by` is automatically set to the authenticated user
- Wallet starts in active, unlocked state
- Password must be stored securely by the client; it cannot be recovered
- At least 8 character password recommended for security

---

#### `GET /api/v1/wallets/{id}`

**Description:** Retrieves details for a specific wallet by ID

**Authentication:** Required (Session token or X-User-ID header)

**Request Parameters:**
- **Path Parameters:**
  - `id` (UUID) - Wallet ID

**Response:**
- **Success (200):**
  ```json
  {
    "data": {
      "id": "850e8400-e29b-41d4-a716-446655440001",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "chain_id": "650e8400-e29b-41d4-a716-446655440001",
      "address": "0x1234567890abcdef1234567890abcdef12345678",
      "public_key": "04abcdef...",
      "wallet_name": "My Main Wallet",
      "wallet_description": "Primary trading wallet",
      "is_active": true,
      "is_locked": false,
      "last_used_at": "2024-01-15T14:30:00Z",
      "password_changed_at": "2024-01-10T10:00:00Z",
      "locked_until": null,
      "created_by": "550e8400-e29b-41d4-a716-446655440000",
      "created_at": "2024-01-10T10:00:00Z",
      "updated_at": "2024-01-15T14:30:00Z"
    }
  }
  ```

- **Error (400):**
  ```json
  {
    "error": {
      "code": "BAD_REQUEST",
      "message": "Invalid wallet ID"
    }
  }
  ```

- **Error (404):**
  ```json
  {
    "error": {
      "code": "NOT_FOUND",
      "message": "Wallet not found"
    }
  }
  ```

- **Error (500):**
  ```json
  {
    "error": {
      "code": "INTERNAL_ERROR",
      "message": "Failed to retrieve wallet"
    }
  }
  ```

**Example Request:**
```bash
curl -X GET http://localhost:3001/api/v1/wallets/850e8400-e29b-41d4-a716-446655440001 \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000"
```

**Notes:**
- Wallet ID must be a valid UUID
- Encrypted private key is never exposed in the response
- Use `POST /api/v1/wallets/{id}/decrypt` to retrieve the decrypted private key

---

#### `PUT /api/v1/wallets/{id}`

**Description:** Updates wallet metadata (name, description, active status). Does not modify cryptographic keys or password.

**Authentication:** Required (Session token or X-User-ID header)

**Request Parameters:**
- **Path Parameters:**
  - `id` (UUID) - Wallet ID

**Request Body:**
```json
{
  "wallet_name": "string (optional, 1-100 chars)",
  "wallet_description": "string (optional, max 500 chars)",
  "is_active": "boolean (optional)"
}
```

**Validation Constraints:**
- `wallet_name`: Optional, 1-100 characters if provided
- `wallet_description`: Optional, maximum 500 characters if provided
- `is_active`: Optional boolean

**Response:**
- **Success (200):**
  ```json
  {
    "data": {
      "id": "850e8400-e29b-41d4-a716-446655440001",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "chain_id": "650e8400-e29b-41d4-a716-446655440001",
      "address": "0x1234567890abcdef1234567890abcdef12345678",
      "public_key": "04abcdef...",
      "wallet_name": "Updated Wallet Name",
      "wallet_description": "Updated description",
      "is_active": false,
      "is_locked": false,
      "last_used_at": "2024-01-15T14:30:00Z",
      "password_changed_at": "2024-01-10T10:00:00Z",
      "locked_until": null,
      "created_by": "550e8400-e29b-41d4-a716-446655440000",
      "created_at": "2024-01-10T10:00:00Z",
      "updated_at": "2024-01-15T15:00:00Z"
    }
  }
  ```

- **Error (400) - Invalid JSON:**
  ```json
  {
    "error": {
      "code": "BAD_REQUEST",
      "message": "Invalid JSON payload"
    }
  }
  ```

- **Error (400) - Validation:**
  ```json
  {
    "error": {
      "code": "VALIDATION_ERROR",
      "message": "Validation failed",
      "details": [
        {
          "field": "wallet_name",
          "message": "must be between 1 and 100 characters"
        }
      ]
    }
  }
  ```

- **Error (404):**
  ```json
  {
    "error": {
      "code": "NOT_FOUND",
      "message": "Wallet not found"
    }
  }
  ```

- **Error (500):**
  ```json
  {
    "error": {
      "code": "INTERNAL_ERROR",
      "message": "Failed to update wallet"
    }
  }
  ```

**Example Request:**
```bash
# Update wallet name and description
curl -X PUT http://localhost:3001/api/v1/wallets/850e8400-e29b-41d4-a716-446655440001 \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{
    "wallet_name": "Updated Main Wallet",
    "wallet_description": "Updated description for my primary wallet"
  }'

# Deactivate wallet
curl -X PUT http://localhost:3001/api/v1/wallets/850e8400-e29b-41d4-a716-446655440001 \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{
    "is_active": false
  }'
```

**Notes:**
- All fields are optional; only provided fields will be updated
- Does not modify cryptographic keys, password, or user/chain associations
- Setting `is_active` to false prevents wallet from being used but retains data
- Returns the complete updated wallet object
- The `updated_at` timestamp is automatically updated

---

#### `DELETE /api/v1/wallets/{id}`

**Description:** Permanently deletes a wallet and its encrypted private key. This operation cannot be undone.

**Authentication:** Required (Session token or X-User-ID header)

**Request Parameters:**
- **Path Parameters:**
  - `id` (UUID) - Wallet ID

**Response:**
- **Success (200):**
  ```json
  {
    "data": {
      "message": "Wallet deleted successfully"
    }
  }
  ```

- **Error (400):**
  ```json
  {
    "error": {
      "code": "BAD_REQUEST",
      "message": "Invalid wallet ID"
    }
  }
  ```

- **Error (404):**
  ```json
  {
    "error": {
      "code": "NOT_FOUND",
      "message": "Wallet not found"
    }
  }
  ```

- **Error (500):**
  ```json
  {
    "error": {
      "code": "INTERNAL_ERROR",
      "message": "Failed to delete wallet"
    }
  }
  ```

**Example Request:**
```bash
curl -X DELETE http://localhost:3001/api/v1/wallets/850e8400-e29b-41d4-a716-446655440001 \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000"
```

**Notes:**
- **DANGER**: This operation permanently deletes the wallet and encrypted private key
- Deletion cannot be undone; ensure you have backed up the private key if needed
- Recommend decrypting and exporting private key before deletion
- Associated transactions and history are retained but wallet cannot be restored

---

#### `POST /api/v1/wallets/{id}/decrypt`

**Description:** Decrypts the wallet's private key using the provided password. Returns the decrypted private key for immediate use. Failed attempts increment the failure counter and may lock the wallet.

**Authentication:** Required (Session token or X-User-ID header)

**Request Parameters:**
- **Path Parameters:**
  - `id` (UUID) - Wallet ID

**Request Body:**
```json
{
  "password": "string (required)"
}
```

**Validation Constraints:**
- `password`: Required string

**Response:**
- **Success (200):**
  ```json
  {
    "data": {
      "id": "850e8400-e29b-41d4-a716-446655440001",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "chain_id": "650e8400-e29b-41d4-a716-446655440001",
      "address": "0x1234567890abcdef1234567890abcdef12345678",
      "public_key": "04abcdef...",
      "private_key": "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
      "wallet_name": "My Main Wallet",
      "wallet_description": "Primary trading wallet",
      "is_active": true,
      "is_locked": false,
      "last_used_at": "2024-01-15T15:30:00Z",
      "password_changed_at": "2024-01-10T10:00:00Z",
      "locked_until": null,
      "created_by": "550e8400-e29b-41d4-a716-446655440000",
      "created_at": "2024-01-10T10:00:00Z",
      "updated_at": "2024-01-15T15:30:00Z"
    }
  }
  ```

- **Error (400) - Invalid JSON:**
  ```json
  {
    "error": {
      "code": "BAD_REQUEST",
      "message": "Invalid JSON payload"
    }
  }
  ```

- **Error (400) - Validation:**
  ```json
  {
    "error": {
      "code": "VALIDATION_ERROR",
      "message": "Validation failed",
      "details": [
        {
          "field": "password",
          "message": "is required"
        }
      ]
    }
  }
  ```

- **Error (401) - Wrong Password:**
  ```json
  {
    "error": {
      "code": "UNAUTHORIZED",
      "message": "Invalid password"
    }
  }
  ```

- **Error (403) - Wallet Locked:**
  ```json
  {
    "error": {
      "code": "FORBIDDEN",
      "message": "Wallet is locked due to too many failed attempts"
    }
  }
  ```

- **Error (404):**
  ```json
  {
    "error": {
      "code": "NOT_FOUND",
      "message": "Wallet not found"
    }
  }
  ```

- **Error (500):**
  ```json
  {
    "error": {
      "code": "INTERNAL_ERROR",
      "message": "Failed to decrypt wallet"
    }
  }
  ```

**Example Request:**
```bash
curl -X POST http://localhost:3001/api/v1/wallets/850e8400-e29b-41d4-a716-446655440001/decrypt \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{
    "password": "SecurePass123!"
  }'
```

**Security Notes:**
- **CRITICAL**: The decrypted private key is returned in the response - handle with extreme care
- Private key should only be used immediately and never logged or stored
- Failed decrypt attempts are tracked; excessive failures lock the wallet
- Wallet may lock after multiple consecutive failed attempts (security feature)
- Use TLS/HTTPS in production to protect private key in transit
- Consider implementing rate limiting on this endpoint
- `last_used_at` timestamp is updated on successful decryption
- Clear private key from memory immediately after use in client applications

**Lock Behavior:**
- Failed attempts increment `failed_decrypt_attempts` counter
- After threshold (typically 5 attempts), wallet becomes locked
- Locked wallet sets `is_locked` to true and may set `locked_until` timestamp
- Use `POST /api/v1/wallets/{id}/unlock` to manually unlock

---

#### `POST /api/v1/wallets/{id}/unlock`

**Description:** Manually unlocks a wallet that was locked due to too many failed decrypt attempts. Resets the failed attempt counter.

**Authentication:** Required (Session token or X-User-ID header)

**Request Parameters:**
- **Path Parameters:**
  - `id` (UUID) - Wallet ID

**Request Body:** None

**Response:**
- **Success (200):**
  ```json
  {
    "data": {
      "message": "Wallet unlocked successfully"
    }
  }
  ```

- **Error (400):**
  ```json
  {
    "error": {
      "code": "BAD_REQUEST",
      "message": "Invalid wallet ID"
    }
  }
  ```

- **Error (404):**
  ```json
  {
    "error": {
      "code": "NOT_FOUND",
      "message": "Wallet not found"
    }
  }
  ```

- **Error (500):**
  ```json
  {
    "error": {
      "code": "INTERNAL_ERROR",
      "message": "Failed to unlock wallet"
    }
  }
  ```

**Example Request:**
```bash
curl -X POST http://localhost:3001/api/v1/wallets/850e8400-e29b-41d4-a716-446655440001/unlock \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000"
```

**Notes:**
- Resets `is_locked` to false
- Clears `locked_until` timestamp
- Resets `failed_decrypt_attempts` counter to 0
- Only the wallet owner or admin should have access to this operation
- Consider additional verification (e.g., email confirmation) before unlocking in production
- Does not require password verification (intentional for account recovery)

---

## Chain Lifecycle

Chains progress through the following statuses:

1. **`draft`** - Initial creation, all configuration provided at creation time
   - Complete chain configuration in single `POST /api/v1/chains` request
   - Can be deleted
   - Encrypted keypair automatically generated

2. **`pending_launch`** - Ready for launch, awaiting payment
   - All configuration complete from creation
   - Scheduled launch time set
   - Cannot modify most settings

3. **`virtual_active`** - Launched and trading in virtual pool
   - Virtual pool created
   - Trading enabled via bonding curve
   - Cannot modify core settings

4. **`graduated`** - Moved to external exchange
   - Graduation threshold reached
   - Virtual pool closed
   - Trading on external DEX

5. **`failed`** - Launch or operation failed
   - Requires investigation
   - May be deleted by admin

---

## Common HTTP Status Codes

- `200 OK` - Successful GET/PUT/DELETE request
- `201 Created` - Successful POST request creating a resource
- `400 Bad Request` - Invalid JSON or malformed request
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - Authenticated but not authorized
- `404 Not Found` - Resource doesn't exist
- `409 Conflict` - Resource conflict (duplicate name, etc.)
- `422 Unprocessable Entity` - Business rule violation
- `500 Internal Server Error` - Server error

---

## Rate Limiting

**Not currently implemented.** Future versions will include:
- Rate limiting per user/IP
- Different limits for authenticated vs. unauthenticated requests
- Rate limit headers in responses

---

## Versioning

**Current Version:** v1

API version is included in the URL path (`/api/v1/`). Future versions will be released as `/api/v2/`, etc., with v1 maintained for backward compatibility.

---

## Support

For API support and bug reports:
- GitHub Issues: https://github.com/enielson/launchpad/issues
- Documentation: See CLAUDE.md for development setup
- Order Processing: See ORDER.md for trading mechanics
