CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,           -- Сессионный ID (UUID)
    user_id INTEGER NOT NULL,      -- Ссылка на таблицу users
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
