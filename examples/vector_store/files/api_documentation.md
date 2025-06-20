# API Documentation

## Overview
This document outlines the API endpoints available in our service. The API follows RESTful principles and communicates using JSON.

## Base URL
All API endpoints are relative to: `https://api.example.com/v1`

## Authentication
All API requests require authentication. Include your API key in the header:
```
Authorization: Bearer YOUR_API_KEY
```

## Endpoints

### Users

#### GET /users
Returns a list of users.

**Parameters:**
- `limit` (optional): Maximum number of results to return (default: 20, max: 100)
- `offset` (optional): Number of results to skip (default: 0)

**Response:**
```json
{
  "users": [
    {
      "id": "user_123",
      "name": "John Doe",
      "email": "john@example.com",
      "created_at": "2023-01-15T12:00:00Z"
    }
  ],
  "total": 50,
  "limit": 20,
  "offset": 0
}
```

#### GET /users/{id}
Returns information about a specific user.

**Response:**
```json
{
  "id": "user_123",
  "name": "John Doe",
  "email": "john@example.com",
  "created_at": "2023-01-15T12:00:00Z",
  "subscription": {
    "plan": "premium",
    "status": "active"
  }
}
```

#### POST /users
Creates a new user.

**Request Body:**
```json
{
  "name": "Jane Smith",
  "email": "jane@example.com",
  "password": "secure_password"
}
```

**Response:**
```json
{
  "id": "user_456",
  "name": "Jane Smith",
  "email": "jane@example.com",
  "created_at": "2023-06-20T15:30:00Z"
}
```

### Documents

#### GET /documents
Returns a list of documents.

**Parameters:**
- `limit` (optional): Maximum number of results to return
- `offset` (optional): Number of results to skip
- `status` (optional): Filter by status ("draft", "published", "archived")

**Response:**
```json
{
  "documents": [
    {
      "id": "doc_789",
      "title": "Sample Document",
      "status": "published",
      "created_at": "2023-05-10T09:45:00Z"
    }
  ],
  "total": 15,
  "limit": 20,
  "offset": 0
}
```

## Error Handling

API errors return appropriate HTTP status codes and include a JSON response body:

```json
{
  "error": {
    "code": "invalid_request",
    "message": "The request was invalid",
    "details": "Email address is already in use"
  }
}
```

Common error codes:
- `invalid_request`: The request was malformed
- `authentication_error`: Invalid credentials
- `permission_denied`: Lacks permissions
- `not_found`: Resource not found
- `rate_limit_exceeded`: Too many requests

## Rate Limits
Requests are limited to 60 per minute per API key. Rate limit information is included in response headers:
- `X-RateLimit-Limit`: Total requests allowed per window
- `X-RateLimit-Remaining`: Remaining requests in current window
- `X-RateLimit-Reset`: Time when the rate limit resets (Unix timestamp) 