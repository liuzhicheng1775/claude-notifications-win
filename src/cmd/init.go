package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"claude-notifications-win/src/config"
)

// RunInit 交互式配置向导：配置飞书 + 写 config.json + 自动配置 Claude Code hooks。
func RunInit() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Claude Code 通知插件配置向导")
	fmt.Println("================================")

	// 加载现有配置（保留已配的字段）
	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{}
	}

	// 1. 配置飞书
	fmt.Println()
	if promptYes(reader, "是否配置飞书通知？") {
		webhook := prompt(reader, "请输入飞书 webhook URL: ")
		for webhook == "" || !strings.HasPrefix(webhook, "http") {
			fmt.Println("webhook 不能为空且须以 http 开头")
			webhook = prompt(reader, "请输入飞书 webhook URL: ")
		}
		secret := prompt(reader, "请输入加签 secret（可选，回车跳过）: ")

		cfg.Notifications.Feishu.Enabled = true
		cfg.Notifications.Feishu.Webhook = webhook
		cfg.Notifications.Feishu.Secret = secret
		fmt.Println("✓ 飞书配置已记录")
	} else {
		fmt.Println("已跳过飞书配置（保留现有设置）")
	}

	// 2. 写 config.json
	configPath := defaultConfigPath()
	if err := writeConfigFile(cfg, configPath); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}
	fmt.Printf("✓ 配置已保存到 %s\n", configPath)

	// 3. 自动配置 Claude Code hooks
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取 exe 路径失败: %w", err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户目录失败: %w", err)
	}
	settingsPath := filepath.Join(home, ".claude", "settings.json")
	if err := installHooks(settingsPath, exePath); err != nil {
		return fmt.Errorf("配置 hooks 失败: %w", err)
	}
	fmt.Printf("✓ Claude Code hooks 已写入 %s\n", settingsPath)

	// 4. 询问是否立即测试
	fmt.Println()
	if promptYes(reader, "是否立即发送测试通知？") {
		fmt.Println()
		_ = RunTest()
	}

	fmt.Println()
	fmt.Println("================================")
	fmt.Println("配置完成！后续 Claude Code 任务完成 / 请求授权时会自动通知。")
	fmt.Println("提示：修改配置可重新运行 .\\notify.exe init，或直接编辑 config.json。")
	return nil
}

// defaultConfigPath 返回 config.json 应写入的路径（LOCALAPPDATA 优先），
// 与 config.Load 的查找顺序第一优先级一致。
func defaultConfigPath() string {
	local := os.Getenv("LOCALAPPDATA")
	if local == "" {
		local = os.TempDir()
	}
	return filepath.Join(local, "claude-notifications-win", "config.json")
}

// writeConfigFile 将配置以缩进 JSON 写入指定路径，目录不存在则创建。
func writeConfigFile(cfg *config.Config, path string) error {
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}

// prompt 读取一行输入，去除首尾空白（兼容 Windows \r\n）。
func prompt(reader *bufio.Reader, label string) string {
	fmt.Print(label)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

// promptYes 询问 y/n，返回是否为 yes。
func promptYes(reader *bufio.Reader, label string) bool {
	s := prompt(reader, label+" (y/n): ")
	return strings.HasPrefix(strings.ToLower(s), "y")
}
