package web

import (
	"net/http"
	"traefik-multi-hosts/cmd/mhos"
	"traefik-multi-hosts/internal/config"
	"traefik-multi-hosts/internal/docker"
	"traefik-multi-hosts/internal/redis"

	"github.com/gin-gonic/gin"
)

func getAllHostsWithServices(c *gin.Context, redisClient redis.RedisClient) {
	hosts, err := redisClient.GetAllHostsWithServices()
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

func serveIndexPage(c *gin.Context, redisClient redis.RedisClient) {
	hosts, err := redisClient.GetAllHostsWithServices()
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

func freshScan(c *gin.Context, dockerClient docker.DockerClient, redisClient redis.RedisClient) {
	err := mhos.FreshScan(dockerClient, redisClient)
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
