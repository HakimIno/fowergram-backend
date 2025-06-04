package domain

import (
	"sync"
	"time"
)

type DeviceSession struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id"`
	DeviceID   string    `json:"device_id"`
	DeviceType string    `json:"device_type"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	Location   string    `json:"location"`
	LastActive time.Time `json:"last_active"`
	mu         sync.RWMutex
}

func (d *DeviceSession) GetLocation() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Location
}

func (d *DeviceSession) SetLocation(location string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Location = location
}

type LoginHistory struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`
	DeviceID  string    `json:"device_id"`
	IPAddress string    `json:"ip_address"`
	Location  string    `json:"location"`
	UserAgent string    `json:"user_agent"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type AuthCode struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`
	Code      string    `json:"code"`
	Purpose   string    `json:"purpose"`
	IsUsed    bool      `json:"is_used"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type AccountRecovery struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id"`
	RequestType string    `json:"request_type"`
	Status      string    `json:"status"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// SwitchAccountRequest is used when switching between accounts
type SwitchAccountRequest struct {
	SwitchType  string `json:"switch_type" validate:"required,oneof=token password"` // Type of switch: token or password
	Identifier  string `json:"identifier" validate:"required"`                       // Email or username of target account
	Password    string `json:"password,omitempty"`                                   // Password (for password type)
	StoredToken string `json:"stored_token,omitempty"`                               // Stored token (for token type)
}
