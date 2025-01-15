ALTER TABLE users 
ADD COLUMN notification_enabled BOOLEAN DEFAULT true,
ADD COLUMN theme VARCHAR(10) DEFAULT 'light',
ADD COLUMN language VARCHAR(5) DEFAULT 'en',
ADD COLUMN timezone VARCHAR(50);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username); 