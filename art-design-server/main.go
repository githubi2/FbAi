package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"

	"github.com/githubi2/FbAi/art-design-server/config"
	"github.com/githubi2/FbAi/art-design-server/db"
	"github.com/githubi2/FbAi/art-design-server/routes"
	"github.com/githubi2/FbAi/art-design-server/services"
)

func main() {
	// 加载 .env 文件（如果存在）
	_ = godotenv.Load()

	cfg := config.DefaultConfig()

	// 连接数据库
	if cfg.Database.DSN != "" {
		if err := db.Connect(cfg.Database.DSN); err != nil {
			log.Printf("[WARN] 数据库连接失败: %v（将使用内存数据）", err)
		} else {
			defer db.Close()
		}
	} else {
		log.Println("[WARN] DATABASE_URL 未配置，将使用内存数据")
	}

	// 尝试升级为 Redis 分布式限速（如果配置了 REDIS_URL）
	services.TryUpgradeToRedis()

	router := routes.SetupRouter()

	addr := cfg.Server.Port
	fmt.Printf("╔══════════════════════════════════════════════╗\n")
	fmt.Printf("║   art-design-server  v1.0.0                  ║\n")
	fmt.Printf("║   Gin Framework Backend                      ║\n")
	fmt.Printf("║   监听地址: http://localhost%s              ║\n", addr)
	fmt.Printf("║   健康检查: http://localhost%s/api/v1/ping   ║\n", addr)
	fmt.Printf("║   API 文档: http://localhost%s/api/v1/       ║\n", addr)
	fmt.Printf("╚══════════════════════════════════════════════╝\n")

	if err := router.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
