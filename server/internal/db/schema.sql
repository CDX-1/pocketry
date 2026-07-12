CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT NOT NULL,
	username_normalized TEXT NOT NULL UNIQUE,
	opaque_registration_record BLOB NOT NULL,
	crypto_policy_version INTEGER NOT NULL DEFAULT 1,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS pending_registrations (
	id TEXT PRIMARY KEY,
	username TEXT NOT NULL,
	username_normalized TEXT NOT NULL,
	server_state BLOB NOT NULL,
	expires_at DATETIME NOT NULL,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS pending_logins (
	id TEXT PRIMARY KEY,
	user_id INTEGER NOT NULL,
	server_state BLOB NOT NULL,
	expires_at DATETIME NOT NULL,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_pending_registrations_expires_at
	ON pending_registrations(expires_at);

CREATE INDEX IF NOT EXISTS idx_pending_registrations_username_normalized
	ON pending_registrations(username_normalized);

CREATE INDEX IF NOT EXISTS idx_pending_logins_expires_at
	ON pending_logins(expires_at);


-- improve later
CREATE TABLE IF NOT EXISTS vaults (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL UNIQUE,
	encrypted_blob BLOB NOT NULL,
	revision INTEGER NOT NULL DEFAULT 1,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);