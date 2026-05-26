package upgrade

import "os"

// BinaryUpdater is responsible for local binary upgrade actions.
type BinaryUpdater struct {
	NewVersion string      // 新版本标识
	OldVersion string      // 旧版本标识
	TargetPath string      // 目标文件路径
	TargetMode os.FileMode // 文件权限
	NewBinary  string      // 新二进制文件路径
	Checksum   []byte      // SHA256 校验和
}

// Init 初始化更新器
func (u *BinaryUpdater) Init() error {
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
		if info, err := os.Stat(u.TargetPath); err == nil {
			u.TargetMode = info.Mode() & os.FileMode(0o777)
		} else {
			u.TargetMode = 0o755
		}
	}

	if u.NewBinary == "" {
		u.NewBinary = u.TargetPath + "-" + u.NewVersion + ".tmp"
	}

	return nil
}

// VerifyChecksum 校验文件完整性
func (u *BinaryUpdater) VerifyChecksum() error {
	return verifyChecksum(u.NewBinary, u.Checksum)
}

// CommitBinary 提交更新
func (u *BinaryUpdater) CommitBinary() error {
	return commitBinary(u.NewBinary, u.TargetPath, u.TargetMode, u.OldVersion)
}
