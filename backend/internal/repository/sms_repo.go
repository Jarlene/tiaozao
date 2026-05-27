package repository

import (
	"time"

	"github.com/jarlene/tiaozao/internal/model"
	"gorm.io/gorm"
)

type SmsRepository struct {
	db *gorm.DB
}

func NewSmsRepository(db *gorm.DB) *SmsRepository {
	return &SmsRepository{db: db}
}

func (r *SmsRepository) Save(sms *model.SmsCode) error {
	return r.db.Create(sms).Error
}

// FindLatestValid finds the latest unused, non-expired code for a phone
func (r *SmsRepository) FindLatestValid(phone string) (*model.SmsCode, error) {
	var sms model.SmsCode
	err := r.db.Where("phone = ? AND used = false AND expire_at > ?", phone, time.Now()).
		Order("created_at DESC").
		First(&sms).Error
	if err != nil {
		return nil, err
	}
	return &sms, nil
}

// MarkUsed marks a verification code as used
func (r *SmsRepository) MarkUsed(id uint) error {
	return r.db.Model(&model.SmsCode{}).Where("id = ?", id).Update("used", true).Error
}
