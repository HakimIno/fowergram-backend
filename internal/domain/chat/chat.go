package domain

type ChatType string

const (
	DirectChat    ChatType = "direct"
	GroupChat     ChatType = "group"
	BroadcastChat ChatType = "broadcast"
)

type ChatRole string

const (
	ChatRoleOwner  ChatRole = "owner"
	ChatRoleMember ChatRole = "member"
	ChatRoleAdmin  ChatRole = "admin"
)
