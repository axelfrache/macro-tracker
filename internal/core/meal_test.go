package core

import (
	"testing"
)

func TestCalculerTotalNutriments(t *testing.T) {
	tests := []struct {
		name              string
		meal              Meal
		expectedCalories  float64
		expectedProteines float64
		expectedGlucides  float64
		expectedLipides   float64
	}{
		{
			name: "Repas simple",
			meal: Meal{
				Aliments: []Aliment{
					{
						Nom:       "Poulet",
						Quantite:  150.0, // 150g
						Calories:  165.0, // pour 100g
						Proteines: 31.0,  // pour 100g
						Glucides:  0.0,   // pour 100g
						Lipides:   3.6,   // pour 100g
					},
					{
						Nom:       "Riz",
						Quantite:  100.0, // 100g
						Calories:  130.0, // pour 100g
						Proteines: 2.7,   // pour 100g
						Glucides:  28.0,  // pour 100g
						Lipides:   0.3,   // pour 100g
					},
				},
			},
			expectedCalories:  377.5, // (165 * 1.5) + (130 * 1)
			expectedProteines: 49.2,  // (31 * 1.5) + (2.7 * 1)
			expectedGlucides:  28.0,  // (0 * 1.5) + (28 * 1)
			expectedLipides:   5.7,   // (3.6 * 1.5) + (0.3 * 1)
		},
		{
			name: "Repas vide",
			meal: Meal{
				Aliments: []Aliment{},
			},
			expectedCalories:  0.0,
			expectedProteines: 0.0,
			expectedGlucides:  0.0,
			expectedLipides:   0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calories, proteines, glucides, lipides := tt.meal.CalculerTotalNutriments()

			if calories != tt.expectedCalories {
				t.Errorf("Calories = %v, attendu %v", calories, tt.expectedCalories)
			}
			if proteines != tt.expectedProteines {
				t.Errorf("Protéines = %v, attendu %v", proteines, tt.expectedProteines)
			}
			if glucides != tt.expectedGlucides {
				t.Errorf("Glucides = %v, attendu %v", glucides, tt.expectedGlucides)
			}
			if lipides != tt.expectedLipides {
				t.Errorf("Lipides = %v, attendu %v", lipides, tt.expectedLipides)
			}
		})
	}
}
