package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const apiKey = "DIq75ui1kfOvr1C07gRenmNoPlCemSTaewVWDxvf" // ← Remplace par ta clé API

type Food struct {
	Description string `json:"description"`
	FdcID       int    `json:"fdcId"`
	DataType    string `json:"dataType"`
}

type SearchResponse struct {
	Foods []Food `json:"foods"`
}

func searchFood(query string) {
	baseURL := "https://api.nal.usda.gov/fdc/v1/foods/search"
	params := url.Values{}
	params.Add("api_key", apiKey)
	params.Add("query", query)
	params.Add("pageSize", "5") // Limite à 5 résultats pour l’exemple

	fullURL := baseURL + "?" + params.Encode()

	resp, err := http.Get(fullURL)
	if err != nil {
		fmt.Println("Erreur requête :", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result SearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Erreur JSON :", err)
		return
	}

	fmt.Println("Résultats pour :", query)
	for _, food := range result.Foods {
		fmt.Printf("- %s (ID: %d, Type: %s)\n", food.Description, food.FdcID, food.DataType)
	}
}

func main() {
	searchFood("banana")
}
