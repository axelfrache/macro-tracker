package main

import (
	"bufio"
	"encoding/json"
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/frachea/macro-tracker/config"
	"github.com/frachea/macro-tracker/internal/database"
	"github.com/frachea/macro-tracker/internal/fdc"
)

// Structure pour stocker les objectifs de macronutriments
type MacroTargets struct {
	Calories float64 `json:"calories"`
	Proteins float64 `json:"proteins"`
	Carbs    float64 `json:"carbs"`
	Fats     float64 `json:"fats"`
	Fiber    float64 `json:"fiber"`
}

var (
	currentUser *database.User
	db          *database.DB
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Erreur de configuration: %v\n", err)
		os.Exit(1)
	}

	connStr := "postgres://postgres:postgres@db:5432/macro_tracker?sslmode=disable"
	db, err = database.NewDB(connStr)
	if err != nil {
		fmt.Printf("Erreur de connexion à la base de données: %v\n", err)
		os.Exit(1)
	}

	// Appliquer les migrations
	err = db.ApplyMigrations("./internal/database/migrations")
	if err != nil {
		fmt.Printf("Erreur lors de l'application des migrations: %v\n", err)
		os.Exit(1)
	}

	fdcClient := fdc.NewClient(cfg.FDCApiKey)

	fmt.Println("Bienvenue dans Macro-Tracker!")

	currentUser = setupUser()

	fmt.Println("\nCommandes disponibles:")
	fmt.Println("- search <nom de l'aliment>: rechercher un aliment")
	fmt.Println("- add <fdcId> <quantité> <type de repas>: ajouter un aliment consommé")
	fmt.Println("- report: voir le bilan nutritionnel du jour")
	fmt.Println("- plan: gérer les journées types")
	fmt.Println("- health: afficher les informations de santé (IMC, masse grasse)")
	fmt.Println("- goals: définir ou consulter vos objectifs nutritionnels")
	fmt.Println("- history [jours]: afficher l'historique (défaut: 7 jours)")
	fmt.Println("- profile: modifier vos informations personnelles")
	fmt.Println("- export: exporter vos données en CSV")
	fmt.Println("- exit: quitter l'application")

	scanner := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\n> ")
		input, _ := scanner.ReadString('\n')
		input = strings.TrimSpace(input)
		args := strings.Fields(input)

		if len(args) == 0 {
			continue
		}

		command := args[0]
		switch command {
		case "search":
			if len(args) < 2 {
				fmt.Println("Usage: search <nom de l'aliment>")
				continue
			}
			query := strings.Join(args[1:], " ")
			handleSearch(fdcClient, query)

		case "add":
			if len(args) < 4 {
				fmt.Println("Usage: add <fdcId> <quantité en grammes> <type de repas>")
				fmt.Println("Types de repas disponibles: petit-dejeuner, dejeuner, diner, collation")
				continue
			}
			handleAdd(fdcClient, args[1:])

		case "report":
			handleReport()

		case "plan":
			handlePlanCommand(scanner, db, fdcClient, currentUser)

		case "health":
			handleHealth()

		case "goals":
			if len(args) > 1 {
				handleGoalsUpdate(scanner)
			} else {
				handleGoalsView()
			}

		case "history":
			days := 7 // Par défaut, afficher 7 jours
			if len(args) > 1 {
				days, _ = strconv.Atoi(args[1])
				if days <= 0 {
					days = 7
				}
			}
			handleHistory(days)

		case "profile":
			handleProfile(scanner)

		case "export":
			handleExport()

		case "exit":
			fmt.Println("Au revoir!")
			return

		default:
			fmt.Println("Commande inconnue. Commandes disponibles: search, add, report, plan, health, goals, history, profile, export, exit")
		}
	}
}

func setupUser() *database.User {
	fmt.Print("Entrez votre ID utilisateur (ou 0 pour créer un nouveau compte): ")
	scanner := bufio.NewReader(os.Stdin)
	input, _ := scanner.ReadString('\n')
	id, _ := strconv.Atoi(strings.TrimSpace(input))

	if id > 0 {
		user, err := db.GetUser(id)
		if err == nil {
			fmt.Printf("Bienvenue, %s!\n", user.Name)
			return user
		}
		fmt.Println("Utilisateur non trouvé. Création d'un nouveau compte...")
	}

	user := &database.User{}
	fmt.Print("Nom: ")
	user.Name, _ = scanner.ReadString('\n')
	user.Name = strings.TrimSpace(user.Name)

	fmt.Print("Âge: ")
	ageStr, _ := scanner.ReadString('\n')
	user.Age, _ = strconv.Atoi(strings.TrimSpace(ageStr))

	fmt.Print("Poids (kg): ")
	weightStr, _ := scanner.ReadString('\n')
	user.Weight, _ = strconv.ParseFloat(strings.TrimSpace(weightStr), 64)

	fmt.Print("Taille (cm): ")
	heightStr, _ := scanner.ReadString('\n')
	user.Height, _ = strconv.ParseFloat(strings.TrimSpace(heightStr), 64)

	fmt.Print("Genre (homme/femme): ")
	user.Gender, _ = scanner.ReadString('\n')
	user.Gender = strings.TrimSpace(user.Gender)

	// Initialiser les macros cibles avec un JSON vide
	user.TargetMacros = []byte("{}")

	err := db.AddUser(user)
	if err != nil {
		fmt.Printf("Erreur lors de la création de l'utilisateur: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Compte créé avec succès! Votre ID est: %d\n", user.ID)
	return user
}

func handleSearch(client *fdc.Client, query string) {
	resp, err := client.SearchFoods(query)
	if err != nil {
		fmt.Printf("Erreur lors de la recherche: %v\n", err)
		return
	}

	if len(resp.Foods) == 0 {
		fmt.Println("Aucun aliment trouvé.")
		return
	}

	fmt.Println("\nRésultats de la recherche:")
	for _, food := range resp.Foods {
		fmt.Printf("- ID: %d, Nom: %s\n", food.FdcID, food.Description)
	}
}

func handleAdd(client *fdc.Client, args []string) {
	fdcID, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("ID d'aliment invalide")
		return
	}

	amount, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		fmt.Println("Quantité invalide")
		return
	}

	mealType := args[2]
	validMealTypes := map[string]bool{
		"petit-dejeuner": true,
		"dejeuner":       true,
		"diner":          true,
		"collation":      true,
	}

	if !validMealTypes[mealType] {
		fmt.Println("Type de repas invalide. Utilisez: petit-dejeuner, dejeuner, diner, ou collation")
		return
	}

	food, err := client.GetFood(fdcID)
	if err != nil {
		fmt.Printf("Erreur lors de la récupération des détails de l'aliment: %v\n", err)
		return
	}

	proteins, carbs, fats, calories, fiber := food.GetMacros()
	
	// S'assurer que les valeurs sont positives
	if proteins < 0 {
		proteins = 0
	}
	if carbs < 0 {
		carbs = 0
	}
	if fats < 0 {
		fats = 0
	}
	if calories < 0 {
		calories = 0
	}
	if fiber < 0 {
		fiber = 0
	}
	
	// Calculer les valeurs en fonction de la quantité
	mealProteins := proteins * amount / 100
	mealCarbs := carbs * amount / 100
	mealFats := fats * amount / 100
	mealCalories := calories * amount / 100
	mealFiber := fiber * amount / 100

	meal := &database.Meal{
		UserID:   currentUser.ID,
		MealType: mealType,
		MealDate: time.Now(),
		FoodID:   fdcID,
		FoodName: food.Description,
		Amount:   amount,
		Proteins: mealProteins,
		Carbs:    mealCarbs,
		Fats:     mealFats,
		Calories: mealCalories,
		Fiber:    mealFiber,
	}

	// Vérifier que les valeurs sont correctes avant de les enregistrer
	if mealProteins <= 0 && mealCarbs <= 0 && mealFats <= 0 && mealCalories <= 0 {
		fmt.Println("Attention: Aucune valeur nutritionnelle trouvée pour cet aliment. Vérifier l'API ou l'ID de l'aliment.")
	}

	err = db.AddMeal(meal)
	if err != nil {
		fmt.Printf("Erreur lors de l'ajout du repas: %v\n", err)
		return
	}

	fmt.Printf("Aliment ajouté avec succès au repas: %s\n", mealType)
}

func handleReport() {
	today := time.Now()

	meals, err := db.GetDailyMeals(currentUser.ID, today)
	if err != nil {
		fmt.Printf("Erreur lors de la récupération des repas: %v\n", err)
		return
	}

	proteins, carbs, fats, calories, fiber, err := db.GetDailyTotals(currentUser.ID, today)
	if err != nil {
		fmt.Printf("Erreur lors du calcul des totaux: %v\n", err)
		return
	}

	fmt.Printf("\nBilan nutritionnel du %s:\n", today.Format("02/01/2006"))
	fmt.Println("\nRepas de la journée:")
	for _, meal := range meals {
		fmt.Printf("- %s: %s (%.0fg)\n", meal.MealType, meal.FoodName, meal.Amount)
	}

	fmt.Println("\nTotaux journaliers:")
	fmt.Printf("- Calories: %.0f kcal\n", calories)
	fmt.Printf("- Protéines: %.1fg\n", proteins)
	fmt.Printf("- Glucides: %.1fg\n", carbs)
	fmt.Printf("- Lipides: %.1fg\n", fats)
	fmt.Printf("- Fibres: %.1fg\n", fiber)

	// Afficher la comparaison avec les objectifs si définis
	var targets MacroTargets
	if len(currentUser.TargetMacros) > 0 {
		err := json.Unmarshal(currentUser.TargetMacros, &targets)
		if err == nil && targets.Calories > 0 {
			fmt.Println("\nComparaison avec vos objectifs:")
			fmt.Printf("- Calories: %.0f/%.0f kcal (%.0f%%)\n", 
				calories, targets.Calories, calories/targets.Calories*100)
			fmt.Printf("- Protéines: %.1f/%.1fg (%.0f%%)\n", 
				proteins, targets.Proteins, proteins/targets.Proteins*100)
			fmt.Printf("- Glucides: %.1f/%.1fg (%.0f%%)\n", 
				carbs, targets.Carbs, carbs/targets.Carbs*100)
			fmt.Printf("- Lipides: %.1f/%.1fg (%.0f%%)\n", 
				fats, targets.Fats, fats/targets.Fats*100)
			fmt.Printf("- Fibres: %.1f/%.1fg (%.0f%%)\n", 
				fiber, targets.Fiber, fiber/targets.Fiber*100)
		}
	}
}

func handlePlanCommand(reader *bufio.Reader, db *database.DB, fdcClient *fdc.Client, user *database.User) {
	fmt.Print("\nGestion des journées types\n")
	fmt.Print("1. Créer une journée type\n")
	fmt.Print("2. Voir les journées types\n")
	fmt.Print("3. Ajouter un repas à une journée type\n")
	fmt.Print("Choisissez une option (1-3) : ")

	option, _ := reader.ReadString('\n')
	option = strings.TrimSpace(option)

	switch option {
	case "1":
		fmt.Print("Nom de la journée type : ")
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)

		fmt.Print("Description : ")
		description, _ := reader.ReadString('\n')
		description = strings.TrimSpace(description)

		plan := &database.MealPlan{
			UserID:      user.ID,
			Name:        name,
			Description: description,
		}

		err := db.CreateMealPlan(plan)
		if err != nil {
			fmt.Printf("Erreur lors de la création de la journée type : %v\n", err)
			return
		}

		fmt.Printf("Journée type '%s' créée avec succès !\n", name)

	case "2":
		plans, err := db.GetMealPlans(user.ID)
		if err != nil {
			fmt.Printf("Erreur lors de la récupération des journées types : %v\n", err)
			return
		}

		if len(plans) == 0 {
			fmt.Println("Aucune journée type trouvée.")
			return
		}

		fmt.Println("\nJournées types :")
		for _, plan := range plans {
			fmt.Printf("\n%d. %s\n", plan.ID, plan.Name)
			fmt.Printf("   Description : %s\n", plan.Description)

			items, err := db.GetMealPlanItems(plan.ID)
			if err != nil {
				fmt.Printf("Erreur lors de la récupération des repas : %v\n", err)
				continue
			}

			if len(items) > 0 {
				fmt.Println("   Repas :")
				for _, item := range items {
					fmt.Printf("   - %s : %s (%.0fg)\n", item.MealType, item.FoodName, item.Amount)
				}
			}
		}

	case "3":
		plans, err := db.GetMealPlans(user.ID)
		if err != nil {
			fmt.Printf("Erreur lors de la récupération des journées types : %v\n", err)
			return
		}

		if len(plans) == 0 {
			fmt.Println("Aucune journée type trouvée. Créez-en une d'abord.")
			return
		}

		fmt.Println("\nChoisissez une journée type :")
		for _, plan := range plans {
			fmt.Printf("%d. %s\n", plan.ID, plan.Name)
		}

		fmt.Print("Numéro de la journée type : ")
		planIDStr, _ := reader.ReadString('\n')
		planIDStr = strings.TrimSpace(planIDStr)
		planID, err := strconv.Atoi(planIDStr)
		if err != nil {
			fmt.Println("Numéro de journée type invalide.")
			return
		}

		var selectedPlan *database.MealPlan
		for _, plan := range plans {
			if plan.ID == planID {
				selectedPlan = &plan
				break
			}
		}

		if selectedPlan == nil {
			fmt.Println("Journée type non trouvée.")
			return
		}

		fmt.Println("\nTypes de repas disponibles :")
		fmt.Println("1. Petit-déjeuner (breakfast)")
		fmt.Println("2. Collation 1 (snack1)")
		fmt.Println("3. Déjeuner (lunch)")
		fmt.Println("4. Collation 2 (snack2)")
		fmt.Println("5. Dîner (dinner)")

		fmt.Print("Choisissez un type de repas (1-5) : ")
		mealTypeStr, _ := reader.ReadString('\n')
		mealTypeStr = strings.TrimSpace(mealTypeStr)

		var mealType database.MealType
		switch mealTypeStr {
		case "1":
			mealType = database.Breakfast
		case "2":
			mealType = database.Snack1
		case "3":
			mealType = database.Lunch
		case "4":
			mealType = database.Snack2
		case "5":
			mealType = database.Dinner
		default:
			fmt.Println("Type de repas invalide.")
			return
		}

		fmt.Print("\nRechercher un aliment : ")
		query, _ := reader.ReadString('\n')
		query = strings.TrimSpace(query)

		result, err := fdcClient.SearchFoods(query)
		if err != nil {
			fmt.Printf("Erreur lors de la recherche : %v\n", err)
			return
		}

		if len(result.Foods) == 0 {
			fmt.Println("Aucun aliment trouvé.")
			return
		}

		fmt.Println("\nRésultats de la recherche :")
		for i, food := range result.Foods {
			fmt.Printf("%d. %s\n", i+1, food.Description)
		}

		fmt.Print("\nChoisissez un aliment (numéro) : ")
		foodIndexStr, _ := reader.ReadString('\n')
		foodIndexStr = strings.TrimSpace(foodIndexStr)
		foodIndex, err := strconv.Atoi(foodIndexStr)
		if err != nil || foodIndex < 1 || foodIndex > len(result.Foods) {
			fmt.Println("Numéro d'aliment invalide.")
			return
		}

		selectedFood := result.Foods[foodIndex-1]

		fmt.Print("Quantité (en grammes) : ")
		amountStr, _ := reader.ReadString('\n')
		amountStr = strings.TrimSpace(amountStr)
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			fmt.Println("Quantité invalide.")
			return
		}

		proteins, carbs, fats, calories, fiber := selectedFood.GetMacros()
		multiplier := amount / 100.0

		item := &database.MealPlanItem{
			MealPlanID: planID,
			MealType:   mealType,
			FoodID:     selectedFood.FdcID,
			FoodName:   selectedFood.Description,
			Amount:     amount,
			Proteins:   proteins * multiplier,
			Carbs:      carbs * multiplier,
			Fats:       fats * multiplier,
			Calories:   calories * multiplier,
			Fiber:      fiber * multiplier,
		}

		err = db.AddMealPlanItem(item)
		if err != nil {
			fmt.Printf("Erreur lors de l'ajout du repas : %v\n", err)
			return
		}

		fmt.Printf("\nRepas ajouté avec succès à la journée type '%s' !\n", selectedPlan.Name)

	default:
		fmt.Println("Option invalide.")
	}
}

// Calcule l'IMC (Indice de Masse Corporelle)
func calculateBMI(weight, height float64) float64 {
	// Hauteur en mètres
	heightInMeters := height / 100
	return weight / (heightInMeters * heightInMeters)
}

// Calcule le taux de masse grasse (formule simplifiée)
// Utilise la formule de l'US Navy
func calculateBodyFat(user *database.User) float64 {
	// Implémentation basique, à améliorer avec des formules plus précises
	bmi := calculateBMI(user.Weight, user.Height)
	
	// Facteur d'âge (simplifié)
	ageFactor := float64(user.Age) * 0.12
	
	if user.Gender == "homme" {
		return (1.20 * bmi) + (0.23 * ageFactor) - 16.2
	} else {
		return (1.20 * bmi) + (0.23 * ageFactor) - 5.4
	}
}

// Affiche les informations de santé
func handleHealth() {
	bmi := calculateBMI(currentUser.Weight, currentUser.Height)
	bodyFat := calculateBodyFat(currentUser)
	
	fmt.Println("\nInformations de santé:")
	fmt.Printf("- Poids: %.1f kg\n", currentUser.Weight)
	fmt.Printf("- Taille: %.1f cm\n", currentUser.Height)
	fmt.Printf("- IMC: %.1f\n", bmi)
	
	// Interprétation de l'IMC
	var interpretation string
	switch {
	case bmi < 18.5:
		interpretation = "Insuffisance pondérale"
	case bmi < 25:
		interpretation = "Corpulence normale"
	case bmi < 30:
		interpretation = "Surpoids"
	default:
		interpretation = "Obésité"
	}
	
	fmt.Printf("  Interprétation: %s\n", interpretation)
	fmt.Printf("- Taux de masse grasse estimé: %.1f%%\n", bodyFat)
}

// Affiche les objectifs nutritionnels
func handleGoalsView() {
	var targets MacroTargets
	
	// Charger les objectifs depuis la base de données
	if len(currentUser.TargetMacros) > 0 {
		err := json.Unmarshal(currentUser.TargetMacros, &targets)
		if err != nil {
			fmt.Printf("Erreur lors de la lecture des objectifs: %v\n", err)
			return
		}
	}
	
	fmt.Println("\nVos objectifs nutritionnels:")
	if targets.Calories == 0 {
		fmt.Println("Aucun objectif défini. Utilisez 'goals set' pour définir vos objectifs.")
		return
	}
	
	fmt.Printf("- Calories: %.0f kcal\n", targets.Calories)
	fmt.Printf("- Protéines: %.1fg (%.0f%%)\n", targets.Proteins, targets.Proteins*4/targets.Calories*100)
	fmt.Printf("- Glucides: %.1fg (%.0f%%)\n", targets.Carbs, targets.Carbs*4/targets.Calories*100)
	fmt.Printf("- Lipides: %.1fg (%.0f%%)\n", targets.Fats, targets.Fats*9/targets.Calories*100)
	fmt.Printf("- Fibres: %.1fg\n", targets.Fiber)
}

// Met à jour les objectifs nutritionnels
func handleGoalsUpdate(scanner *bufio.Reader) {
	var targets MacroTargets
	
	fmt.Println("\nDéfinition des objectifs nutritionnels:")
	
	fmt.Print("Calories quotidiennes: ")
	caloriesStr, _ := scanner.ReadString('\n')
	targets.Calories, _ = strconv.ParseFloat(strings.TrimSpace(caloriesStr), 64)
	
	fmt.Print("Pourcentage de protéines (ex: 30 pour 30%): ")
	proteinPctStr, _ := scanner.ReadString('\n')
	proteinPct, _ := strconv.ParseFloat(strings.TrimSpace(proteinPctStr), 64)
	
	fmt.Print("Pourcentage de glucides (ex: 40 pour 40%): ")
	carbsPctStr, _ := scanner.ReadString('\n')
	carbsPct, _ := strconv.ParseFloat(strings.TrimSpace(carbsPctStr), 64)
	
	fmt.Print("Pourcentage de lipides (ex: 30 pour 30%): ")
	fatsPctStr, _ := scanner.ReadString('\n')
	fatsPct, _ := strconv.ParseFloat(strings.TrimSpace(fatsPctStr), 64)
	
	// Vérifier que les pourcentages totalisent 100%
	total := proteinPct + carbsPct + fatsPct
	if math.Abs(total-100) > 1 {
		fmt.Printf("Attention: Les pourcentages totalisent %.0f%% au lieu de 100%%\n", total)
		fmt.Print("Voulez-vous ajuster automatiquement? (o/n): ")
		adjust, _ := scanner.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(adjust)) == "o" {
			factor := 100 / total
			proteinPct *= factor
			carbsPct *= factor
			fatsPct *= factor
			fmt.Printf("Pourcentages ajustés: Protéines %.0f%%, Glucides %.0f%%, Lipides %.0f%%\n", 
				proteinPct, carbsPct, fatsPct)
		}
	}
	
	// Calculer les quantités en grammes
	targets.Proteins = (targets.Calories * (proteinPct / 100)) / 4  // 4 kcal/g pour les protéines
	targets.Carbs = (targets.Calories * (carbsPct / 100)) / 4       // 4 kcal/g pour les glucides
	targets.Fats = (targets.Calories * (fatsPct / 100)) / 9         // 9 kcal/g pour les lipides
	
	fmt.Print("Objectif de fibres (g): ")
	fiberStr, _ := scanner.ReadString('\n')
	targets.Fiber, _ = strconv.ParseFloat(strings.TrimSpace(fiberStr), 64)
	
	// Sauvegarder les objectifs dans la base de données
	targetsJSON, err := json.Marshal(targets)
	if err != nil {
		fmt.Printf("Erreur lors de la conversion des objectifs: %v\n", err)
		return
	}
	
	currentUser.TargetMacros = targetsJSON
	err = db.UpdateUser(currentUser)
	if err != nil {
		fmt.Printf("Erreur lors de la sauvegarde des objectifs: %v\n", err)
		return
	}
	
	fmt.Println("Objectifs nutritionnels mis à jour avec succès!")
}

// Affiche l'historique des repas et nutriments sur plusieurs jours
func handleHistory(days int) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days+1)
	
	fmt.Printf("\nHistorique nutritionnel du %s au %s:\n\n", 
		startDate.Format("02/01/2006"), 
		endDate.Format("02/01/2006"))
	
	// Pour chaque jour
	for d := 0; d < days; d++ {
		date := endDate.AddDate(0, 0, -d)
		proteins, carbs, fats, calories, fiber, err := db.GetDailyTotals(currentUser.ID, date)
		
		if err != nil {
			fmt.Printf("Erreur lors de la récupération des données pour le %s: %v\n", 
				date.Format("02/01/2006"), err)
			continue
		}
		
		// N'afficher que les jours avec des données
		if calories > 0 {
			fmt.Printf("- %s: %.0f kcal, P:%.1fg, C:%.1fg, L:%.1fg, F:%.1fg\n", 
				date.Format("02/01/2006"), calories, proteins, carbs, fats, fiber)
		}
	}
}

// Permet de modifier les informations du profil
func handleProfile(scanner *bufio.Reader) {
	fmt.Println("\nModification du profil:")
	fmt.Printf("Nom actuel: %s\n", currentUser.Name)
	fmt.Print("Nouveau nom (laisser vide pour conserver): ")
	
	name, _ := scanner.ReadString('\n')
	name = strings.TrimSpace(name)
	if name != "" {
		currentUser.Name = name
	}
	
	fmt.Printf("Âge actuel: %d ans\n", currentUser.Age)
	fmt.Print("Nouvel âge: ")
	ageStr, _ := scanner.ReadString('\n')
	ageStr = strings.TrimSpace(ageStr)
	if ageStr != "" {
		currentUser.Age, _ = strconv.Atoi(ageStr)
	}
	
	fmt.Printf("Poids actuel: %.1f kg\n", currentUser.Weight)
	fmt.Print("Nouveau poids (kg): ")
	weightStr, _ := scanner.ReadString('\n')
	weightStr = strings.TrimSpace(weightStr)
	if weightStr != "" {
		currentUser.Weight, _ = strconv.ParseFloat(weightStr, 64)
	}
	
	fmt.Printf("Taille actuelle: %.1f cm\n", currentUser.Height)
	fmt.Print("Nouvelle taille (cm): ")
	heightStr, _ := scanner.ReadString('\n')
	heightStr = strings.TrimSpace(heightStr)
	if heightStr != "" {
		currentUser.Height, _ = strconv.ParseFloat(heightStr, 64)
	}
	
	fmt.Printf("Genre actuel: %s\n", currentUser.Gender)
	fmt.Print("Nouveau genre (homme/femme): ")
	gender, _ := scanner.ReadString('\n')
	gender = strings.TrimSpace(gender)
	if gender != "" {
		currentUser.Gender = gender
	}
	
	err := db.UpdateUser(currentUser)
	if err != nil {
		fmt.Printf("Erreur lors de la mise à jour du profil: %v\n", err)
		return
	}
	
	fmt.Println("Profil mis à jour avec succès!")
}

// Exporte les données de l'utilisateur en CSV
func handleExport() {
	// Créer le dossier d'export s'il n'existe pas
	exportDir := "exports"
	if _, err := os.Stat(exportDir); os.IsNotExist(err) {
		os.Mkdir(exportDir, 0755)
	}
	
	// Créer le fichier CSV
	filename := fmt.Sprintf("%s/export_%d_%s.csv", exportDir, currentUser.ID, time.Now().Format("20060102"))
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Erreur lors de la création du fichier: %v\n", err)
		return
	}
	defer file.Close()
	
	// Préparer le writer CSV
	writer := csv.NewWriter(file)
	defer writer.Flush()
	
	// Écrire l'en-tête
	header := []string{"Date", "Type de repas", "Aliment", "Quantité (g)", "Calories", "Protéines", "Glucides", "Lipides", "Fibres"}
	writer.Write(header)
	
	// Récupérer toutes les données de repas
	endDate := time.Now()
	startDate := endDate.AddDate(0, -1, 0) // 1 mois en arrière
	
	meals, err := db.GetMealsBetweenDates(currentUser.ID, startDate, endDate)
	if err != nil {
		fmt.Printf("Erreur lors de la récupération des repas: %v\n", err)
		return
	}
	
	// Écrire les données
	for _, meal := range meals {
		row := []string{
			meal.MealDate.Format("2006-01-02"),
			meal.MealType,
			meal.FoodName,
			fmt.Sprintf("%.1f", meal.Amount),
			fmt.Sprintf("%.1f", meal.Calories),
			fmt.Sprintf("%.1f", meal.Proteins),
			fmt.Sprintf("%.1f", meal.Carbs),
			fmt.Sprintf("%.1f", meal.Fats),
			fmt.Sprintf("%.1f", meal.Fiber),
		}
		writer.Write(row)
	}
	
	fmt.Printf("Données exportées avec succès dans le fichier: %s\n", filename)
}
