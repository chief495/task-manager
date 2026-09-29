package database

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/eduardnesterov/task-manager/internal/models"
	"github.com/eduardnesterov/task-manager/internal/platform/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func isDebugMode() bool {
	envLocal := os.Getenv("TM_DEBUG")
	return envLocal != "false" && envLocal != "0" && envLocal != ""
}

func Open(cfg *config.AppConfig, debug bool) (*gorm.DB, error) {
	dsn := buildDSN(cfg, debug)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Database connection established")
	return db, nil
}

func buildDSN(cfg *config.AppConfig, debug bool) string {
	if debug {
		return fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Yekaterinburg",
			cfg.DBHost, cfg.DBUser, cfg.DBPassword,
			cfg.DBName, cfg.DBPort, cfg.DBSSLMode,
		)
	}

	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword,
		cfg.DBName, cfg.DBPort, cfg.DBSSLMode,
		os.Getenv("DB_TZ"),
	)
}

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.User{},
		&models.Organization{},
		&models.OrganizationContact{},
		&models.Stage{},
		&models.StageTransition{},
		&models.Status{},
		&models.StatusTransition{},
		&models.WorkType{},
		&models.SystemType{},
		&models.ProductStage{},
		&models.Project{},
		&models.ProjectMember{},
		&models.ProjectAccess{},
		&models.ProjectMessage{},
		&models.ProjectPhase{},
		&models.Stakeholder{},
		&models.System{},
		&models.Product{},
		&models.Task{},
		&models.Report{},
		&models.SocialRatingLog{},
		&models.Trip{},
		&models.TaskTemplate{},
		&models.Participant{},
		&models.LaborCategory{},
		&models.LaborNorm{},
		&models.LaborCoefficient{},
		&models.LaborCalculation{},
		&models.LaborCalculationEntry{},
		&models.LaborCalculationEntryCoeff{},
		&models.LaborCalculationGlobalCoeff{},
	)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	return nil
}
