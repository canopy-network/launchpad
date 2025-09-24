# Chain Management REST API Endpoints

This document defines the REST API endpoints for chain creation and management, corresponding to the multi-step launch process.

## Base URL
```
/api/v1
```

## Authentication
All endpoints require Bearer token authentication via `Authorization: Bearer <token>` header.

---

## Core Chain Operations

### GET /chains
Retrieve all chains with optional filtering and pagination.

**Query Parameters:**
- `status` - Filter by chain status (draft, pending_launch, virtual_active, graduated, failed)
- `created_by` - Filter by creator user ID
- `template_id` - Filter by template ID
- `page` - Page number (default: 1)
- `limit` - Items per page (default: 20, max: 100)
- `include` - Include related data (socials, repositories, assets, virtual_pool)

**Response:** Uses `ApiResponse` with `ChainComplete[]` data and `Pagination`.

### GET /chains/{id}
Retrieve a specific chain with all related data.

**Path Parameters:**
- `id` (uuid) - Chain ID

**Query Parameters:**
- `include` - Include related data (socials, repositories, assets, virtual_pool, transactions)

**Response:** Uses `ApiResponse` with `ChainComplete` data.

### POST /chains
Create a new chain (Step 1: Template Selection + Step 2: Chain Configuration).

**Request Body:** Uses `CreateChainRequest` schema.

**Response:** Uses `ApiResponse` with `ChainComplete` data.

---

## Multi-Step Update Endpoints

### PUT /chains/{id}/development-integration
Update development integration settings (Step 3: Development Integration).

**Request Body:** Uses `UpdateDevelopmentIntegrationRequest` schema.

**Response:** Uses `ApiResponse` with `ChainComplete` data.

### PUT /chains/{id}/documentation
Update project documentation (Step 4: Project Documentation).

**Request Body:** Uses `UpdateDocumentationRequest` schema.

**Response:** Uses `ApiResponse` with `ChainComplete` data.

### PUT /chains/{id}/economics
Update launch economics (Step 5: Launch Economics).

**Request Body:** Uses `UpdateEconomicsRequest` schema.

**Response:** Uses `ApiResponse` with `ChainComplete` data.

### PUT /chains/{id}/scheduling
Update launch scheduling (Step 6: Launch Scheduling).

**Request Body:** Uses `UpdateSchedulingRequest` schema.

**Response:** Uses `ApiResponse` with `ChainComplete` data.

### POST /chains/{id}/launch
Finalize and launch the chain (completes Step 6).

**Request Body:** Uses `LaunchChainRequest` schema.

**Response:** Uses `ApiResponse` with `ChainComplete` data.

---

## Additional Endpoints

### GET /chains/{id}/virtual-pool
Get virtual pool data for a chain.

**Response:** Uses `ApiResponse` with `VirtualPool` data.

### GET /chains/{id}/transactions
Get virtual pool transactions for a chain.

**Query Parameters:**
- `user_id` - Filter by user ID
- `transaction_type` - Filter by type (buy, sell)
- `page` - Page number
- `limit` - Items per page

**Response:** Uses `ApiResponse` with `VirtualPoolTransaction[]` data and `Pagination`.

### DELETE /chains/{id}
Delete a chain (only allowed for draft chains).

**Response:**
```json
{
  "message": "Chain deleted successfully"
}
```

---

## Error Responses

All endpoints may return `ApiError` responses with the following error codes:

- `400` - VALIDATION_ERROR
- `401` - UNAUTHORIZED
- `403` - FORBIDDEN
- `404` - NOT_FOUND
- `409` - CONFLICT
- `422` - BUSINESS_RULE_VIOLATION
- `500` - INTERNAL_ERROR

---

## JSON Schema Reference

All request and response schemas are defined in `jsonschema.json` using JSON Schema 2020-12 specification. Key schemas include:

**Request Schemas:**
- `CreateChainRequest` - For POST /chains
- `UpdateDevelopmentIntegrationRequest` - For PUT /chains/{id}/development-integration
- `UpdateDocumentationRequest` - For PUT /chains/{id}/documentation
- `UpdateEconomicsRequest` - For PUT /chains/{id}/economics
- `UpdateSchedulingRequest` - For PUT /chains/{id}/scheduling
- `LaunchChainRequest` - For POST /chains/{id}/launch

**Response Schemas:**
- `ChainComplete` - Complete chain object with all joined relations
- `VirtualPool` - Virtual pool data
- `VirtualPoolTransaction` - Transaction data
- `ApiResponse` - Standard API response wrapper
- `ApiError` - Standard error response
- `Pagination` - Pagination metadata

**Entity Schemas:**
- `ChainTemplate`
- `User`
- `ChainRepository`
- `ChainSocialLink`
- `ChainAsset`

All schemas include proper validation rules, data types, constraints, and relationship definitions matching the database schema.