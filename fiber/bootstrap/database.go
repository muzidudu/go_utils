package bootstrap

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// initDatabase 初始化数据库
func initDatabase(cfg *Config) *gorm.DB {
	sslmode := cfg.Database.SSLMode
	if sslmode == "" {
		sslmode = "disable"
	}
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.Database, sslmode,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("warn: database connect failed: %v (app will run without DB)", err)
		return nil
	}
	return db
}

// AutoMigrate 自动迁移模型（传入需迁移的模型指针）
func AutoMigrate(db *gorm.DB, models ...any) error {
	if db == nil {
		return nil
	}
	return db.AutoMigrate(models...)
}
