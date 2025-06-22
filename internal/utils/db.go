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
	dsn := config.GetDatabaseDSN()
	if dsn == "" {
		log.Println("[FeasOJ] Database connection string is empty, please check config file")
		return nil
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println("[FeasOJ] Database connection failed, please check database config in config.json")
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Println("[FeasOJ] Failed to get database object")
		return nil
	}

	sqlDB.SetMaxIdleConns(config.GetMaxIdleConns())
	sqlDB.SetMaxOpenConns(config.GetMaxOpenConns())
	sqlDB.SetConnMaxLifetime(time.Duration(config.GetMaxLifeTime()) * time.Second)
	return db
}
