CREATE KEYSPACE IF NOT EXISTS fowergram 
WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

CREATE TABLE IF NOT EXISTS fowergram.chats (
    id text,
    name text,
    type text,
    created_by text,
    created_at timestamp,
    updated_at timestamp,
    members list<text>,
    is_private boolean,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS fowergram.chat_messages (
    chat_id text,
    message_id timeuuid,
    sender_id text,
    content text,
    created_at timestamp,
    PRIMARY KEY (chat_id, message_id)
) WITH CLUSTERING ORDER BY (message_id DESC);

CREATE TABLE IF NOT EXISTS fowergram.user_chats (
    user_id text,
    chat_id text,
    last_read_at timestamp,
    PRIMARY KEY (user_id, chat_id)
);

CREATE TABLE IF NOT EXISTS fowergram.chat_invite_links (
    chat_id text,
    code text,
    created_by text,
    created_at timestamp,
    expires_at timestamp,
    max_uses int,
    uses int,
    PRIMARY KEY (chat_id, code)
);

CREATE INDEX IF NOT EXISTS chat_invite_links_by_code ON fowergram.chat_invite_links (code); 