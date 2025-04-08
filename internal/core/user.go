package core

import (
	"time"
)

type User struct {
	ID        int64     `json:"id"`
	Nom       string    `json:"nom"`
	Prenom    string    `json:"prenom"`
	Age       int       `json:"age"`
	Poids     float64   `json:"poids"`
	Taille    float64   `json:"taille"`
	Objectif  string    `json:"objectif"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) CalculerIMC() float64 {
	if u.Taille == 0 {
		return 0
	}
	tailleEnMetres := u.Taille / 100
	return u.Poids / (tailleEnMetres * tailleEnMetres)
}
