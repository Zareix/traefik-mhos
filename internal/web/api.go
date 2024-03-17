package web

import (
	"embed"
	"fmt"
	"html/template"
	"traefik-multi-hosts/internal/log"

	"github.com/gin-gonic/gin"
)

//go:embed templates/*
var f embed.FS

func Serve() {
	log.Info().Msg("Starting web server")
	router := gin.New()
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("\033[34m%s\033[0m %s | %s \"%s\" | %d | %s %s\n",
			fmt.Sprintf("%-6s|", "INFO"),
			param.ClientIP,
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
			param.ErrorMessage,
		)
	}))
	router.SetHTMLTemplate(template.Must(template.New("").ParseFS(f, "templates/*.html")))
	router.Use(gin.Recovery())

	router.GET("/api/health", health)
	router.GET("/api/hosts", getAllHostsWithServices)
	router.POST("/api/scan", freshScan)
	router.GET("/", serveIndexPage)

	router.Run()
}
