package assets

import (
	"embed"
	"net/http"
)

// "GET /":                  controller.Index,
// "GET /banner.png":        controller.Banner,
// "GET /favicon.ico":       controller.Favicon,
// "GET /robots.txt":        controller.Robots,

//go:embed home.html banner.png logo.png robots.txt

var fs embed.FS

var Assets AssetFS

type AssetFS struct {
	embed.FS
	path string
}

func init() {
	Assets = AssetFS{
		FS:   fs,
		path: ".",
	}
}

// Open 实现http.fs 接口
// func (a AssetFS) Open(name string) (stdfs.File, error) {
// 	if filepath.Separator != '/' && strings.ContainsRune(name, filepath.Separator) {
// 		return nil, errors.New("http: invalid character in file path")
// 	}
// 	fullName := filepath.Join(a.path, filepath.FromSlash(path.Clean("/"+name)))
// 	file, err := a.FS.Open(fullName)
// 	return file, err
// }

// HTTPFileSystem http.fs 文件系统
func (a AssetFS) HTTPFileSystem() http.FileSystem {
	return http.FS(a)
}
