package main

import (
	"traefik-multi-hosts/cmd"
	"traefik-multi-hosts/internal/config"
)

func main() {
	config.Init()
	cmd.Run()
}
