package config

// Config 应用配置
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

// ServerConfig 服务配置
type ServerConfig struct {
	Port string // 默认 :8080
	Mode string // debug | release | test
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver string // mysql | postgres | sqlite
	DSN    string // 连接字符串
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret     string
	ExpireHour int // 过期小时数
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: ":9090",
			Mode: "debug",
		},
		Database: DatabaseConfig{
			Driver: "sqlite",
			DSN:    "data.db",
		},
		JWT: JWTConfig{
			Secret:     "art-design-server-secret-key",
			ExpireHour: 24,
		},
	}
}
