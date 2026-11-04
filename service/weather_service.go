package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

type WeatherResult struct {
	Location    string  `json:"location"`
	Temperature float64 `json:"temperature"`
	FeelsLike   float64 `json:"feels_like"`
	Humidity    int     `json:"humidity"`
	Description string  `json:"description"`
	WindSpeed   float64 `json:"wind_speed"`
}

type WeatherService struct {
	client *http.Client
}

func NewWeatherService() *WeatherService {
	return &WeatherService{client: &http.Client{
		Timeout: 10 * time.Second,
	}}
}

// Mengambil cuaca saat ini berdasarkan koordinat
func (s *WeatherService) GetCurrentWeather(ctx context.Context, lat, lon float64) (*WeatherResult, error) {
	apiKey := os.Getenv("OPENWEATHER_API_KEY")
	if apiKey == "" {
		return nil, errors.New("OPENWEATHER_API_KEY belum diatur")
	}

	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&appid=%s&units=metric&lang=id", lat, lon, apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	response, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenWeather mengembalikan status %d", response.StatusCode)
	}

	var raw struct {
		Name string `json:"name"`
		Main struct {
			Temp      float64 `json:"temp"`
			FeelsLike float64 `json:"feels_like"`
			Humidity  int     `json:"humidity"`
		} `json:"main"`
		Weather []struct {
			Description string `json:"description"`
		} `json:"weather"`
		Wind struct {
			Speed float64 `json:"speed"`
		} `json:"wind"`
	}

	if err := json.NewDecoder(response.Body).Decode(&raw); err != nil {
		return nil, err
	}

	description := "Tidak diketahui"
	if len(raw.Weather) > 0 {
		description = raw.Weather[0].Description
	}
	return &WeatherResult{
		Location:    raw.Name,
		Temperature: raw.Main.Temp,
		FeelsLike:   raw.Main.FeelsLike,
		Humidity:    raw.Main.Humidity,
		Description: description,
		WindSpeed:   raw.Wind.Speed,
	}, nil
}
