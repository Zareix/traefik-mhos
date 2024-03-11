package main

import (
	"traefik-multi-hosts/cmd/mhos"
	"traefik-multi-hosts/internal/web"
)

func main() {
	mhos.Run()
	web.Serve()
}
