// Package config 负责应用配置的加载与管理。
// 支持从 YAML 文件和环境变量加载配置。
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 是应用的全局配置结构体
type Config struct {
	Server       ServerConfig       `mapstructure:"server"`
	Database     DatabaseConfig     `mapstructure:"database"`
	JWT          JWTConfig          `mapstructure:"jwt"`
	I18n         I18nConfig         `mapstructure:"i18n"`
	Upload       UploadConfig       `mapstructure:"upload"`
	Log          LogConfig          `mapstructure:"log"`
	Registration RegistrationConfig `mapstructure:"registration"`
	ExchangeRate ExchangeRateConfig `mapstructure:"exchange_rate"`
}

// ExchangeRateConfig 汇率参考服务配置
type ExchangeRateConfig struct {
	APIURL          string        `mapstructure:"api_url"`          // frankfurter.app API 地址
	RefreshInterval time.Duration `mapstructure:"refresh_interval"` // 缓存刷新间隔
	Timeout         time.Duration `mapstructure:"timeout"`          // HTTP 请求超时
}

// RegistrationConfig 注册策略配置
type RegistrationConfig struct {
	Mode string `mapstructure:"mode"` // open（公开注册）| invite_only（邀请制）| closed（完全关闭）
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
	Level  string `mapstructure:"level"`  // debug/info/warn/error
	Format string `mapstructure:"format"` // json/console

	// 文件输出与轮转（留空 FilePath 则仅输出到 stderr）
	FilePath   string `mapstructure:"file_path"`   // 日志文件路径，如 ./logs/gastrack.log
	MaxSize    int    `mapstructure:"max_size"`     // 单个日志文件最大大小（MB），超过后自动轮转
	MaxAge     int    `mapstructure:"max_age"`      // 旧日志文件保留天数，0 表示不按时间清理
	MaxBackups int    `mapstructure:"max_backups"`  // 保留的旧日志文件最大数量，0 表示全部保留
	Compress   bool   `mapstructure:"compress"`     // 是否压缩旧日志文件（gzip）
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
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 显式绑定关键环境变量（确保嵌套 key 生效）
	bindEnvKeys(v)

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
	v.SetDefault("log.file_path", "")    // 默认不写文件，仅 stderr
	v.SetDefault("log.max_size", 100)    // 100MB
	v.SetDefault("log.max_age", 30)      // 保留 30 天
	v.SetDefault("log.max_backups", 10)  // 最多 10 个备份
	v.SetDefault("log.compress", true)   // 压缩旧日志

	// 注册策略
	v.SetDefault("registration.mode", "invite_only") // 内测阶段默认邀请制

	// 汇率参考
	v.SetDefault("exchange_rate.api_url", "https://api.frankfurter.app")
	v.SetDefault("exchange_rate.refresh_interval", 24*time.Hour)
	v.SetDefault("exchange_rate.timeout", 10*time.Second)
}

// bindEnvKeys 显式绑定所有嵌套配置 key 到 GASTRACK_ 前缀的环境变量。
// 例如 "database.host" 绑定到 GASTRACK_DATABASE_HOST。
func bindEnvKeys(v *viper.Viper) {
	keys := []string{
		"server.host", "server.port", "server.cors_origins",
		"database.host", "database.port", "database.user",
		"database.password", "database.dbname", "database.sslmode",
		"jwt.secret",
		"log.level", "log.format", "log.file_path",
		"registration.mode",
	}
	for _, key := range keys {
		v.BindEnv(key)
	}
}
