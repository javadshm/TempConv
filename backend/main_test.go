package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	healthHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("health: got status %d", w.Code)
	}
	var m map[string]string
	if err := json.NewDecoder(w.Body).Decode(&m); err != nil {
		t.Fatal(err)
	}
	if m["status"] != "ok" {
		t.Errorf("health: got %q", m["status"])
	}
}

func TestCelsiusToFahrenheit(t *testing.T) {
	body, _ := json.Marshal(ConversionRequest{Value: 0})
	req := httptest.NewRequest(http.MethodPost, "/api/celsius-to-fahrenheit", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	celsiusToFahrenheitHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("c2f: got status %d", w.Code)
	}
	var res ConversionResponse
	if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}
	if res.Value != 32 {
		t.Errorf("c2f(0): got %f want 32", res.Value)
	}
}

func TestFahrenheitToCelsius(t *testing.T) {
	body, _ := json.Marshal(ConversionRequest{Value: 32})
	req := httptest.NewRequest(http.MethodPost, "/api/fahrenheit-to-celsius", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	fahrenheitToCelsiusHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("f2c: got status %d", w.Code)
	}
	var res ConversionResponse
	if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}
	if res.Value != 0 {
		t.Errorf("f2c(32): got %f want 0", res.Value)
	}
}
