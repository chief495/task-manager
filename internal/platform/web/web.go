package web

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

func LoadTemplates() *template.Template {
	templ := template.New("")
	pattern := filepath.Join(WebDir(), "templates", "*.html")
	templ, err := templ.ParseGlob(pattern)
	if err != nil {
		panic(err)
	}
	return templ
}

func WebDir() string {
	webDir := os.Getenv("WEB_DIR")
	if webDir == "" {
		return "./web"
	}
	return webDir
}

func StaticDir() string {
	return filepath.Join(WebDir(), "static")
}

func AssetURL(urlPath string) string {
	webDir := WebDir()

	rel := strings.TrimPrefix(urlPath, "/static/")
	fsPath := filepath.Join(webDir, "static", filepath.FromSlash(rel))

	info, err := os.Stat(fsPath)
	if err != nil {
		return urlPath
	}

	return fmt.Sprintf("%s?v=%d", urlPath, info.ModTime().Unix())
}
