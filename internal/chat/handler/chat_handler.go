package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	"fowergram/internal/domain/chat"
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
	userIDRaw := c.Locals("user_id")
	if userIDRaw == nil {
		return
	}
	userID := fmt.Sprintf("%d", userIDRaw)
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
	Type         string        `json:"type" validate:"required,oneof=direct group broadcast"`
	Name         string        `json:"name" validate:"required_if=Type group,required_if=Type broadcast"`
	Participants []interface{} `json:"participants" validate:"required,min=1"`
	IsPrivate    bool          `json:"is_private"`
}

func (h *ChatHandler) CreateChat(c *fiber.Ctx) error {
	userIDRaw := c.Locals("user_id")
	if userIDRaw == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized: user_id not found in context")
	}
	userID := fmt.Sprintf("%d", userIDRaw)

	var req CreateChatRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Validation failed: %v", err))
	}

	if len(req.Participants) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Participants list cannot be empty")
	}

	// Convert all participant IDs to string
	participants := make([]string, 0, len(req.Participants))
	for _, p := range req.Participants {
		if p != nil {
			participants = append(participants, fmt.Sprintf("%v", p))
		}
	}

	if len(participants) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "No valid participants")
	}

	switch req.Type {
	case "direct":
		if len(participants) > 1 {
			return fiber.NewError(fiber.StatusBadRequest, "Direct chat can only have one participant")
		}

		// Get all user chats
		existingChats, err := h.chatService.GetUserChats(c.Context(), userID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to check existing chats: %v", err))
		}

		otherUserID := participants[0]
		for _, chat := range existingChats {
			if chat.Type == "direct" && len(chat.Members) == 2 {
				hasCurrentUser := false
				hasOtherUser := false
				for _, member := range chat.Members {
					if member == userID {
						hasCurrentUser = true
					}
					if member == otherUserID {
						hasOtherUser = true
					}
				}
				if hasCurrentUser && hasOtherUser {
					return c.JSON(chat)
				}
			}
		}

	case "group":
		// Validate group requirements
		if len(participants) > 200 { // Example limit
			return fiber.NewError(fiber.StatusBadRequest, "Group cannot have more than 200 participants")
		}
		if req.Name == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Group name is required")
		}
		if len(req.Name) > 255 {
			return fiber.NewError(fiber.StatusBadRequest, "Group name is too long")
		}

	case "broadcast":
		// Validate broadcast requirements
		if req.Name == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Channel name is required")
		}
		if len(req.Name) > 255 {
			return fiber.NewError(fiber.StatusBadRequest, "Channel name is too long")
		}

		// Check for unique channel name
		existingChats, err := h.chatService.GetUserChats(c.Context(), userID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to check existing channels: %v", err))
		}
		for _, chat := range existingChats {
			if chat.Type == "broadcast" && chat.Name == req.Name {
				return fiber.NewError(fiber.StatusConflict, "Channel name already exists")
			}
		}
	}

	// Ensure creator is in the participants list
	hasCreator := false
	for _, participant := range participants {
		if participant == userID {
			hasCreator = true
			break
		}
	}
	if !hasCreator {
		participants = append(participants, userID)
	}

	chat := &domain.Chat{
		ID:        generateID(),
		Name:      req.Name,
		Type:      string(req.Type),
		CreatedBy: userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Members:   participants,
		IsPrivate: req.IsPrivate,
	}

	if err := h.chatService.CreateChat(c.Context(), chat); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to create chat: %v", err))
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
	userIDRaw := c.Locals("user_id")
	if userIDRaw == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized")
	}
	userID := fmt.Sprintf("%d", userIDRaw)

	chats, err := h.chatService.GetUserChats(c.Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(chats)
}

func generateID() string {
	return time.Now().Format("20060102150405.000000")
}

type CreateInviteLinkRequest struct {
	MaxUses   int    `json:"max_uses" validate:"min=0"`
	ExpiresIn string `json:"expires_in" validate:"required"`
}

func (h *ChatHandler) CreateInviteLink(c *fiber.Ctx) error {
	userIDRaw := c.Locals("user_id")
	if userIDRaw == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized")
	}
	userID := fmt.Sprintf("%d", userIDRaw)

	chatID := c.Params("chat_id")
	if chatID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Chat ID is required")
	}

	var req CreateInviteLinkRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
	}

	if err := h.validator.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Validation failed: %v", err))
	}

	duration, err := time.ParseDuration(req.ExpiresIn)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Invalid duration format. Use format like '24h', '7d', '30m': %v", err))
	}

	if duration < 5*time.Minute {
		return fiber.NewError(fiber.StatusBadRequest, "Expiration time must be at least 5 minutes")
	}

	link, err := h.chatService.CreateInviteLink(c.Context(), chatID, userID, req.MaxUses, duration)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to create invite link: %v", err))
	}

	return c.JSON(link)
}

func (h *ChatHandler) GetChatInviteLinks(c *fiber.Ctx) error {
	userIDRaw := c.Locals("user_id")
	if userIDRaw == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized")
	}

	chatID := c.Params("chat_id")
	if chatID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Chat ID is required")
	}

	links, err := h.chatService.GetChatInviteLinks(c.Context(), chatID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to get invite links: %v", err))
	}

	return c.JSON(links)
}

func (h *ChatHandler) JoinChatViaInvite(c *fiber.Ctx) error {
	userIDRaw := c.Locals("user_id")
	if userIDRaw == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized")
	}
	userID := fmt.Sprintf("%d", userIDRaw)

	code := c.Params("code")
	if code == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Invite code is required")
	}

	chat, err := h.chatService.JoinChatViaInvite(c.Context(), code, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Failed to join chat: %v", err))
	}

	return c.JSON(chat)
}

func (h *ChatHandler) DeleteInviteLink(c *fiber.Ctx) error {
	userIDRaw := c.Locals("user_id")
	if userIDRaw == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized")
	}

	chatID := c.Params("chat_id")
	if chatID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Chat ID is required")
	}

	code := c.Params("code")
	if code == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Invite code is required")
	}

	if err := h.chatService.DeleteInviteLink(c.Context(), chatID, code); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to delete invite link: %v", err))
	}

	return c.SendStatus(fiber.StatusNoContent)
}
