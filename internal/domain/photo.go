package domain

import "time"

type Photo struct {
	ID         uint   `gorm:"primaryKey"`
	Filename   string `gorm:"size:255;not null;uniqueIndex"`
	UploadedAt time.Time
	IsCat      bool `gorm:"not null;default:true"`
	UploadJobID *string `gorm:"size:36;uniqueIndex" json:"-"`
}

type UploadJob struct {
	ID         string `gorm:"primaryKey;size:36" json:"id"`
	Filename   string `gorm:"size:255;not null" json:"filename"`
	Status     string `gorm:"size:20;not null;index" json:"status"`
	Error      string `gorm:"size:500" json:"error,omitempty"`
	PhotoID    *uint  `json:"photo_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (UploadJob) TableName() string { return "upload_jobs" }

type OutboxEvent struct {
	ID          uint       `gorm:"primaryKey"`
	JobID       string     `gorm:"size:36;not null;uniqueIndex"`
	CreatedAt   time.Time
	PublishedAt *time.Time
}

func (Photo) TableName() string {
	return "photos"
}
