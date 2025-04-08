package core

import (
	"math"
	"testing"
)

func TestCalculerIMC(t *testing.T) {
	tests := []struct {
		name     string
		user     User
		expected float64
	}{
		{
			name: "IMC normal",
			user: User{
				Poids:  70.0,
				Taille: 175.0,
			},
			expected: 22.86,
		},
		{
			name: "Taille nulle",
			user: User{
				Poids:  70.0,
				Taille: 0.0,
			},
			expected: 0.0,
		},
		{
			name: "IMC surpoids",
			user: User{
				Poids:  90.0,
				Taille: 180.0,
			},
			expected: 27.78,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.CalculerIMC()
			if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("CalculerIMC() = %v, attendu %v (marge d'erreur de 0.01)", result, tt.expected)
			}
		})
	}
}
