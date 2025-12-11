package database

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// DatabaseConfig 数据库配置结构
type DatabaseConfig struct {
	// 数据库文件路径
	FilePath string `mapstructure:"file_path"`
	
	// 连接池配置
	ConnectionPool ConnectionPoolConfig `mapstructure:"connection_pool"`
	
	// 缓存配置
	Cache CacheConfig `mapstructure:"cache"`
	
	// 清理策略配置
	Cleanup CleanupConfig `mapstructure:"cleanup"`
	
	// 备份策略配置
	Backup BackupConfig `mapstructure:"backup"`
}

// ConnectionPoolConfig 连接池配置
type ConnectionPoolConfig struct {
	MaxOpenConns    int           `mapstructure:"max_open_conns"`    // 最大打开连接数
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`    // 最大空闲连接数
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"` // 连接最大生命周期
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Enabled bool `mapstructure:"enabled"` // 是否启用缓存
	Size    int  `mapstructure:"size"`    // 缓存大小
	TTL     int  `mapstructure:"ttl"`     // 缓存过期时间
}

// CleanupConfig 清理策略配置
type CleanupConfig struct {
	Enabled bool   `mapstructure:"enabled"` // 是否启用自动清理
	Interval int   `mapstructure:"interval"` // 清理间隔(秒)
	MaxAge   int   `mapstructure:"max_age"`   // 数据最大保留时间(秒)
}

// BackupConfig 备份策略配置
type BackupConfig struct {
	Enabled       bool   `mapstructure:"enabled"`       // 是否启用自动备份
	Interval      int    `mapstructure:"interval"`      // 备份间隔(秒)
	RetentionDays int    `mapstructure:"retention_days"` // 备份保留天数
	Path          string `mapstructure:"path"`          // 备份文件路径
}

// DefaultConfig 获取默认数据库配置
func DefaultConfig() *DatabaseConfig {
	return &DatabaseConfig{
		FilePath: "./data/gh-pitfall-scraper.db",
		ConnectionPool: ConnectionPoolConfig{
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 300 * time.Second,
		},
		Cache: CacheConfig{
			Enabled: true,
			Size:    1000,
			TTL:     3600,
		},
		Cleanup: CleanupConfig{
			Enabled:  true,
			Interval: 86400, // 24小时
			MaxAge:   2592000, // 30天
		},
		Backup: BackupConfig{
			Enabled:       true,
			Interval:      43200, // 12小时
			RetentionDays: 7,
			Path:          "./backups",
		},
	}
}

// LoadConfig 从配置文件加载数据库配置
func LoadConfig(configPath string) (*DatabaseConfig, error) {
	// 设置默认值
	config := DefaultConfig()

	// 如果配置文件不存在，返回默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("配置文件 %s 不存在，使用默认配置", configPath)
		return config, nil
	}

	// 读取配置文件
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 读取配置
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("读取配置文件失败: %v，使用默认配置", err)
		return config, nil
	}

	// 解析配置
	if err := viper.Unmarshal(&config); err != nil {
		log.Printf("解析配置文件失败: %v，使用默认配置", err)
		return config, nil
	}

	// 处理路径
	config = processPaths(config)

	// 验证配置
	if err := validateConfig(config); err != nil {
		log.Printf("配置验证失败: %v，使用默认配置", err)
		return DefaultConfig(), nil
	}

	log.Printf("成功加载数据库配置: %s", config.FilePath)
	return config, nil
}

// processPaths 处理配置中的路径
func processPaths(config *DatabaseConfig) *DatabaseConfig {
	// 处理数据库文件路径
	if !filepath.IsAbs(config.FilePath) {
		// 获取工作目录
		workDir, err := os.Getwd()
		if err != nil {
			log.Printf("获取工作目录失败: %v", err)
		} else {
			config.FilePath = filepath.Join(workDir, config.FilePath)
		}
	}

	// 确保目录存在
	if err := ensureDirExists(filepath.Dir(config.FilePath)); err != nil {
		log.Printf("创建数据库目录失败: %v", err)
	}

	// 处理备份路径
	if !filepath.IsAbs(config.Backup.Path) {
		workDir, err := os.Getwd()
		if err != nil {
			log.Printf("获取工作目录失败: %v", err)
		} else {
			config.Backup.Path = filepath.Join(workDir, config.Backup.Path)
		}
	}

	// 确保备份目录存在
	if err := ensureDirExists(config.Backup.Path); err != nil {
		log.Printf("创建备份目录失败: %v", err)
	}

	return config
}

// ensureDirExists 确保目录存在
func ensureDirExists(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// validateConfig 验证配置
func validateConfig(config *DatabaseConfig) error {
	if config.ConnectionPool.MaxOpenConns <= 0 {
		return NewConfigError("最大打开连接数必须大于0")
	}
	if config.ConnectionPool.MaxIdleConns < 0 {
		return NewConfigError("最大空闲连接数不能小于0")
	}
	if config.ConnectionPool.ConnMaxLifetime <= 0 {
		return NewConfigError("连接最大生命周期必须大于0")
	}
	if config.Cache.Size <= 0 {
		return NewConfigError("缓存大小必须大于0")
	}
	if config.Cache.TTL <= 0 {
		return NewConfigError("缓存TTL必须大于0")
	}
	if config.Cleanup.MaxAge <= 0 {
		return NewConfigError("数据最大保留时间必须大于0")
	}
	if config.Backup.Interval <= 0 {
		return NewConfigError("备份间隔必须大于0")
	}
	if config.Backup.RetentionDays <= 0 {
		return NewConfigError("备份保留天数必须大于0")
	}

	// 检查文件路径是否有效
	if strings.TrimSpace(config.FilePath) == "" {
		return NewConfigError("数据库文件路径不能为空")
	}

	return nil
}

// ConfigError 配置错误
type ConfigError struct {
	message string
}

func NewConfigError(message string) *ConfigError {
	return &ConfigError{message: message}
}

func (e *ConfigError) Error() string {
	return e.message
}

// GetConnectionString 获取连接字符串
func (c *DatabaseConfig) GetConnectionString() string {
	return c.FilePath
}

// IsBackupEnabled 检查是否启用备份
func (c *DatabaseConfig) IsBackupEnabled() bool {
	return c.Backup.Enabled
}

// IsCleanupEnabled 检查是否启用清理
func (c *DatabaseConfig) IsCleanupEnabled() bool {
	return c.Cleanup.Enabled
}

// IsCacheEnabled 检查是否启用缓存
func (c *DatabaseConfig) IsCacheEnabled() bool {
	return c.Cache.Enabled
}