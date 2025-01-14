package handler

import (
	"context"
	"fowergram/internal/chat/domain"
	"fowergram/internal/chat/service"
	"fowergram/internal/config"
	"fowergram/internal/security"

	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type ChatHandler struct {
	chatService *service.ChatService
	wsManager   *service.WebSocketManager
}

func NewChatHandler(chatService *service.ChatService, wsManager *service.WebSocketManager) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		wsManager:   wsManager,
	}
}

func (h *ChatHandler) HandleWebSocket(c *websocket.Conn) {
	// Wait for authentication message
	var authMsg struct {
		Type  string `json:"type"`
		Token string `json:"token"`
	}

	if err := c.ReadJSON(&authMsg); err != nil {
		log.Printf("Error reading auth message: %v", err)
		return
	}

	if authMsg.Type != "auth" || authMsg.Token == "" {
		log.Printf("Invalid auth message")
		return
	}

	// Validate token and get user ID
	claims, err := security.ValidateJWT(authMsg.Token, config.GetJWTSecret())
	if err != nil {
		log.Printf("Invalid token: %v", err)
		return
	}

	userID := claims.UserID
	if userID == "" {
		log.Printf("No user ID in token")
		return
	}

	h.wsManager.AddClient(userID, c.Conn)
	defer h.wsManager.RemoveClient(userID)

	for {
		var msg struct {
			Type string         `json:"type"`
			Data domain.Message `json:"data,omitempty"`
		}

		if err := c.ReadJSON(&msg); err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		switch msg.Type {
		case "ping":
			if err := c.WriteJSON(map[string]string{"type": "pong"}); err != nil {
				log.Printf("Error sending pong: %v", err)
				return
			}

		case "chat":
			msg.Data.SenderID = userID
			if err := h.chatService.SendMessage(context.Background(), &msg.Data); err != nil {
				if err := c.WriteJSON(fiber.Map{
					"error": "Failed to send message",
				}); err != nil {
					log.Printf("Error sending error message: %v", err)
					break
				}
			}

		default:
			log.Printf("Unknown message type: %s", msg.Type)
		}
	}
}

func (h *ChatHandler) CreateChat(c *fiber.Ctx) error {
	var chat domain.Chat
	if err := c.BodyParser(&chat); err != nil {
		log.Printf("Error parsing chat body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate members
	if len(chat.Members) < 2 {
		log.Printf("Invalid number of members: %d", len(chat.Members))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Chat must have at least 2 members",
		})
	}

	for i, member := range chat.Members {
		if member == "" {
			log.Printf("Empty member ID at index %d", i)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Member IDs cannot be empty",
			})
		}
	}

	log.Printf("Creating chat with members: %v", chat.Members)

	ctx := c.Context()
	if err := h.chatService.CreateChat(ctx, &chat); err != nil {
		log.Printf("Error creating chat: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create chat",
		})
	}

	log.Printf("Chat created successfully with ID: %s", chat.ID)
	return c.JSON(chat)
}

func (h *ChatHandler) GetMessages(c *fiber.Ctx) error {
	chatID := c.Params("id")
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	ctx := c.Context()
	messages, err := h.chatService.GetMessages(ctx, chatID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get messages",
		})
	}

	return c.JSON(messages)
}

func (h *ChatHandler) GetUserChats(c *fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	ctx := c.Context()
	chats, err := h.chatService.GetUserChats(ctx, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user chats",
		})
	}

	return c.JSON(chats)
}
