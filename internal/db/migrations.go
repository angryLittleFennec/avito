package db

import (
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {

	if err := db.AutoMigrate(&Banner{}, &BannerFeatureTag{}); err != nil {
		return err
	}
	return nil

}
