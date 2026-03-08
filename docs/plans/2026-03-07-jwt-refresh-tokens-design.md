# JWT Access & Refresh Token Design

**Date:** 2026-03-07
**Branch:** jwt-refresh
**Status:** Approved

## Overview

Implement a two-token authentication system with short-lived access tokens and longer-lived refresh tokens stored in the database. This improves security while maintaining flexibility for token revocation and session management.

## Requirements

- **Access tokens:** 15 minutes (configurable via `.env`)
- **Refresh tokens:** 7 days (configurable via `.env`)
- **Token storage:** Refresh tokens stored in database with hash
- **Revocation:** Mark tokens as revoked rather than delete
- **Token rotation:** No rotation; same refresh token until expiry
- **Logout support:** Revoke refresh tokens on logout

## Architecture

### Database Schema

**RefreshToken Model:**
- `ID` - Primary key
- `UserID` - Foreign key to users
- `Token` - Hashed refresh token
- `ExpiresAt` - Token expiration timestamp
- `RevokedAt` - Revocation timestamp (NULL if active)
- `CreatedAt` - Creation timestamp for audit trail

### Auth Functions

**In `internal/auth/token.go`:**
- `GenerateRefreshToken(userID int, secret string) (string, error)` - Create random token
- `HashToken(token string) string` - Hash token before storing
- `ValidateRefreshToken(tokenString string, secret string, claims *Claims) error` - Validate against DB

**Existing (unchanged):**
- `GenerateToken()` - Now generates access token only
- `ValidateToken()` - Validates access token (stateless)
- Auth middleware - Validates access token on protected routes

### API Endpoints

**Modified:**
- `POST /auth/login` - Returns `{accessToken, refreshToken}`

**New:**
- `POST /auth/refresh` - Input: `{refreshToken}`, Returns: `{accessToken}`
- `POST /auth/logout` - Protected route, marks refresh token revoked

### Configuration

**New `.env` variables:**
```
JWT_ACCESS_TOKEN_EXPIRY=15m
JWT_REFRESH_TOKEN_EXPIRY=168h
```

**Updated `config.go`:**
- Parse durations from environment
- Store as `time.Duration` in Config struct
- Pass to token generation functions

## Data Flow

### Login
1. Client: POST `/auth/login` with email/password
2. Server: Authenticate user, generate both tokens
3. Server: Store refresh token hash + metadata in DB
4. Response: `{accessToken, refreshToken}`

### Active Use
1. Client: Use accessToken for API calls (15 min window)
2. Server: Validate accessToken (signature only, no DB query)

### Token Refresh
1. Client: accessToken expires, POST `/auth/refresh` with refreshToken
2. Server: Query DB, verify token exists and not revoked/expired
3. Response: New `{accessToken}`
4. Client: Resume with new token

### Logout
1. Client: POST `/auth/logout` with accessToken in header
2. Server: Extract userID from context, mark refreshToken as revoked
3. Response: Success
4. Client: Cannot refresh anymore; must re-login

## Security Considerations

- Refresh tokens stored as hashed values (plaintext never in DB)
- Refresh tokens validated against DB for revocation
- AccessToken validation remains fast (signature only)
- RevokedAt field prevents use of compromised tokens
- Audit trail via CreatedAt and RevokedAt timestamps

## Files to Modify/Create

**Create:**
- `internal/store/refresh_token_migration.go` - DB migration
- `internal/store/refresh_token_sqlite.go` - Repository implementation

**Modify:**
- `internal/auth/token.go` - Add refresh token functions
- `internal/handler/auth.go` - Update login, add refresh/logout handlers
- `cmd/api/config.go` - Add token expiry configuration
- `cmd/api/main.go` - Register new endpoints
- `.env` (example) - Add new variables

## Testing Strategy

- Unit tests for token generation/validation
- Integration tests for login/refresh/logout flow
- Edge cases: expired tokens, revoked tokens, invalid refresh tokens
