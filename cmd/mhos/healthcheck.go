package mhos

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
	"traefik-multi-hosts/internal/config"

	"github.com/rs/zerolog/log"
)

func Healthcheck() {
	port := config.Port()

	url := fmt.Sprintf("http://localhost:%s/api/health", port)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		log.Error().Err(err).Msg("Health check failed")
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error().Int("status_code", resp.StatusCode).Msg("Health check failed")
		os.Exit(1)
	}

	var healthResponse map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&healthResponse); err != nil {
		log.Error().Err(err).Msg("Health check failed: invalid response format")
		os.Exit(1)
	}

	status, exists := healthResponse["status"]
	if !exists || status != "ok" {
		log.Error().Msg("Health check failed: status is not ok")
		os.Exit(1)
	}

	log.Info().Msg("Health check passed: application is healthy")
	os.Exit(0)
}
