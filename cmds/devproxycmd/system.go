package devproxycmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// InstallSystemIntegration 安装系统集成
func InstallSystemIntegration() error {
	// 检查是否为macOS
	if runtime.GOOS != "darwin" && !strings.Contains(strings.ToLower(os.Getenv("OSTYPE")), "darwin") {
		return fmt.Errorf("system integration is only supported on macOS")
	}

	// 创建/etc/resolver目录
	resolverDir := "/etc/resolver"
	if err := os.MkdirAll(resolverDir, 0o755); err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied: need to run with sudo to modify system files")
		}
		return fmt.Errorf("failed to create resolver directory: %w", err)
	}

	// 创建lava resolver配置
	resolverFile := filepath.Join(resolverDir, "lava")
	configContent := fmt.Sprintf(`nameserver 127.0.0.1
port %d
`, config.DNS.Port)

	if err := os.WriteFile(resolverFile, []byte(configContent), 0o644); err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied: need to run with sudo to modify system files")
		}
		return fmt.Errorf("failed to write resolver config: %w", err)
	}

	fmt.Println("System integration installed successfully")
	fmt.Println("Created:", resolverFile)
	fmt.Println("This config will route all *.lava queries to our DNS server")
	return nil
}

// UninstallSystemIntegration 卸载系统集成
func UninstallSystemIntegration() error {
	// 检查是否为macOS
	if runtime.GOOS != "darwin" && !strings.Contains(strings.ToLower(os.Getenv("OSTYPE")), "darwin") {
		return fmt.Errorf("system integration is only supported on macOS")
	}

	// 删除lava resolver配置
	resolverFile := "/etc/resolver/lava"
	if err := os.Remove(resolverFile); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("System integration not found")
			return nil
		}
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied: need to run with sudo to modify system files")
		}
		return fmt.Errorf("failed to remove resolver config: %w", err)
	}

	fmt.Println("System integration uninstalled successfully")
	fmt.Println("Removed:", resolverFile)
	return nil
}

// ShowRoutes 显示当前路由
func ShowRoutes() error {
	fmt.Println("Current routes:")
	fmt.Println("Pattern\t\tTarget\t\tPath")
	fmt.Println("---------------------------------------------------")
	for _, route := range config.Routes {
		path := route.Path
		if path == "" {
			path = "(original)"
		}
		fmt.Printf("%s\t\t%s\t\t%s\n", route.Pattern, route.Target, path)
	}
	return nil
}
