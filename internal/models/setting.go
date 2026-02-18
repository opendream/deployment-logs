package models

import "gorm.io/gorm"

type Setting struct {
	Key   string `gorm:"primaryKey" json:"key"`
	Value string `json:"value"`
}

func GetSetting(db *gorm.DB, key string) (string, error) {
	var s Setting
	if err := db.Where("key = ?", key).First(&s).Error; err != nil {
		return "", err
	}
	return s.Value, nil
}

func SetSetting(db *gorm.DB, key, value string) error {
	s := Setting{Key: key, Value: value}
	return db.Save(&s).Error
}
