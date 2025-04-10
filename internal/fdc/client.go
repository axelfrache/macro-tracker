package fdc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	apiKey string
	client *http.Client
}

type SearchResponse struct {
	Foods []Food `json:"foods"`
}

type Food struct {
	FdcID       int     `json:"fdcId"`
	Description string  `json:"description"`
	DataType    string  `json:"dataType"`
	Nutrients   []Nutrient `json:"foodNutrients"`
}

type Nutrient struct {
	ID       int     `json:"nutrientId"`
	Name     string  `json:"nutrientName"`
	Amount   float64 `json:"value"`
	UnitName string  `json:"unitName"`
}

func (f *Food) GetNutrientValue(nutrientIDs ...int) float64 {
	for _, id := range nutrientIDs {
		for _, n := range f.Nutrients {
			if n.ID == id {
				return n.Amount
			}
		}
	}
	return 0
}

func (f *Food) GetMacros() (proteins, carbs, fats, calories, fiber float64) {
	const (
		ProteinID  = 1003 // Protein
		CarbID     = 1005 // Carbohydrates
		FatID      = 1004 // Total lipids (fat)
		CalorieID  = 1008 // Energy (kcal)
		FiberID    = 1079 // Fiber, total dietary
	)

	proteins = f.GetNutrientValue(ProteinID)
	carbs = f.GetNutrientValue(CarbID)
	fats = f.GetNutrientValue(FatID)
	calories = f.GetNutrientValue(CalorieID)
	fiber = f.GetNutrientValue(FiberID)
	return
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

func (c *Client) SearchFoods(query string) (*SearchResponse, error) {
	baseURL := "https://api.nal.usda.gov/fdc/v1/foods/search"
	
	params := url.Values{}
	params.Add("api_key", c.apiKey)
	params.Add("query", query)
	params.Add("dataType", "Foundation,SR Legacy")
	
	resp, err := c.client.Get(baseURL + "?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) GetFood(fdcID int) (*Food, error) {
	baseURL := fmt.Sprintf("https://api.nal.usda.gov/fdc/v1/food/%d", fdcID)
	
	params := url.Values{}
	params.Add("api_key", c.apiKey)
	params.Add("format", "full")
	
	resp, err := c.client.Get(baseURL + "?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	var food Food
	if err := json.NewDecoder(resp.Body).Decode(&food); err != nil {
		return nil, err
	}

	return &food, nil
}
