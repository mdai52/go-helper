package upgrade

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"io"
	"os"
)

func verifyChecksum(filePath string, checksum []byte) error {
	if len(checksum) == 0 {
		return nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	if !bytes.Equal(checksum, hash.Sum(nil)) {
		return errors.New("updated file has wrong checksum")
	}

	return nil
}

func commitBinary(newBinary, targetPath string, targetMode os.FileMode, oldVersion string) error {
	// 校验文件
	if err := verifyChecksum(newBinary, nil); err != nil {
		return err
	}

	// 设置权限
	if err := os.Chmod(newBinary, targetMode); err != nil {
		return err
	}

	// 备份旧文件
	originFile := targetPath + "-" + oldVersion
	if err := os.Rename(targetPath, originFile); err != nil {
		return err
	}

	// 替换文件
	if err := os.Rename(newBinary, targetPath); err != nil {
		// 尝试回滚
		if er2 := os.Rename(originFile, targetPath); er2 != nil {
			return &ErrRollback{err, er2}
		}
		return err
	}

	// 删除备份
	os.Remove(originFile)

	return nil
}