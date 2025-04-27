package web

import (
	"embed"
	"net/http"
	"text/template"
	"traefik-multi-hosts/internal/docker"
	"traefik-multi-hosts/internal/redis"

	"github.com/rs/zerolog/log"
)

//go:embed templates/*
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

func Serve(dockerClient docker.DockerClient, redisClient redis.RedisClient) {
	log.Info().Msg("Starting web server")
	tmpl, _ := template.New("").ParseFS(templateFS, "templates/*.html")
	router := http.NewServeMux()

	router.HandleFunc("GET /api/health", health)
	router.HandleFunc("GET /api/hosts", func(w http.ResponseWriter, r *http.Request) {
		getAllHostsWithServices(w, r, redisClient)
	})
	router.HandleFunc("POST /api/scan", func(w http.ResponseWriter, r *http.Request) {
		freshScan(w, r, dockerClient, redisClient)
	})
	router.Handle("GET /static/", http.FileServer(http.FS(staticFS)))
	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		serveIndexPage(w, r, tmpl, redisClient)
	})

	http.ListenAndServe(":8888", LoggingMiddleware(router))
}
