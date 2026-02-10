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

	"github.com/bmatcuk/doublestar/v4"
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
								Patterns: []string{
									"*.proto",
									"*.go",
									"!**/dist",
									"!**/build",
									"!**/vendor",
									"!**/node_modules",
									"!**/.git",
									"!*.tmp",
									"!*~",
									"!.DS_Store",
								},
								Commands: []string{
									"protobuild gen",
									"go build ./...",
								},
								RunOnStartup: false,
								Timeout:      30,
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
						Patterns: []string{
							"*.proto",
							"*.go",
							"!**/dist",
							"!**/build",
							"!**/vendor",
							"!**/node_modules",
							"!**/.git",
							"!*.tmp",
							"!*~",
							"!.DS_Store",
						},
						Commands: []string{
							"protobuild gen",
							"go build ./...",
						},
						RunOnStartup: false,
						Timeout:      30,
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
	// 合并 pattern
	// 1. patterns
	// 2. !ignore (转换为排除模式)
	// 3. !ignore_patterns (转换为排除模式)
	var finalPatterns []string
	finalPatterns = append(finalPatterns, cfg.Patterns...)

	// 处理 legacy ignore 配置
	for _, ign := range cfg.Ignore {
		finalPatterns = append(finalPatterns, "!"+ign)
	}
	for _, ign := range cfg.IgnorePatterns {
		finalPatterns = append(finalPatterns, "!"+ign)
	}

	// 提取包含和排除列表以便后续使用
	includes, excludes := parsePatterns(finalPatterns)

	// 创建文件系统监控器
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	// 添加监控目录
	err = addWatchDir(watcher, cfg.Directory, excludes, true)
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

			// 检查文件是否匹配
			// 必须匹配某个 include 模式，且不匹配任何 exclude 模式
			if !matchConfig(event.Name, includes, excludes) {
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
					addWatchDir(watcher, event.Name, excludes, true)
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

func parsePatterns(patterns []string) (includes []string, excludes []string) {
	for _, p := range patterns {
		if strings.HasPrefix(p, "!") {
			excludes = append(excludes, strings.TrimPrefix(p, "!"))
		} else {
			includes = append(includes, p)
		}
	}
	// 如果没有 include 模式，默认为匹配所有 (虽然通常会提供 *.go 等)
	// 但如果只提供了排除模式，我们假设用户想要 backup 行为 - (暂不处理，假设必有include)
	return
}

func matchConfig(filename string, includes []string, excludes []string) bool {
	// 1. 检查排除
	if matchAny(filename, excludes) {
		return false
	}

	// 2. 检查包含 (如果是空，默认不匹配任何东西? 或者匹配所有? usually includes required)
	if len(includes) == 0 {
		return true
	}
	return matchAny(filename, includes)
}

func matchAny(filename string, patterns []string) bool {
	for _, pattern := range patterns {
		// 如果模式包含路径分隔符，则尝试匹配完整路径
		if strings.ContainsAny(pattern, "/\\") {
			if matched, _ := doublestar.Match(pattern, filename); matched {
				return true
			}
		} else {
			// 否则仅匹配文件名
			if matched, _ := doublestar.Match(pattern, filepath.Base(filename)); matched {
				return true
			}
		}
	}
	return false
}

func addWatchDir(watcher *fsnotify.Watcher, dir string, excludes []string, verbose bool) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 忽略某些目录
		if info.IsDir() {
			// 如果是根目录，不忽略
			if path == dir {
				// 添加目录到监控
				if err := watcher.Add(path); err != nil {
					return err
				}
				if verbose {
					log.Printf("Watching directory: %s", path)
				}
				return nil
			}

			// 检查是否应该排除该目录
			if matchAny(path, excludes) {
				return filepath.SkipDir
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
