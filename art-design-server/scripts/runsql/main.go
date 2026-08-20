// runsql — 使用 .env 中的 DATABASE_URL 执行指定 SQL 文件（迁移手动执行工具）
// 用法: go run ./scripts/runsql migrations/011_fb_pages_cache.sql
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run ./scripts/runsql <sql文件路径>")
		os.Exit(1)
	}

	_ = godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Println("DATABASE_URL 未设置")
		os.Exit(1)
	}

	sqlBytes, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("读取 SQL 文件失败: %v\n", err)
		os.Exit(1)
	}

	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		fmt.Printf("连接数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	if _, err := conn.Exec(context.Background(), string(sqlBytes)); err != nil {
		fmt.Printf("执行 SQL 失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("SQL 执行成功: %s\n", os.Args[1])
}
