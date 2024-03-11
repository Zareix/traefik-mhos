package main

import (
	"traefik-multi-hosts/cmd"
	"traefik-multi-hosts/internal/web"
)

func main() {
	cmd.Run()
	web.Serve()
}
