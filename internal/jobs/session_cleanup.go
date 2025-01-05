package jobs

import (
	"fowergram/internal/core/domain"
	"time"

	"gorm.io/gorm"
)

func StartSessionCleanup(db *gorm.DB) {
	ticker := time.NewTicker(24 * time.Hour)
	for range ticker.C {
		CleanupInactiveSessions(db)
	}
}

func CleanupInactiveSessions(db *gorm.DB) {
	// ลบ sessions ที่ไม่ได้ใช้งานเกิน 30 วัน
	db.Where("last_active < ? AND is_current = ?",
		time.Now().AddDate(0, 0, -30), false).
		Delete(&domain.DeviceSession{})
}
