package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"webapi/internal/logger"

	_ "github.com/lib/pq"
)

const currentDBVersion = 1

var db *sql.DB

func Init(databaseURL string) (*sql.DB, error) {
	var err error
	db, err = sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// 检查并初始化数据库
	if err := initializeDatabase(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	logger.Info("Database initialized successfully")
	return db, nil
}

func GetDB() *sql.DB {
	return db
}

func initializeDatabase() error {
	// 检查是否已初始化
	var version int
	err := db.QueryRow("SELECT version FROM general LIMIT 1").Scan(&version)
	
	if err != nil {
		if err == sql.ErrNoRows {
			// 数据库未初始化，执行初始化脚本
			return executeInitScript()
		}
		// 表不存在，执行初始化脚本
		return executeInitScript()
	}

	// 执行迁移
	if version < currentDBVersion {
		return migrateDatabase(version, currentDBVersion)
	}

	return nil
}

func executeInitScript() error {
	logger.Info("Initializing database...")
	
	initSQL, err := ioutil.ReadFile(filepath.Join("sql", "init.sql"))
	if err != nil {
		return fmt.Errorf("failed to read init.sql: %w", err)
	}

	_, err = db.Exec(string(initSQL))
	if err != nil {
		return fmt.Errorf("failed to execute init.sql: %w", err)
	}

	logger.Info("Database initialized")
	return nil
}

func migrateDatabase(from, to int) error {
	logger.Infof("Migrating database from version %d to %d", from, to)
	
	// 这里可以添加迁移逻辑
	// 目前没有迁移脚本，所以直接返回
	
	return nil
}