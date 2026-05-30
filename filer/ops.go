package filer

import (
	"io/fs"
	"os"
	"path/filepath"
)

// ============================================================
// 文件状态判断函数
// ============================================================

// IsDir 判断路径是否为目录
func IsDir(p string) bool {
	info, err := os.Stat(p)
	if err != nil {
		return false
	}

	return info.IsDir()
}

// IsLink 判断路径是否为符号链接
func IsLink(p string) bool {
	info, err := os.Lstat(p)
	if err != nil {
		return false
	}

	return info.Mode()&os.ModeSymlink != 0
}

// Exist 判断文件或目录是否存在
func Exist(p string) bool {
	if _, err := os.Stat(p); err != nil {
		return os.IsExist(err)
	}

	return true
}

// NotExist 判断文件或目录是否不存在
func NotExist(p string) bool {
	if _, err := os.Stat(p); err != nil {
		return os.IsNotExist(err)
	}
	return false
}

// ============================================================
// 文件类型检测函数
// ============================================================

// IsText 判断文件是否为文本文件
// 通过读取文件的前 512 字节，检查是否包含非文本字符（如 null 字节）
func IsText(p string) bool {
	f, err := os.Open(p)
	if err != nil {
		return false
	}
	defer f.Close()

	// 读取前 512 字节进行判断（与 http.DetectContentType 的逻辑类似）
	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && n == 0 {
		return false
	}
	buf = buf[:n]

	// 检查是否包含 null 字节，包含则认为是二进制文件
	for _, b := range buf {
		if b == 0 {
			return false
		}
	}

	// 检查是否包含过多非打印字符
	nonPrintable := 0
	for _, b := range buf {
		// 允许：tab (9), newline (10), carriage return (13), 空格 (32) 及以上
		if b < 9 || (b > 13 && b < 32) || b == 127 {
			nonPrintable++
		}
	}

	// 如果非打印字符超过 10%，认为是二进制文件
	if len(buf) > 0 && float64(nonPrintable)/float64(len(buf)) > 0.1 {
		return false
	}

	return true
}

// ============================================================
// 文件操作函数
// ============================================================

// Glob 根据模式匹配文件路径（如 *.txt, /path/*.log）
// 返回匹配的文件路径列表
func Glob(pattern string) ([]string, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	return matches, nil
}

// Move 移动文件或目录
// 如果目标路径已存在，根据 overwrite 参数决定是否覆盖
func Move(srcPath, dstPath string, overwrite bool) error {
	// 检查目标是否存在
	if info, err := os.Stat(dstPath); err == nil {
		if !overwrite {
			return os.ErrExist
		}
		// 如果目标是目录，且源也是目录，需要特殊处理
		if info.IsDir() {
			// 目标存在且是目录，可以将源移动到目标目录下
			dstPath = filepath.Join(dstPath, filepath.Base(srcPath))
		}
		// 如果目标存在且允许覆盖，先删除目标
		if err := os.RemoveAll(dstPath); err != nil {
			return err
		}
	}

	return os.Rename(srcPath, dstPath)
}

// DirSize 递归计算目录大小（包含所有子目录和文件）
func DirSize(dirPath string) (int64, error) {
	var totalSize int64

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 跳过符号链接，避免循环或重复计算
		if d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			if info.Mode()&os.ModeSymlink != 0 {
				return fs.SkipDir
			}
			return nil
		}

		// 获取文件信息并累加大小
		info, err := d.Info()
		if err != nil {
			return err
		}
		totalSize += info.Size()

		return nil
	})

	if err != nil {
		return 0, err
	}

	return totalSize, nil
}
