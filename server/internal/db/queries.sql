-- CreateUser registers a new OPAQUE user after registration has been completed.
-- name: CreateUser :exec
INSERT INTO users (
	username,
	username_normalized,
	opaque_registration_record
) VALUES (?, ?, ?);

-- GetUserByUsernameNormalized retrieves a user by normalized username for OPAQUE login.
-- name: GetUserByUsernameNormalized :one
SELECT
	id,
	username,
	username_normalized,
	opaque_registration_record,
	crypto_policy_version,
	created_at,
	updated_at
FROM users
WHERE username_normalized = ?
LIMIT 1;

-- GetUserByID retrieves a user by internal user ID.
-- name: GetUserByID :one
SELECT
	id,
	username,
	username_normalized,
	crypto_policy_version,
	created_at,
	updated_at
FROM users
WHERE id = ?
LIMIT 1;

-- UsernameExists checks whether a normalized username is already registered.
-- name: UsernameExists :one
SELECT COUNT(*) > 0 AS username_exists
FROM users
WHERE username_normalized = ?;

-- CreatePendingRegistration stores temporary OPAQUE registration state.
-- name: CreatePendingRegistration :exec
INSERT INTO pending_registrations (
	id,
	username,
	username_normalized,
	server_state,
	expires_at
) VALUES (?, ?, ?, ?, ?);

-- GetPendingRegistration retrieves pending OPAQUE registration state.
-- name: GetPendingRegistration :one
SELECT
	id,
	username,
	username_normalized,
	server_state,
	expires_at,
	created_at
FROM pending_registrations
WHERE id = ?
LIMIT 1;

-- name: ConsumePendingRegistration :one
DELETE FROM pending_registrations
WHERE id = ?
RETURNING
	username,
	username_normalized,
	server_state,
	expires_at;

-- DeletePendingRegistration removes pending OPAQUE registration state after completion or cancellation.
-- name: DeletePendingRegistration :exec
DELETE FROM pending_registrations
WHERE id = ?;

-- DeleteExpiredPendingRegistrations removes expired OPAQUE registration attempts.
-- name: DeleteExpiredPendingRegistrations :exec
DELETE FROM pending_registrations
WHERE expires_at <= ?;

-- DeletePendingRegistrationsByUsernameNormalized removes older pending registration attempts for a username.
-- name: DeletePendingRegistrationsByUsernameNormalized :exec
DELETE FROM pending_registrations
WHERE username_normalized = ?;

-- CreatePendingLogin stores temporary OPAQUE login state.
-- name: CreatePendingLogin :exec
INSERT INTO pending_logins (
	id,
	user_id,
	server_state,
	expires_at
) VALUES (?, ?, ?, ?);

-- GetPendingLogin retrieves pending OPAQUE login state.
-- name: GetPendingLogin :one
SELECT
	id,
	user_id,
	server_state,
	expires_at,
	created_at
FROM pending_logins
WHERE id = ?
LIMIT 1;

-- name: ConsumePendingLogin :one
DELETE FROM pending_logins
WHERE id = ?
RETURNING
	user_id,
	server_state,
	expires_at;

-- DeletePendingLogin removes pending OPAQUE login state after completion or cancellation.
-- name: DeletePendingLogin :exec
DELETE FROM pending_logins
WHERE id = ?;

-- DeleteExpiredPendingLogins removes expired OPAQUE login attempts.
-- name: DeleteExpiredPendingLogins :exec
DELETE FROM pending_logins
WHERE expires_at <= ?;

-- DeletePendingLoginsByUserID removes older pending login attempts for a user.
-- name: DeletePendingLoginsByUserID :exec
DELETE FROM pending_logins
WHERE user_id = ?;

-- CreateVault initializes a user's encrypted vault.
-- name: CreateVault :exec
INSERT INTO vaults (
	user_id,
	encrypted_blob
) VALUES (?, ?);

-- GetVaultByUserID fetches the encrypted vault blob and revision for a user.
-- name: GetVaultByUserID :one
SELECT
	encrypted_blob,
	revision,
	created_at,
	updated_at
FROM vaults
WHERE user_id = ?
LIMIT 1;

-- UpdateVaultIfRevisionMatches updates a vault only if the expected revision matches.
-- name: UpdateVaultIfRevisionMatches :execrows
UPDATE vaults
SET
	encrypted_blob = ?,
	revision = revision + 1,
	updated_at = CURRENT_TIMESTAMP
WHERE user_id = ?
AND revision = ?;

-- ListUsers lists registered users within the specified limit
-- name: ListUsers :many
SELECT
	id,
	username,
	crypto_policy_version,
	created_at,
	updated_at
FROM users
ORDER BY id ASC
LIMIT ?;

-- DeleteUser deletes a user and all associated data.
-- name: DeleteUser :exec
DELETE FROM users
WHERE id = ?;