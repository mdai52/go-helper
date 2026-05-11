package command

import (
	"os/exec"
	"runtime"

	"github.com/rehiy/libgo/logman"
)

// AllowedShells 白名单：只允许已知安全的 shell
var AllowedShells = map[string]bool{
	"bash": true, "sh": true, "zsh": true,
	"powershell": true, "pwsh": true, "cmd": true,
}

// GetShell 获取可用的 shell
// 如果指定的 shell 不在白名单中或不可用，则返回默认 shell
func GetShell(shell string) string {
	if !AllowedShells[shell] {
		safeInput := shell
		if len(safeInput) > 16 {
			safeInput = safeInput[:16] + "..."
		}
		logman.Warn("GetShell: 拒绝非白名单 shell", "input", safeInput)
		return DefaultShell()
	}
	if _, err := exec.LookPath(shell); err == nil {
		return shell
	}
	return DefaultShell()
}

// DefaultShell 获取当前系统的默认 shell
func DefaultShell() string {
	var candidates []string
	switch runtime.GOOS {
	case "windows":
		candidates = []string{"powershell", "pwsh", "cmd"}
	case "darwin":
		candidates = []string{"zsh", "bash", "sh"}
	default:
		candidates = []string{"bash", "sh", "zsh"}
	}
	for _, fb := range candidates {
		if _, err := exec.LookPath(fb); err == nil {
			return fb
		}
	}
	return "sh"
}
