package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
)

type file struct {
	Url         string
	DeleteUrl   string
	DisplayName template.HTML
	DisplaySize template.HTML
	DisplayTime template.HTML
}

type templateData struct {
	CurrentPath    string
	RootAssetsPath string
	Files          []file
}

func (s *server) getTemplateData(r *http.Request, files []os.FileInfo) templateData {
	data := templateData{
		CurrentPath:    r.URL.Path,
		RootAssetsPath: s.rootAssetsPath,
		Files:          make([]file, 0),
	}

	for _, item := range files {
		name := item.Name()
		size := fileSizeBytes(item.Size()).String()
		// dir has path separator at the end: path/to/dir/
		// file doesn't have separator: path/to/file
		if item.IsDir() {
			name += string(os.PathSeparator)
			size = ""
		}
		_url := r.URL.Path + name

		data.Files = append(data.Files, file{
			Url:         _url,
			DeleteUrl:   "",
			DisplayName: template.HTML(name),
			DisplaySize: template.HTML(size),
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
