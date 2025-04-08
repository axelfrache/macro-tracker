package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/frachea/macro-tracker/config"
	"github.com/frachea/macro-tracker/internal/database"
	"github.com/frachea/macro-tracker/internal/fdc"
)

var (
	currentUser *database.User
	db         *database.DB
)

func main() {
	// Charger la configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Erreur de configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialiser la base de données
	connStr := "postgres://macrotracker:macrotracker@localhost:5434/macrotracker?sslmode=disable"
	db, err = database.NewDB(connStr)
	if err != nil {
		fmt.Printf("Erreur de connexion à la base de données: %v\n", err)
		os.Exit(1)
	}

	// Initialiser le client FDC
	fdcClient := fdc.NewClient(cfg.FDCApiKey)

	fmt.Println("Bienvenue dans Macro-Tracker!")
	
	// Demander l'ID utilisateur ou créer un nouvel utilisateur
	currentUser = setupUser()

	fmt.Println("\nCommandes disponibles:")
	fmt.Println("- search <nom de l'aliment>: rechercher un aliment")
	fmt.Println("- add <fdcId> <quantité> <type de repas>: ajouter un aliment consommé")
	fmt.Println("- report: voir le bilan nutritionnel du jour")
	fmt.Println("- plan: gérer les journées types")
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

		case "exit":
			fmt.Println("Au revoir!")
			return

		default:
			fmt.Println("Commande inconnue. Commandes disponibles: search, add, report, plan, exit")
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
		"dejeuner":      true,
		"diner":         true,
		"collation":     true,
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

	meal := &database.Meal{
		UserID:    currentUser.ID,
		MealType:  mealType,
		MealDate:  time.Now(),
		FoodID:    fdcID,
		FoodName:  food.Description,
		Amount:    amount,
		Proteins:  proteins * amount / 100,
		Carbs:     carbs * amount / 100,
		Fats:      fats * amount / 100,
		Calories:  calories * amount / 100,
		Fiber:     fiber * amount / 100,
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
		// Créer une journée type
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
		// Voir les journées types
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

			// Afficher les repas de la journée type
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
		// Ajouter un repas à une journée type
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

		// Vérifier que la journée type existe et appartient à l'utilisateur
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

		// Choisir le type de repas
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

		// Rechercher un aliment
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

		// Demander la quantité
		fmt.Print("Quantité (en grammes) : ")
		amountStr, _ := reader.ReadString('\n')
		amountStr = strings.TrimSpace(amountStr)
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			fmt.Println("Quantité invalide.")
			return
		}

		// Calculer les macronutriments pour la quantité donnée
		proteins, carbs, fats, calories, fiber := selectedFood.GetMacros()
		multiplier := amount / 100.0

		// Créer l'item de la journée type
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
