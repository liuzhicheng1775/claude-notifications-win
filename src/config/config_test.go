package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// setupConfigEnv 创建一个隔离的临时目录作为 LOCALAPPDATA/APPDATA，
// 并返回配置文件应写入的目录（<tmp>/claude-notifications-win/）。
// 测试结束后恢复环境变量并清理临时目录。
// Go 1.14 没有 t.Setenv / t.TempDir，需手动管理。
func setupConfigEnv(t *testing.T) (string, func()) {
	t.Helper()

	tmpDir, err := ioutil.TempDir("", "cfgtest")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}

	cfgDir := filepath.Join(tmpDir, "claude-notifications-win")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("创建配置目录失败: %v", err)
	}

	savedLocal := os.Getenv("LOCALAPPDATA")
	savedApp := os.Getenv("APPDATA")
	os.Setenv("LOCALAPPDATA", tmpDir)
	os.Setenv("APPDATA", tmpDir)

	cleanup := func() {
		os.Setenv("LOCALAPPDATA", savedLocal)
		os.Setenv("APPDATA", savedApp)
		os.RemoveAll(tmpDir)
	}
	return cfgDir, cleanup
}

// writeConfig 将给定内容写入配置目录下的 config.json。
func writeConfig(t *testing.T, cfgDir, content string) {
	t.Helper()
	path := filepath.Join(cfgDir, "config.json")
	if err := ioutil.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("写入配置文件失败: %v", err)
	}
}

func TestLoad_DefaultsWhenMissing(t *testing.T) {
	_, cleanup := setupConfigEnv(t)
	defer cleanup()
	// 不写入任何配置文件

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() 返回意外错误: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() 返回 nil 配置")
	}
	if !cfg.Notifications.Stop.Enabled {
		t.Errorf("默认 Stop.Enabled 期望 true, 实际 false")
	}
	if !cfg.Notifications.Permission.Enabled {
		t.Errorf("默认 Permission.Enabled 期望 true, 实际 false")
	}
	if cfg.Notifications.Feishu.Enabled {
		t.Errorf("默认 Feishu.Enabled 期望 false, 实际 true")
	}
}

func TestLoad_FeishuConfig(t *testing.T) {
	cfgDir, cleanup := setupConfigEnv(t)
	defer cleanup()
	writeConfig(t, cfgDir, `{"notifications":{"feishu":{"enabled":true,"webhook":"https://example.com/hook","secret":"sec"}}}`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() 返回意外错误: %v", err)
	}
	if !cfg.Notifications.Feishu.Enabled {
		t.Errorf("Feishu.Enabled 期望 true, 实际 false")
	}
	if cfg.Notifications.Feishu.Webhook != "https://example.com/hook" {
		t.Errorf("Feishu.Webhook 期望 https://example.com/hook, 实际 %q", cfg.Notifications.Feishu.Webhook)
	}
	if cfg.Notifications.Feishu.Secret != "sec" {
		t.Errorf("Feishu.Secret 期望 sec, 实际 %q", cfg.Notifications.Feishu.Secret)
	}
}

func TestLoad_ValidConfig(t *testing.T) {
	cfgDir, cleanup := setupConfigEnv(t)
	defer cleanup()
	writeConfig(t, cfgDir, `{"notifications":{"stop":{"enabled":false},"permission":{"enabled":false}}}`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() 返回意外错误: %v", err)
	}
	if cfg.Notifications.Stop.Enabled {
		t.Errorf("Stop.Enabled 期望 false, 实际 true")
	}
	if cfg.Notifications.Permission.Enabled {
		t.Errorf("Permission.Enabled 期望 false, 实际 true")
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	cfgDir, cleanup := setupConfigEnv(t)
	defer cleanup()
	writeConfig(t, cfgDir, "{invalid json")

	cfg, err := Load()
	if err == nil {
		t.Fatal("Load() 期望返回错误, 实际 nil")
	}
	if cfg != nil {
		t.Errorf("Load() 错误时期望 nil 配置, 实际 %v", cfg)
	}
}

func TestLoad_PartialFields(t *testing.T) {
	cfgDir, cleanup := setupConfigEnv(t)
	defer cleanup()
	// 只配置 stop，permission 缺失应为零值 false
	writeConfig(t, cfgDir, `{"notifications":{"stop":{"enabled":true}}}`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() 返回意外错误: %v", err)
	}
	if !cfg.Notifications.Stop.Enabled {
		t.Errorf("Stop.Enabled 期望 true, 实际 false")
	}
	if cfg.Notifications.Permission.Enabled {
		t.Errorf("缺失的 Permission.Enabled 期望零值 false, 实际 true")
	}
}
