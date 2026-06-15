package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool 全局数据库连接池
var Pool *pgxpool.Pool

// Connect 连接数据库
func Connect(dsn string) error {
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL 未设置，请在 .env 文件中配置数据库连接字符串")
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("解析数据库连接字符串失败: %w", err)
	}

	// 连接池配置
	config.MaxConns = 10
	config.MinConns = 2

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("创建数据库连接池失败: %w", err)
	}

	// 测试连接
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	Pool = pool
	log.Println("[DB] PostgreSQL 连接池已建立")
	return nil
}

// Close 关闭数据库连接池
func Close() {
	if Pool != nil {
		Pool.Close()
		log.Println("[DB] 数据库连接池已关闭")
	}
}

// GetPool 获取连接池实例
func GetPool() *pgxpool.Pool {
	return Pool
}
