# API Reference

## Authentication
All API requests require authentication using Bearer tokens.

## Users API
### GET /users
Returns a list of all users.

### POST /users
Creates a new user.

### GET /users/{id}
Returns details for a specific user.

### PUT /users/{id}
Updates an existing user.

### DELETE /users/{id}
Deletes a user.

## Error Codes
- 200: Success
- 400: Bad Request
- 401: Unauthorized
- 404: Not Found
- 500: Internal Server Error
