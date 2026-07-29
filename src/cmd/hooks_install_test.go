package cmd

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

const fakeExe = `C:\fake\notify.exe`

// fakeExeCmd 是 fakeExe 对应的 bash 兼容 hook 命令前缀。
const fakeExeCmd = `"C:/fake/notify.exe"`

func TestInstallHooks_EmptyFile(t *testing.T) {
	dir, _ := ioutil.TempDir("", "hooktest")
	defer os.RemoveAll(dir)
	settingsPath := filepath.Join(dir, "settings.json")

	if err := installHooks(settingsPath, fakeExe); err != nil {
		t.Fatalf("installHooks 失败: %v", err)
	}

	data := loadSettings(t, settingsPath)
	hooks, ok := data["hooks"].(map[string]interface{})
	if !ok {
		t.Fatal("缺少 hooks 字段")
	}

	stopCmds := extractCommands(t, hooks, "Stop")
	if len(stopCmds) != 1 || stopCmds[0] != fakeExeCmd+" stop" {
		t.Errorf("Stop 命令期望 [%s stop], 实际 %v", fakeExeCmd, stopCmds)
	}

	notifGroups, _ := hooks["Notification"].([]interface{})
	if len(notifGroups) != 1 {
		t.Fatalf("Notification 应有 1 个 group, 实际 %d", len(notifGroups))
	}
	gm := notifGroups[0].(map[string]interface{})
	if gm["matcher"] != "permission_prompt" {
		t.Errorf("matcher 期望 permission_prompt, 实际 %v", gm["matcher"])
	}
}

func TestInstallHooks_NoDuplicate(t *testing.T) {
	dir, _ := ioutil.TempDir("", "hooktest")
	defer os.RemoveAll(dir)
	settingsPath := filepath.Join(dir, "settings.json")

	installHooks(settingsPath, fakeExe)
	if err := installHooks(settingsPath, fakeExe); err != nil {
		t.Fatalf("第二次 installHooks 失败: %v", err)
	}

	data := loadSettings(t, settingsPath)
	hooks := data["hooks"].(map[string]interface{})
	stopCmds := extractCommands(t, hooks, "Stop")
	if len(stopCmds) != 1 {
		t.Errorf("重复安装不应增加 Stop hook, 实际 %d 个: %v", len(stopCmds), stopCmds)
	}
	notifCmds := extractCommands(t, hooks, "Notification")
	if len(notifCmds) != 1 {
		t.Errorf("重复安装不应增加 Notification hook, 实际 %d 个: %v", len(notifCmds), notifCmds)
	}
}

func TestInstallHooks_PreserveOtherFields(t *testing.T) {
	dir, _ := ioutil.TempDir("", "hooktest")
	defer os.RemoveAll(dir)
	settingsPath := filepath.Join(dir, "settings.json")
	initial := `{"env":{"FOO":"bar"},"model":"test","hooks":{"Stop":[{"hooks":[{"type":"command","command":"other.exe"}]}]}}`
	ioutil.WriteFile(settingsPath, []byte(initial), 0644)

	if err := installHooks(settingsPath, fakeExe); err != nil {
		t.Fatalf("installHooks 失败: %v", err)
	}

	data := loadSettings(t, settingsPath)
	env, _ := data["env"].(map[string]interface{})
	if env["FOO"] != "bar" {
		t.Errorf("env.FOO 应保留为 bar, 实际 %v", env["FOO"])
	}
	if data["model"] != "test" {
		t.Errorf("model 应保留, 实际 %v", data["model"])
	}

	hooks := data["hooks"].(map[string]interface{})
	stopGroups, _ := hooks["Stop"].([]interface{})
	if len(stopGroups) != 2 {
		t.Errorf("Stop 应有 2 个 group (other + notify), 实际 %d", len(stopGroups))
	}
	stopCmds := extractCommands(t, hooks, "Stop")
	found := false
	for _, c := range stopCmds {
		if c == fakeExeCmd+" stop" {
			found = true
		}
	}
	if !found {
		t.Errorf("应包含 notify.exe stop, 实际 %v", stopCmds)
	}
}

func TestInstallHooks_PreserveOtherMatcher(t *testing.T) {
	dir, _ := ioutil.TempDir("", "hooktest")
	defer os.RemoveAll(dir)
	settingsPath := filepath.Join(dir, "settings.json")
	initial := `{"hooks":{"Notification":[{"matcher":"other","hooks":[{"type":"command","command":"x.exe"}]}]}}`
	ioutil.WriteFile(settingsPath, []byte(initial), 0644)

	installHooks(settingsPath, fakeExe)

	data := loadSettings(t, settingsPath)
	hooks := data["hooks"].(map[string]interface{})
	notifGroups, _ := hooks["Notification"].([]interface{})
	if len(notifGroups) != 2 {
		t.Errorf("Notification 应有 2 个 group (other + permission_prompt), 实际 %d", len(notifGroups))
	}
	// 验证 permission_prompt 存在
	foundMatcher := false
	for _, g := range notifGroups {
		gm, _ := g.(map[string]interface{})
		if gm["matcher"] == "permission_prompt" {
			foundMatcher = true
		}
	}
	if !foundMatcher {
		t.Error("应包含 permission_prompt matcher")
	}
}

func TestInstallHooks_ReplacesLegacyCommands(t *testing.T) {
	dir, _ := ioutil.TempDir("", "hooktest")
	defer os.RemoveAll(dir)
	settingsPath := filepath.Join(dir, "settings.json")
	// 旧版本写入的 bash 不兼容命令（反斜杠未加引号）+ 一个无关命令
	initial := `{"hooks":{` +
		`"Stop":[{"hooks":[{"type":"command","command":"C:\\fake\\notify.exe stop"}]}],` +
		`"Notification":[{"matcher":"permission_prompt","hooks":[{"type":"command","command":"C:/fake/notify.exe permission"}]},` +
		`{"matcher":"other","hooks":[{"type":"command","command":"other.exe"}]}]` +
		`}}`
	ioutil.WriteFile(settingsPath, []byte(initial), 0644)

	if err := installHooks(settingsPath, fakeExe); err != nil {
		t.Fatalf("installHooks 失败: %v", err)
	}

	data := loadSettings(t, settingsPath)
	hooks := data["hooks"].(map[string]interface{})

	stopCmds := extractCommands(t, hooks, "Stop")
	if len(stopCmds) != 1 || stopCmds[0] != fakeExeCmd+" stop" {
		t.Errorf("旧 Stop 命令应被替换为 [%s stop], 实际 %v", fakeExeCmd, stopCmds)
	}

	notifCmds := extractCommands(t, hooks, "Notification")
	if len(notifCmds) != 2 {
		t.Fatalf("Notification 应有 2 个命令 (notify + other), 实际 %v", notifCmds)
	}
	foundNew, foundOther := false, false
	for _, c := range notifCmds {
		if c == fakeExeCmd+" permission" {
			foundNew = true
		}
		if c == "other.exe" {
			foundOther = true
		}
	}
	if !foundNew {
		t.Errorf("旧 Notification 命令应被替换为 [%s permission], 实际 %v", fakeExeCmd, notifCmds)
	}
	if !foundOther {
		t.Errorf("无关命令 other.exe 应保留, 实际 %v", notifCmds)
	}
}

// loadSettings 读 settings.json 解析为 map。
func loadSettings(t *testing.T, path string) map[string]interface{} {
	t.Helper()
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("读 settings 失败: %v", err)
	}
	var data map[string]interface{}
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("解析 settings 失败: %v", err)
	}
	return data
}

// extractCommands 提取某 event 下所有 command 字符串。
func extractCommands(t *testing.T, hooks map[string]interface{}, event string) []string {
	t.Helper()
	groups, _ := hooks[event].([]interface{})
	var cmds []string
	for _, g := range groups {
		gm, _ := g.(map[string]interface{})
		inner, _ := gm["hooks"].([]interface{})
		for _, h := range inner {
			hm, _ := h.(map[string]interface{})
			if c, _ := hm["command"].(string); c != "" {
				cmds = append(cmds, c)
			}
		}
	}
	return cmds
}
