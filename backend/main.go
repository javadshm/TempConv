package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// ConversionRequest holds the temperature value to convert.
type ConversionRequest struct {
	Value float64 `json:"value"`
}

// ConversionResponse holds the converted temperature.
type ConversionResponse struct {
	Value float64 `json:"value"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/celsius-to-fahrenheit", celsiusToFahrenheitHandler)
	mux.HandleFunc("/api/fahrenheit-to-celsius", fahrenheitToCelsiusHandler)

	port := "8080"
	log.Printf("TempConv backend listening on :%s", port)
	if err := http.ListenAndServe(":"+port, cors(mux)); err != nil {
		log.Fatal(err)
	}
}

// cors wraps a handler to add CORS headers (needed when frontend and backend differ by origin).
func cors(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func celsiusToFahrenheitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	value, err := parseValue(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	result := value*9/5 + 32
	writeJSON(w, ConversionResponse{Value: result})
}

func fahrenheitToCelsiusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	value, err := parseValue(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	result := (value - 32) * 5 / 9
	writeJSON(w, ConversionResponse{Value: result})
}

func parseValue(r *http.Request) (float64, error) {
	if r.Header.Get("Content-Type") == "application/json" {
		var req ConversionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return 0, err
		}
		return req.Value, nil
	}
	valueStr := r.URL.Query().Get("value")
	if valueStr == "" {
		return 0, strconv.ErrSyntax
	}
	return strconv.ParseFloat(valueStr, 64)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
