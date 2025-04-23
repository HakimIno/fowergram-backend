package postgres

import (
	"fmt"
	"time"

	"fowergram/internal/core/domain"

	"gorm.io/gorm"
)

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) *authRepository {
	return &authRepository{db: db}
}

func (r *authRepository) CreateUser(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *authRepository) FindUserByEmail(email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) UpdateUser(user *domain.User) error {
	return r.db.Save(user).Error
}

func (r *authRepository) CreateDeviceSession(session *domain.DeviceSession) error {
	// Deactivate other sessions if this is a new device
	if err := r.db.Model(&domain.DeviceSession{}).
		Where("user_id = ? AND device_id != ?", session.UserID, session.DeviceID).
		Update("is_current", false).Error; err != nil {
		return err
	}
	return r.db.Create(session).Error
}

func (r *authRepository) GetActiveSessions(userID uint) ([]*domain.DeviceSession, error) {
	var sessions []*domain.DeviceSession
	err := r.db.Where("user_id = ? AND is_current = ?", userID, true).Find(&sessions).Error
	return sessions, err
}

func (r *authRepository) RevokeSession(userID uint, deviceID string) error {
	return r.db.Model(&domain.DeviceSession{}).
		Where("user_id = ? AND device_id = ?", userID, deviceID).
		Update("is_current", false).Error
}

func (r *authRepository) LogLogin(history *domain.LoginHistory) error {
	return r.db.Create(history).Error
}

func (r *authRepository) GetLoginHistory(userID uint) ([]*domain.LoginHistory, error) {
	var history []*domain.LoginHistory
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(10).
		Find(&history).Error
	return history, err
}

func (r *authRepository) CreateAuthCode(code *domain.AuthCode) error {
	// Invalidate existing codes for the same purpose
	if err := r.db.Model(&domain.AuthCode{}).
		Where("user_id = ? AND purpose = ? AND is_used = ?", code.UserID, code.Purpose, false).
		Update("is_used", true).Error; err != nil {
		return err
	}
	return r.db.Create(code).Error
}

func (r *authRepository) ValidateAuthCode(userID uint, code, purpose string) error {
	var authCode domain.AuthCode
	if err := r.db.Where("user_id = ? AND code = ? AND purpose = ? AND is_used = ? AND expires_at > ?",
		userID, code, purpose, false, time.Now()).First(&authCode).Error; err != nil {
		return fmt.Errorf("invalid or expired code")
	}

	// Mark code as used
	authCode.IsUsed = true
	return r.db.Save(&authCode).Error
}

func (r *authRepository) CreateAccountRecovery(recovery *domain.AccountRecovery) error {
	// Cancel existing recovery requests
	if err := r.db.Model(&domain.AccountRecovery{}).
		Where("user_id = ? AND status = ?", recovery.UserID, "pending").
		Update("status", "cancelled").Error; err != nil {
		return err
	}
	return r.db.Create(recovery).Error
}

func (r *authRepository) UpdateAccountRecovery(recovery *domain.AccountRecovery) error {
	return r.db.Save(recovery).Error
}

// FindUserByID finds a user by their ID
func (r *authRepository) FindUserByID(id uint) (*domain.User, error) {
	var user domain.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindUserByUsername finds a user by their username
func (r *authRepository) FindUserByUsername(username string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
