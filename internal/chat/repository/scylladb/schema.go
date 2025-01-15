package scylladb

import (
	"fmt"
	"strings"

	"github.com/gocql/gocql"
)

const (
	createKeyspaceTemplate = `
		CREATE KEYSPACE IF NOT EXISTS %s
		WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}
	`

	createTablesQuery = `
		CREATE TABLE IF NOT EXISTS chats (
			id text PRIMARY KEY,
			name text,
			type text,
			created_by text,
			created_at timestamp,
			updated_at timestamp,
			is_private boolean,
			members list<text>
		);

		CREATE TABLE IF NOT EXISTS chat_members (
			chat_id text,
			user_id text,
			role text,
			joined_at timestamp,
			updated_at timestamp,
			PRIMARY KEY ((chat_id), user_id)
		);

		CREATE INDEX IF NOT EXISTS chat_members_user_id_idx ON chat_members (user_id);

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
	`
)

func InitializeSchema(hosts []string, keyspace string) error {
	// Connect to ScyllaDB without keyspace
	cluster := gocql.NewCluster(hosts...)
	cluster.Consistency = gocql.Quorum
	cluster.ProtoVersion = 4

	session, err := cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Create keyspace
	createKeyspace := fmt.Sprintf(createKeyspaceTemplate, keyspace)
	if err := session.Query(createKeyspace).Exec(); err != nil {
		return fmt.Errorf("failed to create keyspace: %w", err)
	}

	// Connect to the created keyspace
	cluster.Keyspace = keyspace
	session, err = cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("failed to connect to keyspace: %w", err)
	}
	defer session.Close()

	// Execute each CREATE TABLE statement separately
	for _, query := range strings.Split(createTablesQuery, ";") {
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}

		if err := session.Query(query).Exec(); err != nil {
			return fmt.Errorf("failed to execute query %q: %w", query, err)
		}
	}

	return nil
}
