package main

import (
	"fmt"
	"traefik-multi-hosts/cmd"
	"traefik-multi-hosts/internal/config"
)

func main() {
	fmt.Println("Starting traefik-multi-hosts")

	config.Init()
	cmd.Run()
}
