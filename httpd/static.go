package httpd

import (
	"embed"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func Static(root, prefix string) {
	hfs := gin.Dir(root, false)
	engine.NoRoute(StaticServe(hfs, prefix))
}

func StaticIndex(root, prefix string) {
	hfs := gin.Dir(root, true)
	engine.NoRoute(StaticServe(hfs, prefix))
}

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
