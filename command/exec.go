package command

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/rehiy/libgo/logman"
)

// ScriptPayload 脚本执行参数
type ScriptPayload struct {
	Name       string `json:"name" note:"脚本名称"`
	ScriptType string `json:"scriptType" note:"脚本类型 BAT|POWERSHELL|SHELL|EXEC"`
	Username   string `json:"username" note:"执行脚本的用户名"`
	WorkDir    string `json:"workDir" note:"脚本工作目录"`
	Content    string `json:"content" note:"脚本内容"`
	Timeout    uint   `json:"timeout" note:"超时时间（秒）"`
}

// RunScript 执行脚本并返回输出
func RunScript(data *ScriptPayload) (string, error) {
	var (
		err error
		tmp string
		bin string
		arg []string
	)

	switch data.ScriptType {
	case "BAT":
		tmp, err = createTempScript(data.Content, ".bat")
		arg = []string{"/c", "CALL", tmp}
		bin = GetShell("cmd")
	case "POWERSHELL":
		tmp, err = createTempScript(data.Content, ".ps1")
		arg = []string{"-File", tmp}
		bin = GetShell("powershell")
	case "SHELL":
		if strings.HasPrefix(data.Content, "#!/") {
			tmp, err = createTempScript(data.Content, "")
			arg = []string{}
			bin = tmp
		} else {
			arg = []string{"-c", data.Content}
			bin = DefaultShell()
			tmp = "-"
		}
	case "EXEC":
		arg = strings.Fields(data.Content)
		bin, arg = arg[0], arg[1:]
		tmp = "-"
	default:
		err = fmt.Errorf("unsupported script type: %s", data.ScriptType)
	}

	if err != nil || tmp == "" {
		return "", err
	}

	if tmp != "-" {
		defer os.Remove(tmp)
	}

	workDir := data.WorkDir
	if workDir == "" {
		workDir = filepath.Dir(bin)
	}

	return RunCommand(bin, arg, workDir, data.Timeout)
}

// NewCommand 创建命令对象
func NewCommand(ctx context.Context, bin string, arg []string, workDir string) *exec.Cmd {
	var cmd *exec.Cmd
	env := os.Environ()

	switch runtime.GOOS {
	case "windows":
		if len(arg) > 0 {
			fullCmd := bin + " " + quoteArgs(arg)
			cmd = exec.CommandContext(ctx, "cmd", "/C", "chcp 65001 >nul && "+fullCmd)
		} else {
			cmd = exec.CommandContext(ctx, "cmd", "/C", "chcp 65001 >nul && "+bin)
		}
		env = append(env, "TERM=dumb")
	case "darwin":
		cmd = exec.CommandContext(ctx, bin, arg...)
		env = append([]string{"TERM=xterm-256color", "CLICOLOR=1"}, env...)
	default:
		cmd = exec.CommandContext(ctx, bin, arg...)
		env = append([]string{"TERM=xterm-256color"}, env...)
	}

	cmd.Dir = workDir
	cmd.Env = env
	return cmd
}

// RunCommand 执行命令并返回输出
func RunCommand(bin string, arg []string, workDir string, timeout uint) (string, error) {
	logman.Debug("执行命令", "bin", bin, "arg", arg)

	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()
	}

	cmd := NewCommand(ctx, bin, arg, workDir)
	ret, err := cmd.CombinedOutput()
	return string(ret), err
}

// createTempScript 创建临时脚本文件
func createTempScript(code string, ext string) (string, error) {
	tf, err := os.CreateTemp("", "tmp-*"+ext)
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer tf.Close()

	code = fixLineEnding(code)

	if _, err = tf.WriteString(code); err != nil {
		return "", fmt.Errorf("write temp file: %w", err)
	}

	if runtime.GOOS != "windows" {
		tf.Chmod(0755)
	}

	return tf.Name(), nil
}

// fixLineEnding 修复换行符以匹配当前系统
func fixLineEnding(code string) string {
	switch {
	case strings.Contains(code, "\r\n"):
		if runtime.GOOS == "windows" {
			return code
		}
		return strings.ReplaceAll(code, "\r\n", "\n")
	case strings.Contains(code, "\n"):
		if runtime.GOOS != "windows" {
			return code
		}
		return strings.ReplaceAll(code, "\n", "\r\n")
	case strings.Contains(code, "\r"):
		if runtime.GOOS != "windows" {
			return strings.ReplaceAll(code, "\r", "\n")
		}
		return strings.ReplaceAll(code, "\r", "\r\n")
	default:
		return code
	}
}

// quoteArgs 为包含空格的参数添加引号
func quoteArgs(arg []string) string {
	var sb strings.Builder
	for i, a := range arg {
		if i > 0 {
			sb.WriteByte(' ')
		}
		if strings.ContainsAny(a, " \t") {
			sb.WriteByte('"')
			sb.WriteString(a)
			sb.WriteByte('"')
		} else {
			sb.WriteString(a)
		}
	}
	return sb.String()
}
