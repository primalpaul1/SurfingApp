package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type ForecastResponse struct {
	SpotID         string `json:"spotId"`
	Location       string `json:"location"`
	WaveHeight     string `json:"waveHeight"`
	WindSpeed      string `json:"windSpeed"`
	WindDirection  string `json:"windDirection"`
	Tide           string `json:"tide"`
	Timestamp      int64  `json:"timestamp"`
}

// Map of Surfline spot IDs to location names
var spotLocations = map[string]string{
	"5842041f4e65fad6a7708814": "Malibu, CA",
	"5842041f4e65fad6a770883d": "Huntington Beach, CA",
	"5842041f4e65fad6a7709115": "Tamarindo, CR",
	"5842041f4e65fad6a7709117": "Jaco, CR",
	"5842041f4e65fad6a7709116": "Dominical, CR",
}

// Simple in-memory cache
var forecastCache = make(map[string]CacheItem)

type CacheItem struct {
	Response  ForecastResponse
	ExpiresAt int64
}

const CACHE_DURATION = 30 * 60 // 30 minutes in seconds

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/forecast", handleForecast)
	mux.HandleFunc("/health", handleHealth)
	
	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func handleForecast(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	spotID := r.URL.Query().Get("spotId")
	if spotID == "" {
		http.Error(w, "Missing spotId parameter", http.StatusBadRequest)
		return
	}

	// Check if we should bypass cache
	bypassCache := false
	bypassCacheParam := r.URL.Query().Get("bypassCache")
	if bypassCacheParam != "" {
		var err error
		bypassCache, err = strconv.ParseBool(bypassCacheParam)
		if err != nil {
			bypassCache = false
		}
	}
	
	// Check cache first
	now := time.Now().Unix()
	if !bypassCache {
		if cacheItem, ok := forecastCache[spotID]; ok && cacheItem.ExpiresAt > now {
			log.Printf("Cache hit for spot ID: %s", spotID)
			json.NewEncoder(w).Encode(cacheItem.Response)
			return
		}
	}
	
	log.Printf("Fetching fresh data for spot ID: %s", spotID)
	
	// For now, we'll return mock data since we're not actually connecting to Surfline yet
	// In a real implementation, you would use the surflinef library here
	response := getMockForecastResponse(spotID)
	
	// Cache the response
	forecastCache[spotID] = CacheItem{
		Response:  response,
		ExpiresAt: now + CACHE_DURATION,
	}
	
	// Return the response
	json.NewEncoder(w).Encode(response)
}

func getMockForecastResponse(spotID string) ForecastResponse {
	// Get the location name
	location, ok := spotLocations[spotID]
	if !ok {
		location = "Unknown Location"
	}
	
	// Create mock data based on the spot ID
	var waveHeight, windSpeed, windDirection, tide string
	
	switch spotID {
	case "5842041f4e65fad6a7708814": // Malibu
		waveHeight = "3.8 ft at 12 seconds 215 degrees"
		windSpeed = "5 mph"
		windDirection = "Offshore"
		tide = "Rising, 2.5ft at 10:30am"
	case "5842041f4e65fad6a770883d": // Huntington
		waveHeight = "2.5 ft at 10 seconds 220 degrees"
		windSpeed = "8 mph"
		windDirection = "Cross-shore"
		tide = "Falling, 3.2ft at 9:15am"
	case "5842041f4e65fad6a7709115": // Tamarindo
		waveHeight = "4.5 ft at 14 seconds 210 degrees"
		windSpeed = "3 mph"
		windDirection = "Offshore"
		tide = "High, 4.1ft at 11:45am"
	case "5842041f4e65fad6a7709117": // Jaco
		waveHeight = "3.7 ft at 12 seconds 205 degrees"
		windSpeed = "6 mph"
		windDirection = "Offshore"
		tide = "Low, 1.2ft at 8:30am"
	case "5842041f4e65fad6a7709116": // Dominical
		waveHeight = "5.2 ft at 16 seconds 207 degrees"
		windSpeed = "4 mph"
		windDirection = "Offshore"
		tide = "Mid, 2.8ft at 9:45am"
	default:
		waveHeight = "Unknown"
		windSpeed = "Unknown"
		windDirection = "Unknown"
		tide = "Unknown"
	}
	
	return ForecastResponse{
		SpotID:         spotID,
		Location:       location,
		WaveHeight:     waveHeight,
		WindSpeed:      windSpeed,
		WindDirection:  windDirection,
		Tide:           tide,
		Timestamp:      time.Now().Unix(),
	}
}
