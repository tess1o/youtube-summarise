package handlers

import (
	"embed"
	"github.com/labstack/echo/v4"
	"html/template"
	"io"
	"net/http"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func NewIndexTemplate(fs embed.FS) *Template {
	return &Template{
		templates: template.Must(template.ParseFS(fs, "templates/index.html")),
	}
}

func NewIndexHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", map[string]interface{}{})
}
