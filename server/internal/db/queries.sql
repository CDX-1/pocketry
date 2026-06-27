-- CreateUser registers a new user in the system with a unique username and a hashed password.
-- name: CreateUser :exec
INSERT INTO users (
    username, 
    password_hash
) VALUES (?, ?);

-- GetUserByUsername retrieves a user's ID, username, and password hash by their username.
-- name: GetUserByUsername :one
SELECT id, username, password_hash
FROM users
WHERE username = ?
LIMIT 1;

-- CreateSession creates a new active session for a user, mapping their user ID to a hashed session token with an expiration timestamp.
-- name: CreateSession :exec
INSERT INTO sessions (
    user_id,
    token_hash,
    expires_at
) VALUES (?, ?, ?);

-- GetSessionByTokenHash retrieves the user ID associated with a session token, provided the session has not expired yet.
-- name: GetSessionByTokenHash :one
SELECT user_id
FROM sessions
WHERE token_hash = ?
AND expires_at > CURRENT_TIMESTAMP
LIMIT 1;

-- DeleteSessionByTokenHash permanently removes a session by its token hash. Typically invoked during user logout.
-- name: DeleteSessionByTokenHash :exec
DELETE FROM sessions
WHERE token_hash = ?;

-- CreateVault initializes a new encrypted storage vault for a specific user.
-- name: CreateVault :exec
INSERT INTO vaults (
    user_id,
    encrypted_blob
) VALUES (?, ?);

-- UpdateVaultIfRevisionMatches implements optimistic concurrency control to update a user's vault.
-- It increments the vault revision and updates the blob only if the provided revision matches the current state in the database.
-- Returns the number of affected rows (0 means a conflict occurred and the update failed).
-- name: UpdateVaultIfRevisionMatches :execrows
UPDATE vaults
SET 
    encrypted_blob = ?,
    revision = revision + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = ?
AND revision = ?;

-- GetVaultByUserID fetches the encrypted data blob, current revision number, and last update timestamp for a user's vault.
-- name: GetVaultByUserID :one
SELECT encrypted_blob, revision, updated_at
FROM vaults
WHERE user_id = ?
LIMIT 1;