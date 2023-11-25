package main

import (
	"bytes"
	_ "embed"
	"html/template"
	"io"
	"net/http"
	"time"
)

//go:embed frontend/index.html
var defaultTplStr string

//go:embed frontend/favicon.ico
var defaultFavicon []byte

//go:embed frontend/index.css
var defaultCss []byte

var modifiedTime = time.Now()

type Theme struct {
	Template *template.Template
	Assets   Assets
}

func (t *Theme) RenderPage(w io.Writer, data interface{}) error {
	return t.Template.Execute(w, data)
}

func (t *Theme) RenderAsset(w http.ResponseWriter, r *http.Request, assetPath string) {
	asset, ok := t.Assets[assetPath]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", asset.ContentType)
	http.ServeContent(w, r, assetPath, modifiedTime, asset.ReadSeeker)
}

func newTheme() *Theme {
	tpl := template.New("page")
	tpl, err := tpl.Parse(defaultTplStr)
	if err != nil {
		tpl, _ = tpl.Parse("Builtin Template Error")
	}

	assets := Assets{
		"favicon.ico": {"image/x-icon", bytes.NewReader(defaultFavicon)},
		"index.css":   {"text/css; charset=utf-8", bytes.NewReader(defaultCss)},
	}

	return &Theme{
		Template: tpl,
		Assets:   assets,
	}
}

type Asset struct {
	ContentType string
	ReadSeeker  io.ReadSeeker
}

type Assets map[string]Asset
