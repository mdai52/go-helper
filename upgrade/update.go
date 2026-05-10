package upgrade

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	"github.com/rehiy/libgo/logman"
	"github.com/rehiy/libgo/request"
)

// UpdateParam 更新参数
type UpdateParam struct {
	Server  string `note:"更新服务器"`
	Version string `note:"当前版本"`
}

// UpdateInfo 更新信息
type UpdateInfo struct {
	Type    string `note:"更新方式"`
	Error   string `note:"错误信息"`
	Message string `note:"提示信息"`
	Release string `note:"更新说明"`
	Version string `note:"最新版本"`
	Package string `note:"下载地址"`
}

// ErrRollback 回滚错误
type ErrRollback struct {
	error          // original error
	Rollback error // error encountered while rolling back
}

// ErrNoUpdate 无需更新
var ErrNoUpdate = errors.New("no update")

// Updater 二进制更新器
type Updater struct {
	NewVersion string        // 新版本标识
	OldVersion string        // 旧版本标识
	TargetPath string        // 目标文件路径
	TargetMode os.FileMode   // 文件权限
	NewBinary  string        // 新二进制文件路径
	Checksum   []byte        // SHA256 校验和
}

// Init 初始化更新器
func (u *Updater) Init() error {
	if u.NewVersion == "" {
		u.NewVersion = "new"
	}

	if u.OldVersion == "" {
		u.OldVersion = "old"
	}

	if u.TargetPath == "" {
		p, err := os.Executable()
		if err != nil {
			return err
		}
		u.TargetPath = p
	}

	if u.TargetMode == 0 {
		u.TargetMode = 0755
	}

	u.NewBinary = u.TargetPath + "-" + u.NewVersion

	return nil
}

// VerifyChecksum 校验文件完整性
func (u *Updater) VerifyChecksum() error {
	return verifyChecksum(u.NewBinary, u.Checksum)
}

// CommitBinary 提交更新
func (u *Updater) CommitBinary() error {
	return commitBinary(u.NewBinary, u.TargetPath, u.TargetMode, u.OldVersion)
}

// ApplyUpdate 执行更新
func ApplyUpdate(param *UpdateParam) error {
	logger := logman.Named("upgrade")

	logger.Info(
		"checking update",
		"version", param.Version,
		"url", param.Server,
	)

	// 检查更新
	info, err := CheckUpdate(param)
	if err != nil {
		logger.Error("check update failed", "error", err)
		return err
	}

	if !strings.HasPrefix(info.Package, "https://") {
		logger.Info("no need to update", "info", info)
		return ErrNoUpdate
	}

	// 初始化更新器
	updater := &Updater{OldVersion: param.Version, NewVersion: info.Version}
	if err := updater.Init(); err != nil {
		logger.Error("init updater failed", "error", err)
		return err
	}

	// 下载二进制
	_, err = request.Download(info.Package, updater.NewBinary, true)
	if err != nil {
		logger.Error("download binary failed", "error", err)
		return err
	}

	// 应用更新
	if err = updater.CommitBinary(); err != nil {
		logger.Error("apply binary failed", "error", err)
		if _, ok := err.(*ErrRollback); ok {
			logger.Error("failed to rollback from bad update")
		}
		return err
	}

	return nil
}

// CheckUpdate 检查更新
func CheckUpdate(param *UpdateParam) (*UpdateInfo, error) {
	info := &UpdateInfo{}

	url := param.Server
	url += "?ver=" + param.Version
	url += "&os=" + runtime.GOOS
	url += "&arch=" + runtime.GOARCH

	body, err := request.Get(url, request.H{})
	if err != nil {
		return info, err
	}

	err = json.Unmarshal(body, &info)
	if err != nil {
		return info, err
	}

	if info.Error != "" {
		return info, errors.New(info.Error)
	}
	if info.Package == "" {
		return info, errors.New("get package url failed")
	}

	return info, nil
}

// Restart 重启应用
func Restart() error {
	self, err := os.Executable()
	if err != nil {
		return err
	}

	args, env := os.Args, os.Environ()

	// Windows 不支持 exec syscall
	if runtime.GOOS == "windows" {
		cmd := exec.Command(self, args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Env = env
		if err = cmd.Start(); err != nil {
			return err
		}
		os.Exit(0)
	}

	// 其他系统
	return syscall.Exec(self, args, env)
}