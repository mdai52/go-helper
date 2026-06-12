package httpd

import (
	"embed"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// Static 静态文件服务
func Static(root, prefix string) {
	hfs := gin.Dir(root, false)
	engine.NoRoute(StaticServe(hfs, prefix))
}

// StaticIndex 静态文件服务（带目录列表）
func StaticIndex(root, prefix string) {
	hfs := gin.Dir(root, true)
	engine.NoRoute(StaticServe(hfs, prefix))
}

// StaticEmbed 嵌入式静态文件服务
func StaticEmbed(efs embed.FS, prefix, subdir string) {
	var hfs http.FileSystem

	if subdir == "" {
		hfs = http.FS(efs)
	} else {
		sub, _ := fs.Sub(efs, subdir)
		hfs = http.FS(sub)
	}

	engine.NoRoute(StaticServe(hfs, prefix))
}

// StaticServe 静态文件服务处理器
func StaticServe(hfs http.FileSystem, prefix string) gin.HandlerFunc {
	fileHandler := http.FileServer(hfs)
	if prefix != "" {
		fileHandler = http.StripPrefix(prefix, fileHandler)
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		f, err := hfs.Open(path)
		if err == nil {
			f.Close()
			fileHandler.ServeHTTP(c.Writer, c.Request)
			return
		}

		// http.FS(embed.FS) 无法直接 Open 目录路径（如 /openapi/），
		// 需通过探测目录下的 index.html 来确认该目录存在
		if strings.HasSuffix(path, "/") {
			if fi, idxErr := hfs.Open(path + "index.html"); idxErr == nil {
				fi.Close()
				fileHandler.ServeHTTP(c.Writer, c.Request)
				return
			}
		}

		// 静态资源文件不存在时返回 404
		if filepath.Ext(path) != "" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		// SPA 路由：返回 index.html
		c.Request.URL.Path = "/"
		fileHandler.ServeHTTP(c.Writer, c.Request)
	}
}
