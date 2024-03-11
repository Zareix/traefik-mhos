package web

import (
	"embed"
	"html/template"
	"net/http"
	"traefik-multi-hosts/internal/config"
	"traefik-multi-hosts/internal/log"
	"traefik-multi-hosts/internal/redis"

	"github.com/gin-gonic/gin"
)

//go:embed templates/*
var f embed.FS

func Serve() {
	log.Info().Msg("Starting web server")
	r := gin.Default()
	templ := template.Must(template.New("").ParseFS(f, "templates/*.html"))
	r.SetHTMLTemplate(templ)

	r.GET("/api/health", health)
	r.GET("/api/hosts", getAllServices)
	r.GET("/", func(c *gin.Context) {
		hosts, err := redis.GetAllHostsWithServices()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Hosts":       hosts,
			"CurrentHost": config.AppConfig.HostIP,
		})
	})

	r.Run()
}

func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"health": "ok",
	})
}

func getAllServices(c *gin.Context) {
	hosts, err := redis.GetAllHostsWithServices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, hosts)
}
