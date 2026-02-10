package watchcmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pubgo/redant"
	"gopkg.in/yaml.v3"

	"github.com/pubgo/lava/v2/pkg/cliutil"
)

type WatcherConfig struct {
	Name           string   `yaml:"name"`
	Directory      string   `yaml:"directory"`
	Patterns       []string `yaml:"patterns"`
	Commands       []string `yaml:"commands"`
	Ignore         []string `yaml:"ignore"`
	IgnorePatterns []string `yaml:"ignore_patterns"`
	RunOnStartup   bool     `yaml:"run_on_startup"`
	Timeout        int      `yaml:"timeout"`
}

type WatchConfig struct {
	Watchers []WatcherConfig `yaml:"watchers"`
}

type Config struct {
	Watch WatchConfig `yaml:"watch"`
}

func New() *redant.Command {
	return &redant.Command{
		Use:     "watch",
		Aliases: []string{"w"},
		Short:   cliutil.UsageDesc("Watch files for changes and run commands"),
		Long:    "Watch files for changes and run commands automatically",
		Handler: func(ctx context.Context, i *redant.Invocation) error {
			// 加载配置文件
			cfg, err := loadConfig()
			if err != nil {
				log.Printf("Warning: failed to load config: %v, using default config", err)
				// 使用默认配置
				cfg = &Config{
					Watch: WatchConfig{
						Watchers: []WatcherConfig{
							{
								Name:      "default",
								Directory: ".",
								Patterns:  []string{"*.proto", "*.go"},
								Commands: []string{
									"protobuild gen",
									"go build ./...",
								},
								Ignore:         []string{".git", "node_modules", "vendor", "dist", "build"},
								IgnorePatterns: []string{"*.tmp", "*~", ".DS_Store"},
								RunOnStartup:   false,
								Timeout:        30,
							},
						},
					},
				}
			}

			// 如果没有配置 watcher，使用默认配置
			if len(cfg.Watch.Watchers) == 0 {
				cfg.Watch.Watchers = []WatcherConfig{
					{
						Name:      "default",
						Directory: ".",
						Patterns:  []string{"*.proto", "*.go"},
						Commands: []string{
							"protobuild gen",
							"go build ./...",
						},
						Ignore:         []string{".git", "node_modules", "vendor", "dist", "build"},
						IgnorePatterns: []string{"*.tmp", "*~", ".DS_Store"},
						RunOnStartup:   false,
						Timeout:        30,
					},
				}
			}

			// 运行所有 watcher
			return runWatchers(cfg.Watch.Watchers)
		},
	}
}

func loadConfig() (*Config, error) {
	// 按优先级查找配置文件
	configPaths := []string{
		".lava/lava.yaml",
		".lava.yaml",
		"lava.yaml",
	}

	var configPath string
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			configPath = path
			break
		}
	}

	if configPath == "" {
		return nil, fmt.Errorf("no config file found")
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// 解析 YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	return &cfg, nil
}

func runWatchers(watchers []WatcherConfig) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(watchers))

	// 为每个 watcher 启动一个 goroutine
	for _, watcherCfg := range watchers {
		wg.Add(1)
		go func(cfg WatcherConfig) {
			defer wg.Done()
			if err := runWatcher(cfg); err != nil {
				errChan <- fmt.Errorf("watcher %s: %w", cfg.Name, err)
			}
		}(watcherCfg)
	}

	// 等待所有 watcher 完成
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// 收集错误
	var errors []error
	for err := range errChan {
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("watcher errors: %v", errors)
	}

	return nil
}

func runWatcher(cfg WatcherConfig) error {
	// 创建文件系统监控器
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	// 添加监控目录
	err = addWatchDir(watcher, cfg.Directory, cfg.Ignore, true)
	if err != nil {
		return fmt.Errorf("failed to add watch directory %s: %w", cfg.Directory, err)
	}

	// 如果配置了启动时执行命令
	if cfg.RunOnStartup {
		log.Printf("[%s] Running commands on startup...", cfg.Name)
		for _, cmdStr := range cfg.Commands {
			runCommand(cfg.Name, cmdStr, cfg.Timeout)
		}
	}

	// 监控文件变更
	log.Printf("[%s] Watching directory: %s", cfg.Name, cfg.Directory)
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// 检查文件是否匹配 patterns
			if !matchPatterns(event.Name, cfg.Patterns) {
				continue
			}

			// 检查是否应该忽略
			if shouldIgnore(event.Name, cfg.Ignore, cfg.IgnorePatterns) {
				continue
			}

			// 处理文件变更
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
				log.Printf("[%s] File changed: %s", cfg.Name, event.Name)

				// 运行配置的命令
				for _, cmdStr := range cfg.Commands {
					runCommand(cfg.Name, cmdStr, cfg.Timeout)
				}
			}

			// 处理目录创建，添加新目录到监控
			if event.Op&fsnotify.Create != 0 {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					log.Printf("[%s] New directory created: %s, adding to watch list", cfg.Name, event.Name)
					addWatchDir(watcher, event.Name, cfg.Ignore, true)
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			log.Printf("[%s] Watcher error: %v", cfg.Name, err)
		}
	}
}

func addWatchDir(watcher *fsnotify.Watcher, dir string, ignoreDirs []string, verbose bool) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 忽略某些目录
		if info.IsDir() {
			name := filepath.Base(path)
			for _, ignoreDir := range ignoreDirs {
				if name == ignoreDir {
					return filepath.SkipDir
				}
			}

			// 添加目录到监控
			err := watcher.Add(path)
			if err != nil {
				return err
			}

			if verbose {
				log.Printf("Watching directory: %s", path)
			}
		}

		return nil
	})
}

func matchPatterns(filename string, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}

	base := filepath.Base(filename)
	for _, pattern := range patterns {
		matched, err := filepath.Match(pattern, base)
		if err != nil {
			continue
		}
		if matched {
			return true
		}
	}

	return false
}

func shouldIgnore(filename string, ignoreDirs []string, ignorePatterns []string) bool {
	// 检查是否在忽略的目录中
	for _, dir := range ignoreDirs {
		if strings.Contains(filename, "/"+dir+"/") || strings.HasPrefix(filename, dir+"/") {
			return true
		}
	}

	// 检查是否匹配忽略的文件模式
	base := filepath.Base(filename)
	for _, pattern := range ignorePatterns {
		matched, err := filepath.Match(pattern, base)
		if err != nil {
			continue
		}
		if matched {
			return true
		}
	}

	return false
}

func runCommand(watcherName string, cmdStr string, timeout int) {
	log.Printf("[%s] Running command: %s", watcherName, cmdStr)

	// 解析命令
	var cmd *exec.Cmd
	if strings.Contains(cmdStr, " ") {
		parts := strings.Split(cmdStr, " ")
		cmd = exec.Command(parts[0], parts[1:]...)
	} else {
		cmd = exec.Command(cmdStr)
	}

	// 设置命令环境
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	// 运行命令
	err := cmd.Start()
	if err != nil {
		log.Printf("[%s] Command failed to start: %v", watcherName, err)
		return
	}

	// 等待命令完成或超时
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			log.Printf("[%s] Command failed: %v", watcherName, err)
		} else {
			log.Printf("[%s] Command completed successfully", watcherName)
		}
	case <-ctx.Done():
		log.Printf("[%s] Command timeout, killing process", watcherName)
		cmd.Process.Kill()
		<-done
		log.Printf("[%s] Command killed", watcherName)
	}
}
