package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"
)

const templateFilePath = "template/index.html"

type file struct {
	Url         string
	DeleteUrl   string
	DisplayName template.HTML
	DisplaySize template.HTML
	DisplayTime template.HTML
}

type templateData struct {
	Title template.HTML
	Files []file
}

func (s *server) getTemplateData(r *http.Request, files []os.FileInfo) templateData {
	data := templateData{
		Title: template.HTML(r.URL.Path),
		Files: make([]file, 0),
	}

	for _, item := range files {
		url_ := r.URL.Path
		name := item.Name()
		if item.IsDir() {
			name += string(os.PathSeparator)
			url_ = path.Join(url_, string(os.PathSeparator))
		}
		data.Files = append(data.Files, file{
			Url:         url_,
			DeleteUrl:   "",
			DisplayName: template.HTML(name),
			DisplaySize: template.HTML(fileSizeBytes(item.Size()).String()),
			DisplayTime: template.HTML(item.ModTime().Format("02-01-2006")),
		})
	}

	return data
}

type fileSizeBytes float64

func (f fileSizeBytes) String() string {
	const (
		KB = 1024.0
		MB = 1024 * KB
		GB = 1024 * MB
	)
	divBy := func(x float64) float64 {
		return float64(f) / x
	}
	switch {
	case f < KB:
		return fmt.Sprintf("%.2fB", f)
	case f < MB:
		return fmt.Sprintf("%.2fK", divBy(KB))
	case f < GB:
		return fmt.Sprintf("%.2fM", divBy(MB))
	default:
		return fmt.Sprintf("%.2fG", divBy(GB))
	}
}
