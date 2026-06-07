ALTER TABLE users ALTER COLUMN verification_token TYPE TEXT;
ADD COLUMN username TEXT UNIQUE,
ADD COLUMN is_email_verified BOOLEAN DEFAULT FALSE,
ADD COLUMN verification_token UUID;
