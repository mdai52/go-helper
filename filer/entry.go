package filer

import (
	"os"
	"path/filepath"
)

type FileInfo struct {
	Name    string      // 文件名
	Size    int64       // 字节大小
	Mode    os.FileMode // 权限，如 0777
	ModTime int64       // 修改时间，Unix时间戳
	Symlink string      // 链接的真实路径，软链接时有效
	Owner   string      // 所属用户
	Group   string      // 所属组
	IsDir   bool        // 是否是目录
	Data    []byte      // 文件数据
}

// 列出目录中的所有文件
func List(dir string) ([]*FileInfo, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var list []*FileInfo
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			return nil, err
		}
		fp := filepath.Join(dir, file.Name())
		uName, gName, _ := getFileOwner(info)
		list = append(list, &FileInfo{
			Name:    info.Name(),
			Size:    info.Size(),
			Mode:    info.Mode().Perm(),
			ModTime: info.ModTime().Unix(),
			Symlink: Readlink(fp),
			IsDir:   IsDir(fp),
			Owner:   uName,
			Group:   gName,
		})
	}

	return list, nil
}

// 获取文件信息和内容
func Detail(file string, read bool) (*FileInfo, error) {
	info, err := os.Stat(file)
	if err != nil {
		return nil, err
	}

	uName, gName, _ := getFileOwner(info)
	detail := &FileInfo{
		Name:    info.Name(),
		Size:    info.Size(),
		Mode:    info.Mode().Perm(),
		ModTime: info.ModTime().Unix(),
		Symlink: Readlink(file),
		IsDir:   info.IsDir(),
		Owner:   uName,
		Group:   gName,
	}

	if read && !info.IsDir() {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		detail.Data = data
	}

	return detail, nil
}

// 获取文件内容
func Read(file string) ([]byte, error) {
	return os.ReadFile(file)
}

// 追加写入文件内容
func Append(file string, data []byte) error {
	if dir := filepath.Dir(file); !Exists(dir) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}

	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(data)
	return err
}

// 写入文件内容，目录不存在时自动创建
func Write(file string, data []byte) error {
	if dir := filepath.Dir(file); !Exists(dir) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}

	return os.WriteFile(file, data, 0644)
}

// 获取软链接的真实路径
func Readlink(file string) string {
	if IsLink(file) {
		if rp, err := os.Readlink(file); err == nil {
			return rp
		}
	}
	return ""
}
