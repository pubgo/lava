package watchcmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/pubgo/redant"

	"github.com/pubgo/lava/v2/pkg/cliutil"
)

func New() *redant.Command {
	return &redant.Command{
		Use:     "watch",
		Aliases: []string{"w"},
		Short:   cliutil.UsageDesc("Watch files for changes and run commands"),
		Long:    "Watch files for changes and run commands automatically",
		Handler: func(ctx context.Context, i *redant.Invocation) error {
			// 简单实现：使用命令行参数
			watchDir := "."
			if len(i.Args) > 0 {
				watchDir = i.Args[0]
			}

			return runWatch(watchDir)
		},
	}
}

func runWatch(watchDir string) error {
	// 创建文件系统监控器
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	// 添加监控目录
	err = addWatchDir(watcher, watchDir, true)
	if err != nil {
		return fmt.Errorf("failed to add watch directory %s: %w", watchDir, err)
	}

	// 监控文件变更
	log.Println("Watching files for changes...")
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// 忽略临时文件
			if strings.HasSuffix(event.Name, "~") || strings.HasPrefix(filepath.Base(event.Name), ".") {
				continue
			}

			// 处理文件变更
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
				log.Printf("File changed: %s", event.Name)

				// 根据文件类型运行不同的命令
				if strings.HasSuffix(event.Name, ".proto") {
					// proto 文件变更
					log.Println("Running protobuild gen...")
					runCommand("protobuild gen")
				} else if strings.HasSuffix(event.Name, ".go") {
					// go 文件变更
					log.Println("Running go build...")
					runCommand("go build ./...")
				}
			}

			// 处理目录创建，添加新目录到监控
			if event.Op&fsnotify.Create != 0 {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					log.Printf("New directory created: %s, adding to watch list", event.Name)
					addWatchDir(watcher, event.Name, true)
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

func addWatchDir(watcher *fsnotify.Watcher, dir string, verbose bool) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 忽略某些目录
		if info.IsDir() {
			name := filepath.Base(path)
			if name == ".git" || name == "node_modules" || name == "vendor" || name == "dist" || name == "build" {
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

func runCommand(cmdStr string) {
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

	// 运行命令
	err := cmd.Run()
	if err != nil {
		log.Printf("Command failed: %v", err)
	} else {
		log.Println("Command completed successfully")
	}
}
