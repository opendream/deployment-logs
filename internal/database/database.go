package database

import (
	"fmt"

	"deployment-logs/internal/config"
	"deployment-logs/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func Setup(cfg *config.Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	switch cfg.DBDriver {
	case "mssql", "sqlserver":
		// Connect to master first to ensure the target database exists
		masterDSN := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=master",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort)
		masterDB, err := gorm.Open(sqlserver.Open(masterDSN), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to master: %w", err)
		}
		masterDB.Exec(fmt.Sprintf("IF NOT EXISTS (SELECT name FROM sys.databases WHERE name = '%s') CREATE DATABASE [%s]", cfg.DBName, cfg.DBName))
		sqlDB, _ := masterDB.DB()
		sqlDB.Close()

		// Now connect to the target database
		dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

		gormCfg := &gorm.Config{}
		if cfg.DBSchema != "" {
			gormCfg.NamingStrategy = schema.NamingStrategy{
				TablePrefix: cfg.DBSchema + ".",
			}
		}

		db, err = gorm.Open(sqlserver.Open(dsn), gormCfg)
		if err != nil {
			return nil, err
		}
		if cfg.DBSchema != "" {
			db.Exec(fmt.Sprintf("IF NOT EXISTS (SELECT * FROM sys.schemas WHERE name = '%s') EXEC('CREATE SCHEMA %s')", cfg.DBSchema, cfg.DBSchema))
		}

	case "postgres":
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
		if cfg.DBSchema != "" {
			dsn += fmt.Sprintf(" search_path=%s", cfg.DBSchema)
		}
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		if cfg.DBSchema != "" {
			db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", cfg.DBSchema))
		}

	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	default: // sqlite
		db, err = gorm.Open(sqlite.Open(cfg.DBName+".db"), &gorm.Config{})
	}

	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&models.RepositoryConfig{}, &models.DeploymentLog{}, &models.LogItem{}, &models.User{}, &models.Setting{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
