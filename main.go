package main

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	tmpl := template.Must(template.ParseFiles(
		"templates/layouts/base.html",
		"templates/index.html",
	))
	r.SetHTMLTemplate(tmpl)

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "meetkat",
		})
	})

	r.Run(":8080")
}
