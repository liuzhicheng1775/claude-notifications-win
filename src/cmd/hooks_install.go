package cmd

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

// installHooks 将 notify.exe 的 Stop / Notification hooks 合并进 settings.json。
// 已存在相同命令则不重复添加；保留所有现有字段（env、model、statusLine 等）。
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

	addHookGroup(hooks, "Stop", "", exePath+" stop")
	addHookGroup(hooks, "Notification", "permission_prompt", exePath+" permission")

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

// addHookGroup 确保 hooks[event] 含一个 group 带指定 command，不重复添加。
// matcher 非空时在新 group 上设置 matcher 字段。
func addHookGroup(hooks map[string]interface{}, event, matcher, command string) {
	groups, _ := hooks[event].([]interface{})

	// 检查是否已存在相同 command
	for _, g := range groups {
		gm, ok := g.(map[string]interface{})
		if !ok {
			continue
		}
		innerHooks, _ := gm["hooks"].([]interface{})
		for _, h := range innerHooks {
			hm, ok := h.(map[string]interface{})
			if !ok {
				continue
			}
			if c, _ := hm["command"].(string); c == command {
				return // 已存在，不重复
			}
		}
	}

	newGroup := map[string]interface{}{
		"hooks": []interface{}{
			map[string]interface{}{"type": "command", "command": command},
		},
	}
	if matcher != "" {
		newGroup["matcher"] = matcher
	}
	hooks[event] = append(groups, newGroup)
}
