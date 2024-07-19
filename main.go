package main

import (
	"io"
	"net/http"
	"slices"
	"strconv"
	"text/template"

	"github.com/labstack/echo/v4"
)

type Task struct {
	Id    string
	Title string
}

type Template struct {
	Templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.Templates.ExecuteTemplate(w, name, data)
}

func NewTemplateRenderer(e *echo.Echo, paths ...string) {
	tmpl := &template.Template{}
	for i := range paths {
		template.Must(tmpl.ParseGlob(paths[i]))
	}
	t := newTemplate(tmpl)
	e.Renderer = t
}

func newTemplate(templates *template.Template) echo.Renderer {
	return &Template{
		Templates: templates,
	}
}

var ntasks []Task
var ftasks []Task

func main() {
	e := echo.New()
	NewTemplateRenderer(e, "public/*.html")
	e.GET("/", func(c echo.Context) error {
		res := map[string]interface{}{
			"ntasks": ntasks,
			"ftasks": ftasks,
		}
		return c.Render(http.StatusOK, "index", res)
	})
	e.POST("/add", func(c echo.Context) error {
		title := c.FormValue("title")
		ntasks = append([]Task{Task{strconv.Itoa(len(ntasks) + len(ftasks)), title}}, ntasks...)
		res := map[string]interface{}{
			"ntasks": ntasks,
			"ftasks": ftasks,
		}
		return c.Render(http.StatusOK, "index", res)
	})
	e.GET("/delete/:id", func(c echo.Context) error {
		id := c.Param("id")
		for i, v := range ntasks {
			if v.Id == id {
				ftasks = append([]Task{ntasks[i]}, ftasks...)
				ntasks = slices.Delete(ntasks, i, i+1)
			}
		}
		res := map[string]interface{}{
			"ntasks": ntasks,
			"ftasks": ftasks,
		}
		return c.Render(http.StatusOK, "index", res)
	})
	e.Logger.Fatal(e.Start(":3333"))
}
