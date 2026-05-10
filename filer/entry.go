package filer

import (
	"os"
	"path/filepath"
)

type FileInfo struct {
	Name    string      `note:"文件名" json:"name"`
	Size    int64       `note:"字节大小" json:"size"`
	Mode    os.FileMode `note:"权限，如 0777" json:"mode"`
	ModTime int64       `note:"修改时间，Unix时间戳" json:"modTime"`
	Symlink string      `note:"链接的真实路径，软链接时有效" json:"symlink"`
	Owner   string      `note:"所属用户" json:"owner"`
	Group   string      `note:"所属组" json:"group"`
	IsDir   bool        `note:"是否是目录" json:"isDir"`
	Data    []byte      `note:"文件数据" json:"data,omitempty"`
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
func Info(file string, read bool) (*FileInfo, error) {
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

// 获取文件内容
func ReadText(file string) (string, error) {
	bytes, err := os.ReadFile(file)
	return string(bytes), err
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

// 追加写入文件内容
func Append(file string, data []byte) error {
	if dir := filepath.Dir(file); NotExist(dir) {
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
	if dir := filepath.Dir(file); NotExist(dir) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}

	return os.WriteFile(file, data, 0644)
}
