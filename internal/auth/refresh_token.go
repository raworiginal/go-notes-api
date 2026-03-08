package auth

import "time"

type RefreshToken struct {
	ID        int    `gorm:"primaryKey"`
	UserID    int    `gorm:"index"`
	Token     string `gorm:"uniqueIndex;index"`
	ExpiresAt time.Time
	RevokedAt *time.Time `gorm:"index"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
}

func (rt *RefreshToken) isValid() bool {
	if rt.RevokedAt != nil {
		return false
	}
	return time.Now().Before(rt.ExpiresAt)
}

func (rt *RefreshToken) Revoke() {
	now := time.Now()
	rt.RevokedAt = &now
}
