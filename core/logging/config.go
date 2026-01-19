package logging

import "fmt"

// LogConfigLoader 配置加载器
type LogConfigLoader struct {
	Log *Config `yaml:"logger"`
}

// Config 日志配置
type Config struct {
	// Level 日志级别: trace, debug, info, warn, error, fatal, panic
	Level string `yaml:"level"`
	// AsJson 终端是否以 JSON 格式输出
	AsJson bool `yaml:"as_json"`
	// DisableLoggers 要禁用日志输出的 logger 名称列表
	// 支持前缀匹配，如 "grpc" 会禁用 "grpc", "grpc.server" 等
	DisableLoggers []string `yaml:"disable_loggers"`
	// Filters expr 表达式过滤器列表，多个表达式以 && 连接
	Filters []string `yaml:"filters"`
	// File 文件输出配置，为空则不输出到文件
	File *FileConfig `yaml:"file"`
}

// FileConfig 文件输出配置
type FileConfig struct {
	// Enabled 是否启用文件输出
	Enabled bool `yaml:"enabled"`
	// Path 日志文件路径
	Path string `yaml:"path"`
	// MaxSize 单个日志文件最大大小（MB）
	MaxSize int `yaml:"max_size"`
	// MaxBackups 保留的旧日志文件最大数量
	MaxBackups int `yaml:"max_backups"`
	// MaxAge 保留旧日志文件的最大天数
	MaxAge int `yaml:"max_age"`
	// Compress 是否压缩旧日志文件
	Compress bool `yaml:"compress"`
}

// DefaultFileConfig 返回默认文件配置
func DefaultFileConfig() *FileConfig {
	return &FileConfig{
		Enabled:    false,
		Path:       "logs/app.log",
		MaxSize:    100, // 100MB
		MaxBackups: 10,
		MaxAge:     30, // 30 天
		Compress:   true,
	}
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Level:  "info",
		AsJson: false,
		File:   DefaultFileConfig(),
	}
}

// Validate 验证配置有效性
func (c *Config) Validate() error {
	if c == nil {
		return nil
	}

	validLevels := map[string]bool{
		"":      true, // 空字符串使用默认值
		"trace": true,
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
		"panic": true,
	}

	if !validLevels[c.Level] {
		return fmt.Errorf("invalid log level: %s", c.Level)
	}

	// 验证文件配置
	if c.File != nil && c.File.Enabled {
		if c.File.Path == "" {
			return fmt.Errorf("file path is required when file output is enabled")
		}
		if c.File.MaxSize <= 0 {
			c.File.MaxSize = 100
		}
		if c.File.MaxBackups <= 0 {
			c.File.MaxBackups = 10
		}
		if c.File.MaxAge <= 0 {
			c.File.MaxAge = 30
		}
	}

	return nil
}

// Merge 合并配置，优先使用 other 的非零值
func (c *Config) Merge(other *Config) *Config {
	if other == nil {
		return c
	}
	if c == nil {
		return other
	}

	result := *c
	if other.Level != "" {
		result.Level = other.Level
	}
	if other.AsJson {
		result.AsJson = other.AsJson
	}
	if len(other.DisableLoggers) > 0 {
		result.DisableLoggers = other.DisableLoggers
	}
	if len(other.Filters) > 0 {
		result.Filters = other.Filters
	}
	if other.File != nil {
		result.File = other.File
	}
	return &result
}

// GetLogFilePath 获取日志文件路径（用于 loggerdebug 读取）
func (c *Config) GetLogFilePath() string {
	if c.File != nil && c.File.Enabled {
		return c.File.Path
	}
	return ""
}
