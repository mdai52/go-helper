package secure

import (
	"crypto/md5"
	"encoding/hex"
	"os"
)

// MD5sum 计算 MD5 哈希
func MD5sum(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// FileMD5sum 计算文件 MD5 哈希
func FileMD5sum(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return MD5sum(string(b)), nil
}
