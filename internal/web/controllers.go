package web

import (
	"net/http"
	"text/template"
	"traefik-multi-hosts/cmd/mhos"
	"traefik-multi-hosts/internal/config"
	"traefik-multi-hosts/internal/docker"
	"traefik-multi-hosts/internal/redis"
)

func getAllHostsWithServices(w http.ResponseWriter, redisClient *redis.ClientImpl) {
	hosts, err := redisClient.GetAllHostsWithServices()
	if err != nil {
		responseJSONWithStatus(w, http.StatusInternalServerError, UntypedJSON{
			"error": err.Error(),
		})
		return
	}
	responseJSON(w, hosts)
}

func health(w http.ResponseWriter) {
	responseJSON(w, UntypedJSON{
		"status": "ok",
	})
}

func serveIndexPage(w http.ResponseWriter, tmpl *template.Template, redisClient *redis.ClientImpl) {
	hosts, err := redisClient.GetAllHostsWithServices()
	if err != nil {
		responseJSONWithStatus(w, http.StatusInternalServerError, UntypedJSON{
			"error": err.Error(),
		})
		return
	}
	totalServices := 0
	for _, services := range hosts {
		totalServices += len(services)
	}
	err = tmpl.ExecuteTemplate(w, "index.html", UntypedJSON{
		"Hosts":         hosts,
		"CurrentHost":   config.HostIP(),
		"TotalServices": totalServices,
	})
	if err != nil {
		responseJSONWithStatus(w, http.StatusInternalServerError, UntypedJSON{
			"error": err.Error(),
		})
		return
	}
}

func freshScan(w http.ResponseWriter, dockerClient *docker.ClientImpl, redisClient *redis.ClientImpl) {
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
