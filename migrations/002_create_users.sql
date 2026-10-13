-- Create users table for JWT authentication
CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    username      TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed default admin user (password: admin123, bcrypt hash)
-- $2a$10$7Z2vA9r5x1b1h3jQkH3F7.pYqTjE1H4n/Qe3F6Y0mJ3Y2r9Q0Zk6m (bcrypt for admin123)
INSERT INTO users (username, password_hash)
VALUES ('admin', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi')
ON CONFLICT (username) DO NOTHING;
