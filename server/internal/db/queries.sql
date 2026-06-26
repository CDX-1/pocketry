-- name: CreateUser :exec
-- :exec tells sqlc this query modifies the database but doesn't return any rows.
INSERT INTO users (
    username, 
    password_hash
) VALUES (?, ?);

-- name: GetUserByUsername :one
-- :one tells sqlc to expect exactly one row and scan it directly into a struct.
SELECT id, password_hash 
FROM users 
WHERE username = ? 
LIMIT 1;

-- name: SaveVault :exec
-- This handles the upsert logic. If a vault for the user exists, overwrite it; otherwise, create it.
INSERT INTO vaults (
    user_id, 
    encrypted_blob
) VALUES (?, ?)
ON CONFLICT(user_id) DO UPDATE SET 
    encrypted_blob = excluded.encrypted_blob, 
    updated_at = CURRENT_TIMESTAMP;

-- name: GetVaultByUserID :one
-- Fetches just the encrypted string blob for a specific user ID.
SELECT encrypted_blob 
FROM vaults 
WHERE user_id = ? 
LIMIT 1;

-- name: CreateSession :exec
-- Creates a new session for a user allowing vault access.
INSERT INTO sessions (
    user_id,
    token_hash,
    expires_at
) VALUES (?, ?, ?);

-- name: GetSessionByTokenHash :one
-- Retrives a session by its corresponding token hash.
SELECT user_id
FROM sessions
WHERE token_hash = ?
AND expires_at > CURRENT_TIMESTAMP
LIMIT 1;

-- name: DeleteSession :exec
-- Deletes a session by its corresponding token hash.
DELETE FROM sessions
WHERE token_hash = ?;