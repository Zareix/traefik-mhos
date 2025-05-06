.PHONY: live/server live/tailwind build/tailwind build/server dev build lint

BINARY_NAME = traefik-mhos
BUILD_DIR = bin
OS = linux darwin windows
ARCH = amd64 arm64

live/server:
	REDIS_DB=0 \
	REDIS_PASSWORD=password \
	LOG_LEVEL=debug \
	PORT=8888 \
	LISTEN_EVENTS=true \
	gow -e=go,mod,html run .

live/tailwind:
	bun run --bun tailwindcss -i internal/web/static/css/input.css -o internal/web/static/css/style.css --watch

build/tailwind:
	bun run --bun tailwindcss -i internal/web/static/css/input.css -o internal/web/static/css/style.css --minify

build/server:
	$(foreach os,$(OS),\
		$(foreach arch,$(ARCH),\
			GOOS=$(os) GOARCH=$(arch) go build -o $(BUILD_DIR)/$(BINARY_NAME)_$(os)-$(arch)$(if $(filter windows,$(os)),.exe,);\
		)\
	)

install:
	bun install && go mod download

dev:
	$(MAKE) install && $(MAKE) -j2 live/tailwind live/server

build:
	$(MAKE) install build/tailwind build/server

lint:
	staticcheck ./...