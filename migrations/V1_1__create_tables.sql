CREATE TABLE users (
    id UUID NOT NULL,
    email TEXT NOT NULL,
    UNIQUE(id),
    UNIQUE (email)
);

CREATE TABLE sessions (
    id SERIAL PRIMARY KEY,
    refresh_hash TEXT NOT NULL,
    ip TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (refresh_hash)
);

CREATE FUNCTION update_session_function()
RETURNS TRIGGER
AS $$
    BEGIN
        NEW.created_at = CURRENT_TIMESTAMP;
        RETURN NEW;
    END;

$$ LANGUAGE PLPGSQL;

CREATE TRIGGER update_session_trigger
BEFORE UPDATE
ON sessions
FOR EACH ROW
EXECUTE FUNCTION update_session_function();

CREATE TABLE user_session (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    session_id INT NOT NULL,
    UNIQUE (session_id),
    FOREIGN KEY (user_id) REFERENCES users (id)
                          ON DELETE CASCADE,
    FOREIGN KEY (session_id) REFERENCES sessions (id)
                          ON DELETE CASCADE
);

CREATE VIEW session_view (refresh_hash, user_id, user_ip, user_email, created_at) AS
SELECT s.refresh_hash, us.user_id, s.ip, u.email, s.created_at
FROM sessions s
JOIN user_session us ON us.session_id = s.id
JOIN users u ON u.id = us.user_id;