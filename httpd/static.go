package httpd

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func Static(root, prefix string) {
	hfs := gin.Dir(root, false)
	engine.Use(StaticServe(hfs, prefix))
}

func StaticIndex(root, prefix string) {
	hfs := gin.Dir(root, true)
	engine.Use(StaticServe(hfs, prefix))
}

func StaticEmbed(efs embed.FS, prefix, subdir string) {
	var hfs http.FileSystem

	if subdir == "" {
		hfs = http.FS(efs)
	} else {
		sub, _ := fs.Sub(efs, subdir)
		hfs = http.FS(sub)
	}

	engine.Use(StaticServe(hfs, prefix))
}

func StaticServe(hfs http.FileSystem, prefix string) gin.HandlerFunc {
	fileServer := http.FileServer(hfs)
	if prefix != "" {
		fileServer = http.StripPrefix(prefix, fileServer)
	}

	return func(c *gin.Context) {
		relPath := strings.TrimPrefix(c.Request.URL.Path, prefix)
		// 检查是否匹配或是否在根目录下
		if len(relPath) < len(c.Request.URL.Path) || (prefix == "" && c.Request.URL.Path != "") {
			// 清理路径：空路径替换为根目录
			cleanPath := strings.TrimPrefix(relPath, "/")
			if cleanPath == "" {
				cleanPath = "."
			}
			// 检查文件是否存在
			if f, err := hfs.Open(cleanPath); err == nil {
				f.Close()
				fileServer.ServeHTTP(c.Writer, c.Request)
				c.Abort()
			}
		}
	}
}
