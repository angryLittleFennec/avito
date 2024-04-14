package db

import (
	"encoding/json"
	"time"
)

type Banner struct {
	ID        uint            `gorm:"primaryKey"`
	Content   json.RawMessage `gorm:"type:json"`
	CreatedAt time.Time       `gorm:"autoCreateTime"`
	UpdatedAt time.Time       `gorm:"autoUpdateTime"`
	IsActive  bool
}

type BannerFeatureTag struct {
	ID        uint `gorm:"primaryKey"`
	BannerID  uint
	FeatureID int `gorm:"index:idx_feature_tag,unique"`
	TagID     int `gorm:"index:idx_feature_tag,unique"`
}

func (BannerFeatureTag) TableName() string {
	return "banner_feature_tags"
}
