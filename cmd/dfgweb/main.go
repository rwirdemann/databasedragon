package main

import (
	"embed"
	"github.com/rwirdemann/datafrog/web/app"
	"github.com/rwirdemann/simpleweb"
	"log"
)

// Expects all HTML templates in datafrog/cmd/dfgweb/templates
//
//go:embed all:templates
var templates embed.FS

func init() {
	simpleweb.Init(templates, []string{"templates/layout.html"}, app.Conf.Web.Port)
}

func main() {
	app.RegisterHandler()
	simpleweb.ShowRoutes()
	log.Printf("Connecting backend: http://localhost:/%d", app.Conf.Api.Port)
	simpleweb.Run()
}
