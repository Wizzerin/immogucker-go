CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT DEFAULT 'user' NOT NULL, -- 'user' или 'admin'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
