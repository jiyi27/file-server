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
	SharedUrl   string
	IsDir       bool
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
		// init basic info of file entity.
		filename := item.Name()
		size := fileSizeBytes(item.Size()).String()
		filePath := pwd + string(os.PathSeparator) + filename
		canDelete := true
		isDir := false
		var sharedUrl string

		// if file is directory, no shared url, no size info.
		if item.IsDir() {
			isDir = true
			// the full path of dir has a separator "/" at the end: path/to/dir/
			filename += string(os.PathSeparator)
			filePath = pwd + string(os.PathSeparator) + filename
			size = ""
			sharedUrl = ""

			// only empty dir can be deleted.
			empty, err := s.isEmpty(filePath)
			if err != nil {
				log.Printf("failed to check if file %v is emtpy: %v", filePath, err)
				continue
			}
			if !empty {
				canDelete = false
			}
		} else {
			// only files can be shared.
			id, err := s.generateID(10)
			if err != nil {
				log.Printf(err.Error())
			} else {
				s.sharedFiles[id] = filePath
				s.domain = strings.TrimSuffix(s.domain, `/`)
				sharedUrl = fmt.Sprintf("%v%v?shared_id=%v", s.domain, filePath, id)
			}
		}

		data.Files = append(data.Files, file{
			Url:         filePath,
			IsDir:       isDir,
			SharedUrl:   sharedUrl,
			CanDelete:   canDelete,
			DisplayName: template.HTML(filename),
			DisplaySize: template.HTML(size),
			DisplayTime: template.HTML(item.ModTime().Format("02-01-2006")),
		})
	}

	return data
}

func (s *server) generateID(n int) (string, error) {
	for {
		id, err := randomString(n)
		if err != nil {
			return "", fmt.Errorf("failed to generate file ID: %v", err)
		}

		_, ok := s.sharedFiles[id]
		if ok {
			continue
		}

		return id, err
	}
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
