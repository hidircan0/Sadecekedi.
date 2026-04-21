package domain

import "time"

type Photo struct {
	ID         uint   `gorm:"primaryKey"`
	Filename   string `gorm:"size:255;not null;uniqueIndex"`
	UploadedAt time.Time
	IsCat      bool `gorm:"not null;default:true"`
}

func (Photo) TableName() string {
	return "photos"
}
