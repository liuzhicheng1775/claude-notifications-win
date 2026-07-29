package cmd

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

// installHooks 将 notify.exe 的 Stop / Notification hooks 合并进 settings.json。
// 已存在相同命令则不重复添加；保留所有现有字段（env、model、statusLine 等）。
// 旧版本写入的 bash 不兼容命令（反斜杠未加引号）会被自动替换。
func installHooks(settingsPath, exePath string) error {
	// 读现有 settings（不存在则从空 map 开始）
	data := map[string]interface{}{}
	if raw, err := ioutil.ReadFile(settingsPath); err == nil {
		if err := json.Unmarshal(raw, &data); err != nil {
			return err
		}
	}

	// 获取或创建 hooks 对象
	hooks, ok := data["hooks"].(map[string]interface{})
	if !ok {
		hooks = map[string]interface{}{}
		data["hooks"] = hooks
	}

	// Claude Code 在 Windows 上通过 Git Bash 执行 hook 命令。
	// 旧版本直接写入 os.Executable() 的原始路径（C:\Users\...），
	// bash 会把 \U 等当转义符吃掉，导致 "C:Users...: command not found"。
	// 这里把旧格式（反斜杠 / 正斜杠未加引号）标记为废弃，重写时自动替换。
	upsertHookGroup(hooks, "Stop", "", hookCommand(exePath, "stop"), legacyCommands(exePath, "stop"))
	upsertHookGroup(hooks, "Notification", "permission_prompt", hookCommand(exePath, "permission"), legacyCommands(exePath, "permission"))

	// 确保目录存在
	if dir := filepath.Dir(settingsPath); dir != "" {
		os.MkdirAll(dir, 0755)
	}

	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(settingsPath, out, 0644)
}

// hookCommand 生成 bash 兼容的 hook 命令：反斜杠转正斜杠，路径加双引号。
// 正斜杠路径 C:/Users/... 在 Git Bash 和 cmd 下都能执行；引号兼容带空格的路径。
// 注意不能用 fmt %q：它会把部分非 ASCII 字符转成 Go 转义形式，bash 不认识。
func hookCommand(exePath, sub string) string {
	return "\"" + filepath.ToSlash(exePath) + "\" " + sub
}

// legacyCommands 返回旧版本可能写入的同语义命令（用于清理替换）。
func legacyCommands(exePath, sub string) map[string]bool {
	return map[string]bool{
		exePath + " " + sub:                   true, // 反斜杠原始路径
		filepath.ToSlash(exePath) + " " + sub: true, // 正斜杠但未加引号
	}
}

// upsertHookGroup 确保 hooks[event] 含一个 group 带指定 command：
// 1. 先删除 obsolete 中的旧命令（并清掉因此变空的 group）
// 2. 新命令已存在则不重复添加
// matcher 非空时在新 group 上设置 matcher 字段。
func upsertHookGroup(hooks map[string]interface{}, event, matcher, command string, obsolete map[string]bool) {
	groups, _ := hooks[event].([]interface{})

	// 清理旧命令，保留其他工具和当前新命令
	kept := make([]interface{}, 0, len(groups))
	exists := false
	for _, g := range groups {
		gm, ok := g.(map[string]interface{})
		if !ok {
			kept = append(kept, g)
			continue
		}
		innerHooks, _ := gm["hooks"].([]interface{})
		keptInner := make([]interface{}, 0, len(innerHooks))
		for _, h := range innerHooks {
			hm, ok := h.(map[string]interface{})
			if !ok {
				keptInner = append(keptInner, h)
				continue
			}
			c, _ := hm["command"].(string)
			if obsolete[c] {
				continue // 旧格式，丢弃
			}
			if c == command {
				exists = true
			}
			keptInner = append(keptInner, h)
		}
		if len(keptInner) > 0 {
			gm["hooks"] = keptInner
			kept = append(kept, gm)
		}
	}
	hooks[event] = kept

	if exists {
		return
	}

	newGroup := map[string]interface{}{
		"hooks": []interface{}{
			map[string]interface{}{"type": "command", "command": command},
		},
	}
	if matcher != "" {
		newGroup["matcher"] = matcher
	}
	hooks[event] = append(kept, newGroup)
}
