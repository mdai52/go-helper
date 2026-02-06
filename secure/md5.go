package secure

import (
	"crypto/md5"
	"encoding/hex"
	"os"
)

// 计算 MD5 哈希
func Md5sum(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// 计算文件 MD5 哈希
func FileMd5sum(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return Md5sum(string(b)), nil
}
