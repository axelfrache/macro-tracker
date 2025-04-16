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
	FdcID       int        `json:"fdcId"`
	Description string     `json:"description"`
	DataType    string     `json:"dataType"`
	Nutrients   []Nutrient `json:"foodNutrients"`
}

type Nutrient struct {
	ID       int     `json:"nutrientId"`
	Name     string  `json:"nutrientName"`
	Amount   float64 `json:"value"`
	UnitName string  `json:"unitName"`
}

// GetNutrientValue récupère la valeur d'un nutriment par son ID
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
		
		ProteinID2  = 203 // Protein (ancien ID)
		CarbID2     = 205 // Carbohydrates (ancien ID)
		FatID2      = 204 // Total lipids (fat) (ancien ID)
		CalorieID2  = 208 // Energy (kcal) (ancien ID)
		FiberID2    = 291 // Fiber, total dietary (ancien ID)
	)

	proteins = f.GetNutrientValue(ProteinID, ProteinID2)
	carbs = f.GetNutrientValue(CarbID, CarbID2)
	fats = f.GetNutrientValue(FatID, FatID2)
	calories = f.GetNutrientValue(CalorieID, CalorieID2)
	fiber = f.GetNutrientValue(FiberID, FiberID2)

	if proteins == 0 && carbs == 0 && fats == 0 && calories == 0 {
		fmt.Printf("Aucune valeur nutritionnelle trouvée pour: %s\n", f.Description)
		if len(f.Nutrients) > 0 {
			fmt.Printf("Premier nutriment: ID=%d, Name=%s, Valeur=%f\n", 
				f.Nutrients[0].ID, f.Nutrients[0].Name, f.Nutrients[0].Amount)
		}
	}
	
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la lecture du corps de la réponse: %v", err)
	}
	
	fmt.Printf("Réponse brute de l'API (début): %.300s...\n", string(body))
	
	var result SearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("erreur lors de la désérialisation: %v", err)
	}
	
	fmt.Printf("Nombre d'aliments trouvés: %d\n", len(result.Foods))
	if len(result.Foods) > 0 {
		food := result.Foods[0]
		fmt.Printf("Premier aliment: ID=%d, Description=%s, Nutriments=%d\n", 
			food.FdcID, food.Description, len(food.Nutrients))
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la lecture du corps de la réponse: %v", err)
	}
	
	fmt.Printf("Réponse brute de GetFood (début): %.300s...\n", string(body))
	
	var food Food
	if err := json.Unmarshal(body, &food); err != nil {
		return nil, fmt.Errorf("erreur lors de la désérialisation: %v", err)
	}

	fmt.Printf("Aliment récupéré: ID=%d, Description=%s, Nutriments=%d\n",
		food.FdcID, food.Description, len(food.Nutrients))

	return &food, nil
}
