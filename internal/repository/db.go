package repository

import (
	"fmt"
	"log"
	"time"

	"agnos-assessment/internal/config"
	"agnos-assessment/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Bangkok",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode,
	)

	var db *gorm.DB
	var err error

	// Retry loop for Docker startup when Postgres is coming up
	for i := 0; i < 15; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err == nil {
			sqlDB, pingErr := db.DB()
			if pingErr == nil && sqlDB.Ping() == nil {
				log.Println("Connected to PostgreSQL successfully.")
				break
			}
		}
		log.Printf("Waiting for PostgreSQL database... retry %d/15\n", i+1)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate tables
	if err := db.AutoMigrate(&model.Hospital{}, &model.Staff{}, &model.Patient{}); err != nil {
		return nil, fmt.Errorf("auto migration failed: %w", err)
	}

	log.Println("Database AutoMigrate completed successfully.")
	return db, nil
}
