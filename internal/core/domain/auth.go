package domain

import "time"

type DeviceSession struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id"`
	DeviceID   string    `json:"device_id"`
	DeviceName string    `json:"device_name"`
	DeviceType string    `json:"device_type"`
	IPAddress  string    `json:"ip_address"`
	Location   string    `json:"location"`
	UserAgent  string    `json:"user_agent"`
	LastActive time.Time `json:"last_active"`
	IsCurrent  bool      `json:"is_current"`
	CreatedAt  time.Time `json:"created_at"`
}

type LoginHistory struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"column:user_id"`
	DeviceID  string    `json:"device_id" gorm:"column:device_id"`
	IPAddress string    `json:"ip_address" gorm:"column:ip_address"`
	Location  string    `json:"location" gorm:"column:location"`
	UserAgent string    `json:"user_agent" gorm:"column:user_agent"`
	Status    string    `json:"status" gorm:"column:status"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
}

func (LoginHistory) TableName() string {
	return "login_history"
}

type AuthCode struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`
	Code      string    `json:"code"`
	Purpose   string    `json:"purpose"`
	ExpiresAt time.Time `json:"expires_at"`
	IsUsed    bool      `json:"is_used"`
	CreatedAt time.Time `json:"created_at"`
}

type AccountRecovery struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	UserID      uint       `json:"user_id"`
	RequestType string     `json:"request_type"`
	Status      string     `json:"status"`
	InitiatedAt time.Time  `json:"initiated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	ExpiresAt   time.Time  `json:"expires_at"`
}

type RefreshToken struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
