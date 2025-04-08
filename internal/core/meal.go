package core

import (
	"time"
)

type Meal struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Nom       string    `json:"nom"`
	Type      string    `json:"type"`
	Date      time.Time `json:"date"`
	Aliments  []Aliment `json:"aliments"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Aliment struct {
	ID        int64   `json:"id"`
	FdcID     string  `json:"fdc_id"`
	Nom       string  `json:"nom"`
	Quantite  float64 `json:"quantite"`
	Calories  float64 `json:"calories"`
	Proteines float64 `json:"proteines"`
	Glucides  float64 `json:"glucides"`
	Lipides   float64 `json:"lipides"`
	Fibres    float64 `json:"fibres"`
}

func (m *Meal) CalculerTotalNutriments() (calories, proteines, glucides, lipides float64) {
	for _, aliment := range m.Aliments {
		ratio := aliment.Quantite / 100
		calories += aliment.Calories * ratio
		proteines += aliment.Proteines * ratio
		glucides += aliment.Glucides * ratio
		lipides += aliment.Lipides * ratio
	}
	return
}
