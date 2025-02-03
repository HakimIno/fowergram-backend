-- Down migration
DROP TABLE IF EXISTS chat_messages;

CREATE TABLE chat_messages (
    -- โครงสร้างตารางเดิม
    conversation_id text,
    message_id text,
    sender_id text,
    content text,
    type text,
    created_at timestamp,
    PRIMARY KEY (conversation_id, created_at, message_id)
) WITH CLUSTERING ORDER BY (created_at DESC);