package main

import (
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"text/template"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/r3labs/sse/v2"
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

	server := sse.New()
	server.AutoReplay = false
	_ = server.CreateStream("tasks")

	go func(s *sse.Server) {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		a := ""
		for {
			ndata := ""
			fdata := ""
			select {
			case <-ticker.C:
				for i := range ftasks {
					fdata += fmt.Sprintf("<a id=\"deadtask\">%v</a><br>", ftasks[i].Title)
				}
				for i := range ntasks {
					ndata += fmt.Sprintf("<a id=\"livetask\" hx-get=\"/delete/%v\" hx-swap=\"none\">%v</a><br>", ntasks[i].Id, ntasks[i].Title)
				}
				data := "<div>" + ndata + fdata + "</div>"

				if a != data {
					s.Publish("tasks", &sse.Event{
						Event: []byte("tasks"),
						Data:  []byte(data),
					})
					a = data
					fmt.Println("Refresh")
				}
			}
		}
	}(server)

	NewTemplateRenderer(e, "public/*.html")
	e.GET("/", func(c echo.Context) error {
		res := map[string]interface{}{
			"ntasks": ntasks,
			"ftasks": ftasks,
		}
		return c.Render(http.StatusOK, "index", res)
	})
	e.GET("/sse", func(c echo.Context) error {
		go func() {
			<-c.Request().Context().Done()
			return
		}()

		server.ServeHTTP(c.Response(), c.Request())
		return nil
	})
	e.POST("/add", func(c echo.Context) error {
		title := c.FormValue("title")
		ntasks = append([]Task{{strconv.Itoa(len(ntasks) + len(ftasks)), title}}, ntasks...)
		return c.NoContent(http.StatusOK)
	})
	e.GET("/delete/:id", func(c echo.Context) error {
		id := c.Param("id")
		for i, v := range ntasks {
			if v.Id == id {
				ftasks = append([]Task{ntasks[i]}, ftasks...)
				ntasks = slices.Delete(ntasks, i, i+1)
			}
		}
		return c.NoContent(http.StatusOK)
	})
	e.Logger.Fatal(e.Start(":3333"))
}
