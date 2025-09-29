package web

import (
	"embed"
	"fmt"
	"net/http"
	"text/template"
	"traefik-multi-hosts/internal/config"
	"traefik-multi-hosts/internal/docker"
	"traefik-multi-hosts/internal/redis"

	"github.com/rs/zerolog/log"
)

//go:embed templates/*
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

func Serve(dockerClient *docker.ClientImpl, redisClient *redis.ClientImpl) {
	log.Info().Msg("Starting web server")
	tmpl, _ := template.New("").ParseFS(templateFS, "templates/*.html")
	router := http.NewServeMux()

	router.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		health(w)
	})
	router.HandleFunc("GET /api/hosts", func(w http.ResponseWriter, r *http.Request) {
		getAllHostsWithServices(w, redisClient)
	})
	router.HandleFunc("POST /api/scan", func(w http.ResponseWriter, r *http.Request) {
		freshScan(w, dockerClient, redisClient)
	})
	router.Handle("GET /static/", http.FileServer(http.FS(staticFS)))
	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		serveIndexPage(w, tmpl, redisClient)
	})

	port := fmt.Sprintf(":%s", config.Port())
	log.Info().Str("port", port).Msg("Starting web server on port")
	err := http.ListenAndServe(port, LoggingMiddleware(router))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start web server")
		return
	}
}
