package routes

import (
	"fowergram/internal/chat/handler"
	"fowergram/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func SetupChatRoutes(api fiber.Router, chatHandler *handler.ChatHandler, jwtSecret string) {
	chat := api.Group("/chat")

	// Add auth middleware to all chat routes
	chat.Use(middleware.ValidateAuth(jwtSecret))

	chat.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	chat.Get("/ws", websocket.New(chatHandler.HandleWebSocket))
	chat.Post("/", chatHandler.CreateChat)
	chat.Get("/:id/messages", chatHandler.GetMessages)
	chat.Get("/user/:id", chatHandler.GetUserChats)

	// Invite link routes
	chat.Post("/:chat_id/invite", chatHandler.CreateInviteLink)
	chat.Get("/:chat_id/invite", chatHandler.GetChatInviteLinks)
	chat.Delete("/:chat_id/invite/:code", chatHandler.DeleteInviteLink)
	chat.Post("/join/:code", chatHandler.JoinChatViaInvite)
}
