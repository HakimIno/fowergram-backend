package routes

import (
	"fowergram/internal/chat/handler"
	"fowergram/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// SetupChatRoutes sets up all chat-related routes with authentication middleware
func SetupChatRoutes(api fiber.Router, chatHandler *handler.ChatHandler, jwtSecret string) {
	chat := api.Group("/chat")

	// Apply JWT middleware to all chat routes
	chat.Use(middleware.JWTMiddleware(jwtSecret))

	// WebSocket endpoint for real-time chat
	chat.Get("/ws", websocket.New(chatHandler.HandleWebSocket))

	// Chat management endpoints
	chat.Post("/", chatHandler.CreateChat)
	chat.Get("/", chatHandler.GetUserChats)
	chat.Get("/:chat_id/messages", chatHandler.GetMessages)
	chat.Get("/:chat_id", chatHandler.GetChat)

	// Chat member management
	chat.Post("/:chat_id/members", chatHandler.AddChatMember)
	chat.Delete("/:chat_id/members/:user_id", chatHandler.RemoveChatMember)
	chat.Put("/:chat_id/members/:user_id/role", chatHandler.UpdateChatMemberRole)

	// Invite link management
	chat.Post("/:chat_id/invite-links", chatHandler.CreateInviteLink)
	chat.Get("/:chat_id/invite-links", chatHandler.GetChatInviteLinks)
	chat.Post("/join/:code", chatHandler.JoinChatViaInvite)
	chat.Delete("/:chat_id/invite-links/:code", chatHandler.DeleteInviteLink)
}
