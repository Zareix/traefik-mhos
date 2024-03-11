package main

import (
	"traefik-multi-hosts/cmd"
	"traefik-multi-hosts/internal/config"
	"traefik-multi-hosts/internal/web"
)

func main() {
	config.Init()
	cmd.Run()
	web.Serve()
}
