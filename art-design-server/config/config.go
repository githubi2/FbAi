package config

import "os"

// Config 应用配置
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

// ServerConfig 服务配置
type ServerConfig struct {
	Port string // 默认 :9090
	Mode string // debug | release | test
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver string // postgres
	DSN    string // 连接字符串，从环境变量 DATABASE_URL 读取
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret     string
	ExpireHour int // 过期小时数
}

// DefaultConfig 默认配置
// 数据库 DSN 优先从环境变量 DATABASE_URL 读取，fallback 为空（需要在 .env 中配置）
func DefaultConfig() *Config {
	dsn := os.Getenv("DATABASE_URL")

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = ":9090"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "art-design-server-secret-key"
	}

	return &Config{
		Server: ServerConfig{
			Port: port,
			Mode: os.Getenv("GIN_MODE"),
		},
		Database: DatabaseConfig{
			Driver: "postgres",
			DSN:    dsn,
		},
		JWT: JWTConfig{
			Secret:     jwtSecret,
			ExpireHour: 24,
		},
	}
}
