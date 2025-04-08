package fdc

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchFoods(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/fdc/v1/foods/search" {
			t.Errorf("Expected path '/fdc/v1/foods/search', got %s", r.URL.Path)
		}

		apiKey := r.URL.Query().Get("api_key")
		if apiKey != "test-key" {
			t.Errorf("Expected API key 'test-key', got %s", apiKey)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"foods": [
				{
					"fdcId": 123456,
					"description": "Test Food",
					"proteins": 10.5,
					"carbs": 20.3,
					"fats": 5.2,
					"calories": 170.5,
					"fiber": 3.1
				}
			]
		}`))
	}))
	defer server.Close()

	client := &Client{
		apiKey: "test-key",
		client: server.Client(),
	}

	resp, err := client.SearchFoods("test")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(resp.Foods) != 1 {
		t.Fatalf("Expected 1 food item, got %d", len(resp.Foods))
	}

	if resp.Foods[0].FdcID != 123456 {
		t.Errorf("Expected FdcID 123456, got %d", resp.Foods[0].FdcID)
	}

	if resp.Foods[0].Description != "Test Food" {
		t.Errorf("Expected description 'Test Food', got %s", resp.Foods[0].Description)
	}
}

func TestGetFood(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/fdc/v1/food/123456" {
			t.Errorf("Expected path '/fdc/v1/food/123456', got %s", r.URL.Path)
		}

		apiKey := r.URL.Query().Get("api_key")
		if apiKey != "test-key" {
			t.Errorf("Expected API key 'test-key', got %s", apiKey)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"fdcId": 123456,
			"description": "Test Food",
			"proteins": 10.5,
			"carbs": 20.3,
			"fats": 5.2,
			"calories": 170.5,
			"fiber": 3.1
		}`))
	}))
	defer server.Close()

	client := &Client{
		apiKey: "test-key",
		client: server.Client(),
	}

	food, err := client.GetFood(123456)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if food.FdcID != 123456 {
		t.Errorf("Expected FdcID 123456, got %d", food.FdcID)
	}

	if food.Description != "Test Food" {
		t.Errorf("Expected description 'Test Food', got %s", food.Description)
	}
}
