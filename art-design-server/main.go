package main

import (
	"fmt"
	"log"

	"github.com/githubi2/FbAi/art-design-server/config"
	"github.com/githubi2/FbAi/art-design-server/routes"
)

func main() {
	cfg := config.DefaultConfig()

	// 设置 Gin 运行模式
	// gin.SetMode(cfg.Server.Mode)

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
