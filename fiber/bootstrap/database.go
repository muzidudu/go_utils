package bootstrap

import (
	"fmt"
	"log"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// initDatabase 初始化数据库，支持 postgres、mysql、sqlite
func initDatabase(cfg *Config) *gorm.DB {
	driver := strings.ToLower(strings.TrimSpace(cfg.Database.Driver))
	if driver == "" {
		driver = "postgres"
	}

	var dialector gorm.Dialector
	switch driver {
	case "postgres", "pg":
		dialector = openPostgres(cfg)
	case "mysql":
		dialector = openMySQL(cfg)
	case "sqlite":
		dialector = openSQLite(cfg)
	default:
		log.Printf("warn: unknown database driver %q, fallback to postgres", cfg.Database.Driver)
		dialector = openPostgres(cfg)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Printf("warn: database connect failed: %v (app will run without DB)", err)
		return nil
	}
	return db
}

func openPostgres(cfg *Config) gorm.Dialector {
	sslmode := cfg.Database.SSLMode
	if sslmode == "" {
		sslmode = "disable"
	}
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.Database, sslmode,
	)
	return postgres.Open(dsn)
}

func openMySQL(cfg *Config) gorm.Dialector {
	port := cfg.Database.Port
	if port <= 0 {
		port = 3306
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.User, cfg.Database.Password,
		cfg.Database.Host, port, cfg.Database.Database,
	)
	return mysql.Open(dsn)
}

func openSQLite(cfg *Config) gorm.Dialector {
	path := cfg.Database.Path
	if path == "" {
		path = cfg.Database.Database
	}
	if path == "" {
		path = "app.db"
	}
	return sqlite.Open(path)
}

// AutoMigrate 自动迁移模型（传入需迁移的模型指针）
func AutoMigrate(db *gorm.DB, models ...any) error {
	if db == nil {
		return nil
	}
	return db.AutoMigrate(models...)
}
