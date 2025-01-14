package routes

import (
	"fowergram/internal/chat/handler"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func SetupChatRoutes(api fiber.Router, chatHandler *handler.ChatHandler) {
	chat := api.Group("/chat")
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
}
