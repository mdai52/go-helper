package webssh

import (
	"fmt"
	"io"
	"os"
	"path"
	"sync"
	"time"

	"github.com/pkg/sftp"
)

// FileInfo SFTP 文件/目录信息
type FileInfo struct {
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	Mode       string `json:"mode"`
	ModTime    int64  `json:"modTime"`
	IsDir      bool   `json:"isDir"`
	IsLink     bool   `json:"isLink"`
	LinkTarget string `json:"linkTarget,omitempty"`
}

// ListResult 目录列表结果（含实际路径）
type ListResult struct {
	Path  string      `json:"path"`
	Files []*FileInfo `json:"files"`
}

// sftpEntry 连接池中的一个缓存条目
type sftpEntry struct {
	conn     *sftp.Client
	lastUsed time.Time
}

// SFTPClient SFTP 客户端，内置连接池，按 key 复用连接
type SFTPClient struct {
	mu          sync.Mutex
	entries     map[string]*sftpEntry
	idleTimeout time.Duration
	stopCh      chan struct{}
}

// NewSFTPClient 创建 SFTP 客户端，idleTimeout 为 0 时使用默认值 30s
func NewSFTPClient(idleTimeout time.Duration) *SFTPClient {
	if idleTimeout <= 0 {
		idleTimeout = 30 * time.Second
	}
	c := &SFTPClient{
		entries:     make(map[string]*sftpEntry),
		idleTimeout: idleTimeout,
		stopCh:      make(chan struct{}),
	}
	go c.evictLoop()
	return c
}

// Close 关闭客户端，释放所有连接
func (c *SFTPClient) Close() {
	close(c.stopCh)
	// 先在锁内取出所有连接，再在锁外关闭，避免持锁期间阻塞
	c.mu.Lock()
	conns := make([]*sftp.Client, 0, len(c.entries))
	for _, e := range c.entries {
		conns = append(conns, e.conn)
	}
	c.entries = make(map[string]*sftpEntry)
	c.mu.Unlock()
	for _, conn := range conns {
		conn.Close()
	}
}

// List 列出目录内容，dirPath 为空或 "~" 时使用远程 home 目录
func (c *SFTPClient) List(key string, opt *SSHClientOption, dirPath string) (*ListResult, error) {
	conn, err := c.getConn(key, opt)
	if err != nil {
		return nil, err
	}

	if dirPath == "" || dirPath == "~" {
		if wd, err := conn.Getwd(); err == nil {
			dirPath = wd
		} else {
			dirPath = "/"
		}
	}

	entries, err := conn.ReadDir(dirPath)
	if err != nil {
		c.invalidate(key)
		return nil, fmt.Errorf("读取目录失败: %w", err)
	}

	files := make([]*FileInfo, 0, len(entries))
	for _, e := range entries {
		isLink := e.Mode()&os.ModeSymlink != 0
		isDir := e.IsDir()
		var linkTarget string
		if isLink {
			fullPath := path.Join(dirPath, e.Name())
			if info, err := conn.Stat(fullPath); err == nil {
				isDir = info.IsDir()
			} else {
				// Stat 失败（如悬空链接），明确置为 false
				isDir = false
			}
			if target, err := conn.ReadLink(fullPath); err == nil {
				linkTarget = target
			}
		}
		files = append(files, &FileInfo{
			Name:       e.Name(),
			Size:       e.Size(),
			Mode:       e.Mode().String(),
			ModTime:    e.ModTime().Unix(),
			IsDir:      isDir,
			IsLink:     isLink,
			LinkTarget: linkTarget,
		})
	}
	return &ListResult{Path: dirPath, Files: files}, nil
}

// Download 打开远程文件用于下载，返回 ReadCloser 和文件大小。
// 下载使用独占连接（流式传输期间不能被复用），关闭 ReadCloser 时连接一并释放。
func (c *SFTPClient) Download(opt *SSHClientOption, filePath string) (io.ReadCloser, int64, error) {
	conn, err := c.dial(opt)
	if err != nil {
		return nil, 0, err
	}

	stat, err := conn.Stat(filePath)
	if err != nil {
		conn.Close()
		return nil, 0, fmt.Errorf("文件不存在: %w", err)
	}
	if stat.IsDir() {
		conn.Close()
		return nil, 0, fmt.Errorf("不能下载目录")
	}

	f, err := conn.Open(filePath)
	if err != nil {
		conn.Close()
		return nil, 0, fmt.Errorf("打开文件失败: %w", err)
	}

	return &downloadCloser{ReadCloser: f, conn: conn}, stat.Size(), nil
}

// Upload 将 src 上传到远程 destPath（覆盖）
func (c *SFTPClient) Upload(key string, opt *SSHClientOption, destPath string, src io.Reader) error {
	conn, err := c.getConn(key, opt)
	if err != nil {
		return err
	}

	f, err := conn.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		// OpenFile 失败（如权限不足）不代表连接断开，不 invalidate
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, src); err != nil {
		// 写入失败（如磁盘满）不代表连接断开，不 invalidate
		return fmt.Errorf("写入文件失败: %w", err)
	}
	return nil
}

// Remove 删除文件或目录；recursive=true 时递归删除目录
func (c *SFTPClient) Remove(key string, opt *SSHClientOption, targetPath string, recursive bool) error {
	conn, err := c.getConn(key, opt)
	if err != nil {
		return err
	}

	stat, err := conn.Stat(targetPath)
	if err != nil {
		return fmt.Errorf("路径不存在: %w", err)
	}

	if stat.IsDir() {
		if recursive {
			if err := removeAll(conn, targetPath); err != nil {
				c.invalidate(key)
				return fmt.Errorf("递归删除目录失败: %w", err)
			}
		} else {
			if err := conn.RemoveDirectory(targetPath); err != nil {
				return fmt.Errorf("删除目录失败（目录非空时请使用递归删除）: %w", err)
			}
		}
	} else {
		if err := conn.Remove(targetPath); err != nil {
			c.invalidate(key)
			return fmt.Errorf("删除文件失败: %w", err)
		}
	}
	return nil
}

// Mkdir 递归创建目录
func (c *SFTPClient) Mkdir(key string, opt *SSHClientOption, dirPath string) error {
	conn, err := c.getConn(key, opt)
	if err != nil {
		return err
	}
	if err := conn.MkdirAll(dirPath); err != nil {
		c.invalidate(key)
		return fmt.Errorf("创建目录失败: %w", err)
	}
	return nil
}

// Rename 重命名/移动文件或目录
func (c *SFTPClient) Rename(key string, opt *SSHClientOption, oldPath, newPath string) error {
	conn, err := c.getConn(key, opt)
	if err != nil {
		return err
	}
	if err := conn.Rename(oldPath, newPath); err != nil {
		c.invalidate(key)
		return fmt.Errorf("重命名失败: %w", err)
	}
	return nil
}

// ── 内部方法 ────────────────────────────────────────────────────────────────

// dial 建立一个新的独占 SSH+SFTP 连接
func (c *SFTPClient) dial(opt *SSHClientOption) (*sftp.Client, error) {
	sshClient, err := NewSSHClient(opt)
	if err != nil {
		return nil, fmt.Errorf("SSH 连接失败: %w", err)
	}
	conn, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("SFTP 初始化失败: %w", err)
	}
	return conn, nil
}

// getConn 从连接池获取连接，不存在或已失效则新建
// 使用 CAS 风格：先取出旧连接探活，失败后用 dial 建新连接，
// 最后在锁内做 check-then-set，避免并发时 double-close 和连接泄漏。
func (c *SFTPClient) getConn(key string, opt *SSHClientOption) (*sftp.Client, error) {
	// 1. 取出当前缓存的连接（快速路径）
	c.mu.Lock()
	e, ok := c.entries[key]
	c.mu.Unlock()

	if ok {
		if _, err := e.conn.Getwd(); err == nil {
			c.mu.Lock()
			// 再次确认 entry 未被其他 goroutine 替换
			if cur, still := c.entries[key]; still && cur == e {
				e.lastUsed = time.Now()
			}
			c.mu.Unlock()
			return e.conn, nil
		}
		// 探活失败：在锁内确认仍是同一个 entry 再删除，防止重复 close
		c.mu.Lock()
		if cur, still := c.entries[key]; still && cur == e {
			delete(c.entries, key)
			c.mu.Unlock()
			e.conn.Close() // 锁外关闭，避免阻塞
		} else {
			c.mu.Unlock()
		}
	}

	// 2. 建立新连接
	conn, err := c.dial(opt)
	if err != nil {
		return nil, err
	}

	// 3. 写入连接池：若并发时已有其他 goroutine 写入，关闭多余连接
	c.mu.Lock()
	if existing, ok := c.entries[key]; ok {
		c.mu.Unlock()
		conn.Close() // 丢弃多余连接
		return existing.conn, nil
	}
	c.entries[key] = &sftpEntry{conn: conn, lastUsed: time.Now()}
	c.mu.Unlock()
	return conn, nil
}

// invalidate 主动移除某个 key 的缓存（连接出错时调用）
func (c *SFTPClient) invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if e, ok := c.entries[key]; ok {
		e.conn.Close()
		delete(c.entries, key)
	}
}

// evictLoop 定期清理空闲超时的连接
func (c *SFTPClient) evictLoop() {
	ticker := time.NewTicker(c.idleTimeout / 3)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// 先在锁内收集过期连接并从 map 删除，再在锁外关闭，避免持锁期间阻塞
			var expired []*sftp.Client
			c.mu.Lock()
			for key, e := range c.entries {
				if time.Since(e.lastUsed) > c.idleTimeout {
					expired = append(expired, e.conn)
					delete(c.entries, key)
				}
			}
			c.mu.Unlock()
			for _, conn := range expired {
				conn.Close()
			}
		case <-c.stopCh:
			return
		}
	}
}

// downloadCloser 包装 sftp.File，Close 时同时关闭底层独占连接
type downloadCloser struct {
	io.ReadCloser
	conn *sftp.Client
}

func (d *downloadCloser) Close() error {
	fileErr := d.ReadCloser.Close()
	connErr := d.conn.Close()
	if fileErr != nil {
		return fileErr
	}
	return connErr
}

// removeAll 递归删除目录及其所有内容（类似 rm -rf）
func removeAll(conn *sftp.Client, dirPath string) error {
	entries, err := conn.ReadDir(dirPath)
	if err != nil {
		return err
	}
	for _, e := range entries {
		childPath := path.Join(dirPath, e.Name())
		if e.IsDir() {
			if err := removeAll(conn, childPath); err != nil {
				return err
			}
		} else {
			if err := conn.Remove(childPath); err != nil {
				return err
			}
		}
	}
	return conn.RemoveDirectory(dirPath)
}
