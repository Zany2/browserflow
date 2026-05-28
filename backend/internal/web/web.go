package web

import (
	"bytes"
	"embed"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
)

//go:embed dist
var embedded embed.FS

// BindFrontend registers embedded frontend assets on the HTTP server.
func BindFrontend(s *ghttp.Server) {
	dist, err := fs.Sub(embedded, "dist")
	if err != nil {
		return
	}

	handler := func(r *ghttp.Request) {
		serveFrontend(r, dist)
	}
	s.BindHandler("/", handler)
	s.BindHandler("/*any", handler)
}

func serveFrontend(r *ghttp.Request, dist fs.FS) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		r.Response.WriteStatus(http.StatusMethodNotAllowed)
		return
	}

	name := cleanAssetPath(r.URL.Path)
	if name == "" {
		serveEmbeddedFile(r, dist, "index.html")
		return
	}
	if strings.HasPrefix(name, "api/") {
		r.Response.WriteStatus(http.StatusNotFound)
		return
	}

	info, err := fs.Stat(dist, name)
	if err == nil && !info.IsDir() {
		serveEmbeddedFile(r, dist, name)
		return
	}
	if path.Ext(name) != "" || name == "assets" || strings.HasPrefix(name, "assets/") {
		r.Response.WriteStatus(http.StatusNotFound)
		return
	}
	serveEmbeddedFile(r, dist, "index.html")
}

func cleanAssetPath(urlPath string) string {
	name := strings.TrimPrefix(path.Clean("/"+urlPath), "/")
	if name == "." {
		return ""
	}
	return name
}

func serveEmbeddedFile(r *ghttp.Request, dist fs.FS, name string) {
	data, err := fs.ReadFile(dist, name)
	if err != nil {
		r.Response.WriteStatus(http.StatusNotFound)
		return
	}
	if contentType := mime.TypeByExtension(path.Ext(name)); contentType != "" {
		r.Response.Header().Set("Content-Type", contentType)
	}
	http.ServeContent(r.Response.RawWriter(), r.Request, path.Base(name), time.Time{}, bytes.NewReader(data))
}
