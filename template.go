package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

type file struct {
	Url         string
	CanDelete   bool
	DisplayName template.HTML
	DisplaySize template.HTML
	DisplayTime template.HTML
}

type templateData struct {
	IsRoot         bool
	PWD            string // Current path
	ParentDirPath  string
	RootAssetsPath string
	Files          []file
}

func (s *server) getTemplateData(r *http.Request, files []os.FileInfo) templateData {
	pwd := strings.TrimSuffix(r.URL.Path, "/")
	parentDirPath := pwd
	if !(pwd == "") {
		parentDirPath = path.Dir(pwd)
	}

	data := templateData{
		IsRoot:         pwd == "",
		PWD:            pwd,
		ParentDirPath:  parentDirPath,
		RootAssetsPath: s.rootAssetsPath,
		Files:          make([]file, 0),
	}

	for _, item := range files {
		filename := item.Name()
		size := fileSizeBytes(item.Size()).String()
		filePath := pwd + string(os.PathSeparator) + filename
		canDelete := true

		if item.IsDir() {
			// the full path of dir has a separator "/" at the end: path/to/dir/
			filename += string(os.PathSeparator)
			filePath = pwd + string(os.PathSeparator) + filename
			size = ""

			// only empty dir can be deleted.
			empty, err := s.isEmpty(filePath)
			if err != nil {
				log.Printf("failed to check if file %v is emtpy: %v", filePath, err)
				continue
			}
			if !empty {
				canDelete = false
			}
		}

		data.Files = append(data.Files, file{
			Url:         filePath,
			CanDelete:   canDelete,
			DisplayName: template.HTML(filename),
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
