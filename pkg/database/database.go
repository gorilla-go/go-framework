package database

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gorilla-go/go-framework/pkg/config"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var (
	dbInstance *gorm.DB
	dbError    error
	once       sync.Once
)

// Init 初始化数据库连接（全局只能初始化一次）
func Init(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	once.Do(func() {
		dbInstance, dbError = initDB(cfg)
	})
	return dbInstance, dbError
}

// initDB 内部初始化函数
func initDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	// 根据配置的 driver 选择数据库方言
	dialector, err := buildDialector(cfg)
	if err != nil {
		return nil, err
	}

	// 连接数据库
	db, err := gorm.Open(dialector, &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取sqlDB失败: %w", err)
	}

	// 设置最大空闲连接数
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	// 设置最大打开连接数
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	// 设置连接的最大生命周期
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接测试失败: %w", err)
	}

	return db, nil
}

// buildDialector 根据配置的 driver 构建对应的 GORM 方言。
// 支持: mysql（默认）、sqlite。driver 为空时按 mysql 处理，保持向后兼容。
func buildDialector(cfg *config.DatabaseConfig) (gorm.Dialector, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Driver)) {
	case "", "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.Username,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.DBName,
		)
		return mysql.Open(dsn), nil
	case "sqlite", "sqlite3":
		// sqlite 使用 dbname 作为文件路径，支持 ":memory:" 内存库
		dsn := cfg.DBName
		if dsn == "" {
			dsn = "data.db"
		}
		return sqlite.Open(dsn), nil
	default:
		return nil, fmt.Errorf("不支持的数据库驱动: %q（支持 mysql、sqlite）", cfg.Driver)
	}
}
