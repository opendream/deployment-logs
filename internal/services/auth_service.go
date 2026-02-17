package services

import (
	"log"

	"deployment-logs/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func SeedAdminUser(db *gorm.DB, password string) {
	hash, err := HashPassword(password)
	if err != nil {
		log.Fatal("Failed to hash admin password:", err)
	}

	var user models.User
	result := db.Where("username = ?", "admin").First(&user)
	if result.Error == gorm.ErrRecordNotFound {
		user = models.User{
			Username:     "admin",
			PasswordHash: hash,
			Role:         "admin",
		}
		db.Create(&user)
		log.Println("Admin user seeded")
	} else if result.Error == nil {
		db.Model(&user).Update("password_hash", hash)
		log.Println("Admin user password updated")
	}
}
