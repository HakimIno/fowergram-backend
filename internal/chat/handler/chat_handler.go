package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	"fowergram/internal/chat/domain"
	"fowergram/internal/chat/service"
)

type ChatHandler struct {
	chatService *service.ChatService
	wsManager   *service.WebSocketManager
	validator   *validator.Validate
}

func NewChatHandler(chatService *service.ChatService, wsManager *service.WebSocketManager) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		wsManager:   wsManager,
		validator:   validator.New(),
	}
}

func (h *ChatHandler) HandleWebSocket(c *websocket.Conn) {
	userID := c.Locals("user_id").(string)
	h.wsManager.AddConnection(userID, c)
	defer h.wsManager.RemoveConnection(userID, c)

	for {
		messageType, message, err := c.ReadMessage()
		if err != nil {
			break
		}

		if messageType == websocket.TextMessage {
			var msg domain.Message
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}

			msg.SenderID = userID
			msg.CreatedAt = time.Now()

			if err := h.chatService.HandleMessage(context.Background(), &msg); err != nil {
				continue
			}
		}
	}
}

type CreateChatRequest struct {
	Type         string   `json:"type" validate:"required,oneof=direct group broadcast"`
	Participants []string `json:"participants" validate:"required,min=1"`
}

func (h *ChatHandler) CreateChat(c *fiber.Ctx) error {
	fmt.Printf("Received CreateChat request: %s\n", string(c.Body()))

	userID := c.Locals("user_id").(string)
	if userID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized")
	}

	var req CreateChatRequest
	if err := c.BodyParser(&req); err != nil {
		fmt.Printf("Error parsing request: %v\n", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Validation failed")
	}

	if len(req.Participants) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Participants list cannot be empty")
	}

	hasCreator := false
	for _, participant := range req.Participants {
		if participant == userID {
			hasCreator = true
			break
		}
	}
	if !hasCreator {
		req.Participants = append(req.Participants, userID)
	}

	chat := &domain.Chat{
		ID:        generateID(),
		Type:      string(req.Type),
		CreatedBy: userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Members:   req.Participants,
	}

	if err := h.chatService.CreateChat(c.Context(), chat); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create chat")
	}

	return c.JSON(chat)
}

func (h *ChatHandler) GetMessages(c *fiber.Ctx) error {
	chatID := c.Params("chat_id")
	limit := c.QueryInt("limit", 50)

	messages, err := h.chatService.GetMessages(c.Context(), chatID, limit)
	if err != nil {
		return err
	}

	return c.JSON(messages)
}

func (h *ChatHandler) GetUserChats(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	chats, err := h.chatService.GetUserChats(c.Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(chats)
}

func generateID() string {
	return time.Now().Format("20060102150405.000000")
}
