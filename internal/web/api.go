package web

import (
	"embed"
	"fmt"
	"html/template"
	"traefik-multi-hosts/internal/docker"
	"traefik-multi-hosts/internal/log"
	"traefik-multi-hosts/internal/redis"

	"github.com/gin-gonic/gin"
)

//go:embed templates/*
var f embed.FS

func Serve(dockerClient docker.DockerClient, redisClient redis.RedisClient) {
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
	router.GET("/api/hosts", func(c *gin.Context) {
		getAllHostsWithServices(c, redisClient)
	})
	router.POST("/api/scan", func(c *gin.Context) {
		freshScan(c, dockerClient, redisClient)
	})
	router.GET("/", func(c *gin.Context) {
		serveIndexPage(c, redisClient)
	})

	router.Run()
}
