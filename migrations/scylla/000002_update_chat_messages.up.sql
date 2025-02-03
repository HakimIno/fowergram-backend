-- Up migration
DROP TABLE IF EXISTS chat_messages;

CREATE TABLE chat_messages (
    conversation_id text,
    date text,
    created_at timestamp,
    message_id text,
    sender_id text,
    content text,
    type text,
    PRIMARY KEY ((conversation_id, date), created_at, message_id)
) WITH CLUSTERING ORDER BY (created_at DESC, message_id ASC);