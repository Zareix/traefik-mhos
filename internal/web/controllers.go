package web

import (
	"net/http"
	"text/template"
	"traefik-multi-hosts/cmd/mhos"
	"traefik-multi-hosts/internal/config"
	"traefik-multi-hosts/internal/docker"
	"traefik-multi-hosts/internal/redis"
)

func getAllHostsWithServices(w http.ResponseWriter, r *http.Request, redisClient redis.RedisClient) {
	hosts, err := redisClient.GetAllHostsWithServices()
	if err != nil {
		responseJSONWithStatus(w, http.StatusInternalServerError, UntypedJSON{
			"error": err.Error(),
		})
		return
	}
	responseJSON(w, hosts)
}

func health(w http.ResponseWriter, r *http.Request) {
	responseJSON(w, UntypedJSON{
		"status": "ok",
	})
}

func serveIndexPage(w http.ResponseWriter, r *http.Request, tmpl *template.Template, redisClient redis.RedisClient) {
	hosts, err := redisClient.GetAllHostsWithServices()
	if err != nil {
		responseJSONWithStatus(w, http.StatusInternalServerError, UntypedJSON{
			"error": err.Error(),
		})
		return
	}
	tmpl.ExecuteTemplate(w, "index.html", UntypedJSON{
		"Hosts":       hosts,
		"CurrentHost": config.HostIP(),
	})
}

func freshScan(w http.ResponseWriter, r *http.Request, dockerClient docker.DockerClient, redisClient redis.RedisClient) {
	err := mhos.FreshScan(dockerClient, redisClient)
	if err != nil {
		responseJSONWithStatus(w, http.StatusInternalServerError, UntypedJSON{
			"error": err.Error(),
		})
		return
	}
	responseJSON(w, UntypedJSON{
		"status": "ok",
	})
}
