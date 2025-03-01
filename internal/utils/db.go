package utils

import (
	"log"
	"main/internal/config"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 返回数据库连接对象
func ConnectSql() *gorm.DB {
	dsn := config.LoadSqlConfig()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println("[FeasOJ] Database connection failed, please go to config.xml manually to configure.")
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Println("[FeasOJ] Failed to get generic database object.")
		return nil
	}

	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(config.MaxLifeTime * time.Second)
	return db
}
