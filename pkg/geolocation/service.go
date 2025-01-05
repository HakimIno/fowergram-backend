package geolocation

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Service interface {
	GetLocation(ip string) (string, error)
}

type geoService struct {
	apiKey string
}

func NewGeoService(apiKey string) Service {
	return &geoService{apiKey: apiKey}
}

func (s *geoService) GetLocation(ip string) (string, error) {
	url := fmt.Sprintf("https://api.ipstack.com/%s?access_key=%s", ip, s.apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		City    string `json:"city"`
		Country string `json:"country_name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s, %s", result.City, result.Country), nil
}
