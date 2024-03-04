package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func doHTTP(url, key string, target interface{}) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to issue HTTP request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Incident.io Prometheus Exporter - https://github.com/dirsigler/incidentio-exporter")

	client := http.DefaultClient

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to run HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received HTTP statuscode not 200: %v", resp.StatusCode)
	}

	if err = json.NewDecoder(resp.Body).Decode(&target); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}
