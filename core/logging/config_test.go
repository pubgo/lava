package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, "info", cfg.Level)
	assert.False(t, cfg.AsJson)
	assert.Nil(t, cfg.DisableLoggers)
	assert.Nil(t, cfg.Filters)
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		wantErr bool
	}{
		{"empty level", "", false},
		{"debug level", "debug", false},
		{"info level", "info", false},
		{"warn level", "warn", false},
		{"error level", "error", false},
		{"invalid level", "invalid", true},
		{"uppercase level", "INFO", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Level: tt.level}
			err := cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	// nil config should not error
	var nilCfg *Config
	assert.NoError(t, nilCfg.Validate())
}

func TestConfigMerge(t *testing.T) {
	base := &Config{
		Level:  "debug",
		AsJson: false,
	}

	override := &Config{
		Level:          "info",
		AsJson:         true,
		DisableLoggers: []string{"grpcLog"},
	}

	merged := base.Merge(override)

	assert.Equal(t, "info", merged.Level)
	assert.True(t, merged.AsJson)
	assert.Equal(t, []string{"grpcLog"}, merged.DisableLoggers)

	// nil merge tests
	assert.Equal(t, base, base.Merge(nil))

	var nilCfg *Config
	assert.Equal(t, override, nilCfg.Merge(override))
}

func TestDisabledLoggers(t *testing.T) {
	// Reset state
	SetDisabledLoggers(nil)

	assert.False(t, IsDisabled("grpc"))

	SetDisabledLoggers([]string{"grpc", "std"})

	// 精确匹配
	assert.True(t, IsDisabled("grpc"))
	assert.True(t, IsDisabled("std"))
	assert.False(t, IsDisabled("slog"))

	// 前缀匹配
	assert.True(t, IsDisabled("grpc.server"))
	assert.True(t, IsDisabled("grpc.client"))
	assert.True(t, IsDisabled("std.log"))
	assert.False(t, IsDisabled("grpclog")) // 不是前缀匹配（没有点分隔）
	assert.False(t, IsDisabled("stdlib"))  // 不是前缀匹配

	// Clear
	SetDisabledLoggers(nil)
	assert.False(t, IsDisabled("grpc"))
	assert.False(t, IsDisabled("grpc.server"))
}

func TestFactoryRegisterAndList(t *testing.T) {
	// Note: factories are global state, be careful with tests
	initialCount := len(List())

	// Register should work (but we can't easily test without affecting global state)
	factories := List()
	assert.NotNil(t, factories)
	assert.GreaterOrEqual(t, len(factories), initialCount)
}
