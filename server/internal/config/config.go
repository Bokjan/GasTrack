// Package config 负责应用配置的加载与管理。
// 支持从 YAML 文件和环境变量加载配置。
package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config 是应用的全局配置结构体
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	I18n     I18nConfig     `mapstructure:"i18n"`
	Upload   UploadConfig   `mapstructure:"upload"`
	Log      LogConfig      `mapstructure:"log"`
}

// ServerConfig 服务器相关配置
type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	CORSOrigins     []string      `mapstructure:"cors_origins"`
}

// DatabaseConfig 数据库连接配置
type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	SSLMode      string `mapstructure:"sslmode"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

// DSN 返回 PostgreSQL 连接字符串
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode,
	)
}

// JWTConfig JWT 认证相关配置
type JWTConfig struct {
	Secret            string        `mapstructure:"secret"`
	AccessExpiration  time.Duration `mapstructure:"access_expiration"`
	RefreshExpiration time.Duration `mapstructure:"refresh_expiration"`
	Issuer            string        `mapstructure:"issuer"`
}

// I18nConfig 国际化配置
type I18nConfig struct {
	DefaultLocale    string   `mapstructure:"default_locale"`
	SupportedLocales []string `mapstructure:"supported_locales"`
}

// UploadConfig 文件上传配置
type UploadConfig struct {
	MaxFileSize int64  `mapstructure:"max_file_size"` // 单位 bytes
	UploadDir   string `mapstructure:"upload_dir"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `mapstructure:"level"` // debug/info/warn/error
	Format string `mapstructure:"format"` // json/console
}

// Load 从配置文件和环境变量加载配置
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// 设置默认值
	setDefaults(v)

	// 配置文件
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
		v.AddConfigPath("/etc/gastrack")
	}

	// 环境变量
	v.SetEnvPrefix("GASTRACK")
	v.AutomaticEnv()

	// 读取配置文件（不存在也可以，靠默认值和环境变量）
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	return &cfg, nil
}

// setDefaults 设置配置默认值
func setDefaults(v *viper.Viper) {
	// 服务器
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8098)
	v.SetDefault("server.read_timeout", 15*time.Second)
	v.SetDefault("server.write_timeout", 15*time.Second)
	v.SetDefault("server.shutdown_timeout", 10*time.Second)
	v.SetDefault("server.cors_origins", []string{"http://localhost:3000", "http://localhost:5173"})

	// 数据库
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "gastrack")
	v.SetDefault("database.password", "gastrack")
	v.SetDefault("database.dbname", "gastrack")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)

	// JWT
	v.SetDefault("jwt.secret", "change-me-in-production")
	v.SetDefault("jwt.access_expiration", 15*time.Minute)
	v.SetDefault("jwt.refresh_expiration", 7*24*time.Hour)
	v.SetDefault("jwt.issuer", "gastrack")

	// 多语言
	v.SetDefault("i18n.default_locale", "en-US")
	v.SetDefault("i18n.supported_locales", []string{"en-US", "zh-CN", "ja-JP"})

	// 上传
	v.SetDefault("upload.max_file_size", 5*1024*1024) // 5MB
	v.SetDefault("upload.upload_dir", "./uploads")

	// 日志
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
}
