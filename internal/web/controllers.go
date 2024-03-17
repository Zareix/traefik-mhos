package web

import (
	"net/http"
	"traefik-multi-hosts/cmd/mhos"
	"traefik-multi-hosts/internal/config"
	"traefik-multi-hosts/internal/redis"

	"github.com/gin-gonic/gin"
)

func getAllHostsWithServices(c *gin.Context) {
	hosts, err := redis.GetAllHostsWithServices(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, hosts)
}

func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"health": "ok",
	})
}

func serveIndexPage(c *gin.Context) {
	hosts, err := redis.GetAllHostsWithServices(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.HTML(http.StatusOK, "index.html", gin.H{
		"Hosts":       hosts,
		"CurrentHost": config.HostIP(),
	})
}

func freshScan(c *gin.Context) {
	err := mhos.FreshScan(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
