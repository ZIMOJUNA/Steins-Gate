-- steins-gate:drop:start
DROP TABLE IF EXISTS player_data;
DROP TABLE IF EXISTS user_infos;
DROP TABLE IF EXISTS user_accounts;
DROP TABLE IF EXISTS player_accounts;
-- steins-gate:drop:end

-- steins-gate:create:start
CREATE TABLE IF NOT EXISTS user_accounts (
	id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
	email VARCHAR(255) NOT NULL,
	username VARCHAR(64) NULL,
	nickname VARCHAR(64) NOT NULL DEFAULT '',
	password_hash VARCHAR(255) NOT NULL,
	status TINYINT NOT NULL DEFAULT 1,
	login_count BIGINT UNSIGNED NOT NULL DEFAULT 0,
	last_login_ip VARCHAR(64) NOT NULL DEFAULT '',
	last_login_at DATETIME NULL,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (id),
	UNIQUE KEY uk_user_accounts_email (email),
	UNIQUE KEY uk_user_accounts_username (username),
	KEY idx_user_accounts_status (status),
	KEY idx_user_accounts_last_login_at (last_login_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS player_data (
	id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
	account_id BIGINT UNSIGNED NOT NULL,
	game_key VARCHAR(64) NOT NULL DEFAULT 'default',
	slot_key VARCHAR(64) NOT NULL DEFAULT 'default',
	data JSON NOT NULL,
	version BIGINT UNSIGNED NOT NULL DEFAULT 1,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (id),
	UNIQUE KEY uk_player_data_account_game_slot (account_id, game_key, slot_key),
	KEY idx_player_data_account_id (account_id),
	CONSTRAINT fk_player_data_account
		FOREIGN KEY (account_id) REFERENCES user_accounts(id)
		ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
-- steins-gate:create:end
