CREATE TABLE channels (
    id SERIAL PRIMARY KEY,
    telegram_id BIGINT NOT NULL,
    username VARCHAR(255),
    title VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    timezone VARCHAR(50) DEFAULT 'Europe/Moscow',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    content TEXT,
    media_type VARCHAR(50) DEFAULT 'text',
    media_path VARCHAR(500),
    buttons JSONB,
    schedule_time TIMESTAMP,
    status VARCHAR(50) DEFAULT 'draft',
    created_by VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    sent_at TIMESTAMP
);

CREATE TABLE post_channels (
    id SERIAL PRIMARY KEY,
    post_id INTEGER REFERENCES posts(id) ON DELETE CASCADE,
    channel_id INTEGER REFERENCES channels(id) ON DELETE CASCADE,
    message_id INTEGER,
    status VARCHAR(50) DEFAULT 'sent',
    error TEXT,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_posts_status ON posts(status);
CREATE INDEX idx_posts_schedule_time ON posts(schedule_time);
CREATE INDEX idx_post_channels_sent_at ON post_channels(sent_at);