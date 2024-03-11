package web

import (
	"embed"
	"html/template"
	"traefik-multi-hosts/internal/log"

	"github.com/gin-gonic/gin"
)

//go:embed templates/*
var f embed.FS

func Serve() {
	log.Info().Msg("Starting web server")
	r := gin.Default()
	r.SetHTMLTemplate(template.Must(template.New("").ParseFS(f, "templates/*.html")))

	r.GET("/api/health", health)
	r.GET("/api/hosts", getAllHostsWithServices)
	r.GET("/", serveIndexPage)

	r.Run()
}
