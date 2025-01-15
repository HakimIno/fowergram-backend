CREATE KEYSPACE IF NOT EXISTS fowergram
WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

USE fowergram;

CREATE TABLE IF NOT EXISTS chat_messages (
    conversation_id text,
    message_id text,
    sender_id text,
    content text,
    type text,
    created_at timestamp,
    PRIMARY KEY ((conversation_id), created_at, message_id)
) WITH CLUSTERING ORDER BY (created_at DESC, message_id ASC);

CREATE TABLE IF NOT EXISTS user_status (
    user_id text PRIMARY KEY,
    status text,
    last_seen timestamp,
    updated_at timestamp
);

CREATE TABLE IF NOT EXISTS notifications (
    user_id text,
    id text,
    type text,
    content text,
    read boolean,
    created_at timestamp,
    PRIMARY KEY ((user_id), created_at)
) WITH CLUSTERING ORDER BY (created_at DESC); 