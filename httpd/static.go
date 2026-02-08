package httpd

import (
	"embed"
	"io/fs"
	"net/http"

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
		fileHandler.ServeHTTP(c.Writer, c.Request)
	}
}
