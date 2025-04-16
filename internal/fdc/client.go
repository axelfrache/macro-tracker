package fdc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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
	// Structure alternative pour l'API FDC v1
	Nutrient struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"nutrient"`
	Value  float64 `json:"amount"`
	Type   string  `json:"type"`
}

// GetNutrientValue récupère la valeur d'un nutriment par son ID
func (f *Food) GetNutrientValue(nutrientIDs ...int) float64 {
	for _, id := range nutrientIDs {
		for _, n := range f.Nutrients {
			// Format 1: Structure directe avec ID et Amount
			if n.ID == id && n.Amount > 0 {
				return n.Amount
			}
			
			// Format 2: Structure avec nutrient.id et value/amount
			if n.Nutrient.ID == id {
				if n.Value > 0 {
					return n.Value
				}
			}
			
			// Recherche dans des structures alternatives possibles
			// Ces adaptations permettent de gérer différentes versions de l'API
			if n.Type == "FoodNutrient" && n.Nutrient.ID == id {
				return n.Value
			}
		}
	}
	
	// Si rien n'est trouvé, parcourir à nouveau et chercher par nom
	for _, n := range f.Nutrients {
		for _, id := range nutrientIDs {
			// Correspondances de noms pour les nutriments courants
			if (id == 1003 || id == 203) && (containsIgnoreCase(n.Name, "protein") || 
			    containsIgnoreCase(n.Nutrient.Name, "protein")) {
				if n.Amount > 0 {
					return n.Amount
				}
				if n.Value > 0 {
					return n.Value
				}
			}
			
			if (id == 1005 || id == 205) && (containsIgnoreCase(n.Name, "carbohydrate") || 
			    containsIgnoreCase(n.Nutrient.Name, "carbohydrate")) {
				if n.Amount > 0 {
					return n.Amount
				}
				if n.Value > 0 {
					return n.Value
				}
			}
			
			if (id == 1004 || id == 204) && (containsIgnoreCase(n.Name, "fat") || 
			    containsIgnoreCase(n.Nutrient.Name, "fat")) {
				if n.Amount > 0 {
					return n.Amount
				}
				if n.Value > 0 {
					return n.Value
				}
			}
			
			if (id == 1008 || id == 208) && (containsIgnoreCase(n.Name, "energy") || 
			    containsIgnoreCase(n.Nutrient.Name, "energy")) {
				if n.Amount > 0 {
					return n.Amount
				}
				if n.Value > 0 {
					return n.Value
				}
			}
			
			if (id == 1079 || id == 291) && (containsIgnoreCase(n.Name, "fiber") || 
			    containsIgnoreCase(n.Nutrient.Name, "fiber")) {
				if n.Amount > 0 {
					return n.Amount
				}
				if n.Value > 0 {
					return n.Value
				}
			}
		}
	}
	
	return 0
}

// Fonction utilitaire pour vérifier si une chaîne contient une sous-chaîne (insensible à la casse)
func containsIgnoreCase(s, substr string) bool {
	if s == "" || substr == "" {
		return false
	}
	
	// Conversion simple en minuscules pour la comparaison
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
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
			// Ajouter des logs plus détaillés sur la structure des nutriments
			fmt.Printf("Premier nutriment: ID=%d, Name=%s, Valeur=%f\n", 
				f.Nutrients[0].ID, f.Nutrients[0].Name, f.Nutrients[0].Amount)
			fmt.Printf("Structure du premier nutriment: %+v\n", f.Nutrients[0])
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
	
	var result SearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("erreur lors de la désérialisation: %v", err)
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
	
	var food Food
	if err := json.Unmarshal(body, &food); err != nil {
		return nil, fmt.Errorf("erreur lors de la désérialisation: %v", err)
	}

	return &food, nil
}
