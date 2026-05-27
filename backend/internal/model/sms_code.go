package model

import "time"

// SmsCode represents a phone verification code
type SmsCode struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Phone     string    `gorm:"index;size:20;not null" json:"phone"`
	Code      string    `gorm:"size:6;not null" json:"-"`
	ExpireAt  time.Time `gorm:"not null" json:"-"`
	Used      bool      `gorm:"default:false" json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName specifies the table name for SmsCode
func (SmsCode) TableName() string {
	return "sms_codes"
}

// SendCodeRequest represents a request to send verification code
type SendCodeRequest struct {
	Phone string `json:"phone" binding:"required,len=11"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Phone string `json:"phone" binding:"required,len=11"`
	Code  string `json:"code" binding:"required,len=6"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token string              `json:"token"`
	User  UserProfileResponse `json:"user"`
}
