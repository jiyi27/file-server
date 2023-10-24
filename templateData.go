package main

import (
	"html/template"
	"os"
)

type fileHtml struct {
	Url         string
	DeleteUrl   string
	DisplayName template.HTML
	DisplaySize template.HTML
	DisplayTime template.HTML
}

type templateData struct {
	Path template.HTML

	Files     []os.FileInfo
	FilesHtml []fileHtml
}
