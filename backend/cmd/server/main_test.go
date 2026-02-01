package main

import (
	"context"
	"testing"

	pb "github.com/javadshm/TempConv/backend/gen/api"
)

func TestConvert_CelsiusToFahrenheit(t *testing.T) {
	s := &server{}
	req := &pb.ConvertRequest{
		Value:    0,
		FromUnit: pb.Unit_CELSIUS,
		ToUnit:   pb.Unit_FAHRENHEIT,
	}
	resp, err := s.Convert(context.Background(), req)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	if resp.Value != 32 {
		t.Errorf("Convert(0°C to F): got %f, want 32", resp.Value)
	}
}

func TestConvert_FahrenheitToCelsius(t *testing.T) {
	s := &server{}
	req := &pb.ConvertRequest{
		Value:    32,
		FromUnit: pb.Unit_FAHRENHEIT,
		ToUnit:   pb.Unit_CELSIUS,
	}
	resp, err := s.Convert(context.Background(), req)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	if resp.Value != 0 {
		t.Errorf("Convert(32°F to C): got %f, want 0", resp.Value)
	}
}

func TestConvert_SameUnit(t *testing.T) {
	s := &server{}
	req := &pb.ConvertRequest{
		Value:    25,
		FromUnit: pb.Unit_CELSIUS,
		ToUnit:   pb.Unit_CELSIUS,
	}
	resp, err := s.Convert(context.Background(), req)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	if resp.Value != 25 {
		t.Errorf("Convert(25°C to C): got %f, want 25", resp.Value)
	}
}

func TestConvert_AdditionalCases(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		fromUnit pb.Unit
		toUnit   pb.Unit
		want     float64
	}{
		{"100°C to F", 100, pb.Unit_CELSIUS, pb.Unit_FAHRENHEIT, 212},
		{"212°F to C", 212, pb.Unit_FAHRENHEIT, pb.Unit_CELSIUS, 100},
		{"-40°C to F", -40, pb.Unit_CELSIUS, pb.Unit_FAHRENHEIT, -40},
		{"-40°F to C", -40, pb.Unit_FAHRENHEIT, pb.Unit_CELSIUS, -40},
	}

	s := &server{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &pb.ConvertRequest{
				Value:    tt.value,
				FromUnit: tt.fromUnit,
				ToUnit:   tt.toUnit,
			}
			resp, err := s.Convert(context.Background(), req)
			if err != nil {
				t.Fatalf("Convert failed: %v", err)
			}
			if resp.Value != tt.want {
				t.Errorf("got %f, want %f", resp.Value, tt.want)
			}
		})
	}
}
