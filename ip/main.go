package ip

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type HttpBinResponse struct {
	Origin string `json:"origin"`
}

// ExternalIpV4 asks https://httpbin.org/ip for the client's ipv4 address.
// It returns the ipv4 address as a string.
func ExternalIpV4() (string, error) {
	req, err := http.NewRequest("GET", "https://httpbin.org/ip", nil)
	if err != nil {
		return "", fmt.Errorf("http.NewRequest: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("client.Do: %w", err)
	}
	defer resp.Body.Close()
	var httpBinResponse HttpBinResponse
	err = json.NewDecoder(resp.Body).Decode(&httpBinResponse)
	if err != nil {
		return "", fmt.Errorf("json.NewDecoder.Decode: %w", err)
	}
	return httpBinResponse.Origin, nil
}
